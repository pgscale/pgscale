// Copyright 2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgproto3/v2"
	"github.com/pgscale/pgscale/config"
	"github.com/pgscale/pgscale/kontext"
)

// https://www.postgresql.org/docs/current/protocol-flow.html
// https://www.postgresql.org/docs/11/runtime-config-connection.html

var (
	ErrUnknownAuthenticationMethod = errors.New("unknown authentication method")
	ErrSessionNotFound             = errors.New("session not found")
)

const ProtocolVersion3 = 196608 // 3.0

type Session struct {
	ProtocolVersionNumber int
	ApplicationName       string
	User                  string
	Database              string
}

type Auth struct {
	config  *config.Config
	backend *pgproto3.Backend
	conn    net.Conn
}

func SessionFromKontext(k *kontext.Kontext) (*Session, error) {
	i := k.Get(kontext.SessionKey)
	if i == nil {
		return nil, ErrSessionNotFound
	}

	s, ok := i.(*Session)
	if !ok {
		return nil, fmt.Errorf("session: %w", kontext.ErrInvalidType)
	}

	return s, nil
}

func New(c *config.Config, conn net.Conn) *Auth {
	backend := pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn)
	return &Auth{
		config:  c,
		backend: backend,
		conn:    conn,
	}
}

func (a *Auth) authOK() error {
	buf := (&pgproto3.AuthenticationOk{}).Encode(nil)
	buf = (&pgproto3.ParameterStatus{Name: "server_version", Value: "13.4 (Debian 13.4-1.pgdg100+1)"}).Encode(buf)
	buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
	_, err := a.conn.Write(buf)
	return err
}

func (a *Auth) checkMD5Password(s *Session, msg *pgproto3.PasswordMessage, creds map[string]string) bool {
	// PostgreSQL MD5-hashed password format: "md5" + md5(password + username)
	return fmt.Sprintf("md5%s", creds["hash"]) == msg.Password
}

func (a *Auth) checkCleartextPassword(msg *pgproto3.PasswordMessage, creds map[string]string) bool {
	return creds["password"] == msg.Password
}

func (a *Auth) errorResponse(msg string) error {
	e := &pgproto3.ErrorResponse{
		Severity: "FATAL",
		Message:  msg,
	}
	_, err := a.conn.Write(e.Encode(nil))
	if err != nil {
		return fmt.Errorf("error sending error message: %w", err)
	}
	return fmt.Errorf(msg)
}

func (a *Auth) authUserWithPassword(s *Session, credentials map[string]string, msg *pgproto3.PasswordMessage) error {
	authType := credentials["auth_type"]
	switch authType {
	case config.MD5AuthType:
		if !a.checkMD5Password(s, msg, credentials) {
			return a.errorResponse(fmt.Sprintf("password authentication failed for user \"%s\"", s.User))
		}
	case config.PasswordAuthType:
		if !a.checkCleartextPassword(msg, credentials) {
			return a.errorResponse(fmt.Sprintf("password authentication failed for user \"%s\"", s.User))
		}
	default:
		err := a.errorResponse(fmt.Sprintf("%s is an unknown authentication method for user \"%s\"", authType, s.User))
		if err != nil {
			return fmt.Errorf("%w: %s", ErrUnknownAuthenticationMethod, err.Error())
		}

		return fmt.Errorf("%w: %s", ErrUnknownAuthenticationMethod, authType)
	}

	buf := (&pgproto3.AuthenticationOk{}).Encode(nil)
	// TODO: We need more parameters
	buf = (&pgproto3.ParameterStatus{Name: "server_version", Value: "13.4 (Debian 13.4-1.pgdg100+1)"}).Encode(buf)
	buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
	_, err := a.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("error sending ready for query: %w", err)
	}

	return nil
}

func (a *Auth) HandleAuth(s *Session, credentials map[string]string) (*Session, error) {
	frontendMsg, err := a.backend.Receive()
	if err != nil {
		return nil, fmt.Errorf("error receiving auth message: %w", err)
	}

	switch msg := frontendMsg.(type) {
	case *pgproto3.PasswordMessage:
		if err := a.authUserWithPassword(s, credentials, msg); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown startup message: %#v", frontendMsg)
	}

	return s, nil
}

func (a *Auth) doCleartextPasswordAuth() error {
	buf := (&pgproto3.AuthenticationCleartextPassword{}).Encode(nil)
	_, err := a.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("error sending AuthenticationCleartextPassword message: %w", err)
	}

	return nil
}

func (a *Auth) doMD5PasswordAuth() error {
	buf := (&pgproto3.AuthenticationMD5Password{}).Encode(nil)
	_, err := a.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("error sending AuthenticationMD5Password message: %w", err)
	}
	return nil
}

func (a *Auth) HandleStartup() (*Session, error) {
	startupMessage, err := a.backend.ReceiveStartupMessage()
	if err != nil {
		return nil, fmt.Errorf("error receiving startup message: %w", err)
	}

	s := &Session{}

	switch msg := startupMessage.(type) {
	case *pgproto3.StartupMessage:
		// We may need to implement different versions of the Postgres protocol. Keep it.
		s.ProtocolVersionNumber = ProtocolVersion3
		if user, ok := msg.Parameters["user"]; ok {
			s.User = user
		}
		if database, ok := msg.Parameters["database"]; ok {
			s.Database = database
		}
		if applicationName, ok := msg.Parameters["application_name"]; ok {
			s.ApplicationName = applicationName
		}

		// Startup is done.

		// Check authentication method and run
		credentials, ok := a.config.PgScale.Auth.Users[s.User]
		if !ok {
			return nil, a.errorResponse(fmt.Sprintf("no such user: \"%s\"", s.User))
		}

		authType := credentials["auth_type"]

		switch authType {
		case config.TrustAuthType:
			if err = a.authOK(); err != nil {
				return nil, err
			}
			return s, nil
		case config.PasswordAuthType:
			if err = a.doCleartextPasswordAuth(); err != nil {
				return nil, err
			}
			return a.HandleAuth(s, credentials)
		case config.MD5AuthType:
			if err = a.doMD5PasswordAuth(); err != nil {
				return nil, err
			}
			return a.HandleAuth(s, credentials)
		default:
			return nil, fmt.Errorf("unknown auth type: %s", authType)
		}
	case *pgproto3.SSLRequest:
		_, err = a.conn.Write([]byte("N"))
		if err != nil {
			return nil, fmt.Errorf("error sending deny SSL request: %w", err)
		}
		return a.HandleStartup()
	default:
		return nil, fmt.Errorf("unknown startup message: %#v", startupMessage)
	}
}
