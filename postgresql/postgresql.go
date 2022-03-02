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

package postgresql

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/buraksezer/olric/pkg/flog"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pgscale/pgscale/config"
	"github.com/pgscale/pgscale/dmaps"
	"github.com/pgscale/pgscale/kontext"
	"github.com/pgscale/pgscale/postgresql/auth"
	"github.com/pgscale/pgscale/postgresql/dbconn"
	"github.com/pgscale/pgscale/tcp"
	"github.com/pgscale/pgscale/utils"
)

type PostgreSQL struct {
	log     *flog.Logger
	config  *config.Config
	dbconns map[string]map[string]*dbconn.Conn
	server  *tcp.Server
	dmaps   *dmaps.DMaps
	ctx     context.Context
	cancel  context.CancelFunc
}

func New(k *kontext.Kontext) (*PostgreSQL, error) {
	c, err := config.FromKontext(k)
	if err != nil {
		return nil, err
	}

	lg, err := utils.FlogFromKontext(k)
	if err != nil {
		return nil, err
	}

	dms, err := utils.DMapsFromKontext(k)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &PostgreSQL{
		log:     lg,
		config:  c,
		dbconns: make(map[string]map[string]*dbconn.Conn),
		dmaps:   dms,
		ctx:     ctx,
		cancel:  cancel,
	}

	err = p.initializePools()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PostgreSQL) afterRelease(conn *pgx.Conn, db *config.Database) bool {
	_, releaseErr := conn.Query(p.ctx, db.ResetQuery)
	if releaseErr != nil {
		p.log.V(3).Printf("[ERROR] Failed to reset session: %v", releaseErr)
		return false
	}
	return true
}

func (p *PostgreSQL) initializePools() error {
	for dbname, database := range p.config.PgScale.PostgreSQL.Databases {
		cfg, err := pgxpool.ParseConfig(database.ConnString())
		if err != nil {
			return err
		}

		cfg.AfterRelease = func(conn *pgx.Conn) bool {
			return p.afterRelease(conn, p.config.PgScale.PostgreSQL.Databases[dbname])
		}

		_, ok := p.dbconns[cfg.ConnConfig.User]
		if !ok {
			p.dbconns[cfg.ConnConfig.User] = make(map[string]*dbconn.Conn)
		}
		db := p.dbconns[cfg.ConnConfig.User]
		db[cfg.ConnConfig.Database] = &dbconn.Conn{
			Database: database,
			Config:   cfg,
		}
	}
	return nil
}

func (p *PostgreSQL) ListenAndServe() error {
	k := kontext.New()
	k.Set(kontext.ConfigKey, p.config)
	k.Set(kontext.LoggerKey, p.log)
	k.Set(kontext.DMapsKey, p.dmaps)

	s, err := tcp.New(k, p.callback, p.proxyHandler)
	if err != nil {
		return err
	}
	p.server = s

	err = p.server.ListenAndServe()
	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			err = nil
		}
	}

	return err
}

func (p *PostgreSQL) callback() {
	p.log.V(1).Printf("[INFO] PgScale instance addr: %s",
		net.JoinHostPort(
			p.config.PgScale.BindAddr,
			p.config.PgScale.BindPort,
		),
	)
	p.log.V(1).Printf("[INFO] PostgreSQL proxy is ready to accept connections")
}

func (p *PostgreSQL) proxyHandler(conn net.Conn) error {
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			p.log.V(3).Printf("[ERROR] Failed to close client socket: %s", cerr)
		}
	}()

	a := auth.New(p.config, conn)
	session, err := a.HandleStartup()
	if err != nil {
		return err
	}

	var ok bool
	var dc *dbconn.Conn
	for _, db := range p.dbconns {
		dc, ok = db[session.Database]
		if ok {
			break
		}
	}
	if !ok {
		msg := fmt.Sprintf("failed to database: %s in config", session.Database)
		e := &pgproto3.ErrorResponse{
			Severity: "FATAL",
			Message:  msg,
		}
		_, err = conn.Write(e.Encode(nil))
		if err != nil {
			return fmt.Errorf("failed to return error response: %w", err)
		}
		return fmt.Errorf(msg)
	}

	err = dc.CreatePool(p.ctx)
	if err != nil {
		return err
	}

	k := kontext.New()
	k.Set(kontext.LoggerKey, p.log)
	k.Set(kontext.DMapsKey, p.dmaps)
	k.Set(kontext.ConfigKey, p.config)
	k.Set(kontext.DBConnKey, dc)
	k.Set(kontext.SessionKey, session)

	pr, err := NewProxy(k, conn)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- pr.Start()
	}()

	go func() {
		<-p.ctx.Done()

		closeErr := pr.Close()
		if closeErr != nil {
			p.log.V(3).Printf("[ERROR] Failed to close proxy gracefully")
		}
	}()

	return <-errCh
}

func (p *PostgreSQL) Shutdown() error {
	select {
	case <-p.ctx.Done():
		return nil
	default:
	}

	p.cancel()

	for _, db := range p.dbconns {
		for _, conn := range db {
			if conn.Pool != nil {
				conn.Pool.Close()
			}
		}
	}

	return p.server.Shutdown()
}
