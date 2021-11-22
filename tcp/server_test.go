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

package tcp

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/buraksezer/dante/config"
	"github.com/buraksezer/dante/kontext"
	"github.com/buraksezer/dante/testutils"
	"github.com/stretchr/testify/require"
)

func TestTCP_Server(t *testing.T) {
	port, err := testutils.GetFreePort()
	require.NoError(t, err)

	cfg := testutils.NewDanteConfig(t)
	c, err := config.New(cfg)
	require.NoError(t, err)

	c.Dante.BindAddr = "127.0.0.1"
	c.Dante.BindPort = strconv.Itoa(port)

	ctx, cancel := context.WithCancel(context.Background())
	started := func() {
		cancel()
	}

	var message = []byte("Hello, world!")

	echoHandler := func(conn net.Conn) error {
		data := make([]byte, len(message))
		_, err := conn.Read(data)
		if err != nil {
			return err
		}
		_, err = conn.Write(data)
		if err != nil {
			return err
		}
		return nil
	}

	k := kontext.New()
	k.Set(kontext.ConfigKey, c)
	k.Set(kontext.LoggerKey, testutils.NewFlogLogger())
	s, err := New(k, started, echoHandler)
	require.NoError(t, err)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ListenAndServe()
	}()

	t.Run("Started callback", func(t *testing.T) {
		select {
		case <-time.After(time.Second):
			require.Fail(t, "TCP server has not been started yet")
		case <-ctx.Done():
			return
		}
	})

	t.Run("Make connection and send a message", func(t *testing.T) {
		clientConn, err := net.Dial("tcp", s.addr)
		require.NoError(t, err)

		nr, err := clientConn.Write(message)
		require.NoError(t, err)
		require.Len(t, message, nr)

		data := make([]byte, len(message))
		_, err = clientConn.Read(data)
		require.NoError(t, err)
		require.Equal(t, message, data)
	})

	err = s.Shutdown()
	require.NoError(t, err)
}
