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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/pkg/flog"
	"github.com/cespare/xxhash/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pgscale/pgscale/bufpool"
	"github.com/pgscale/pgscale/config"
	"github.com/pgscale/pgscale/dmaps"
	"github.com/pgscale/pgscale/kontext"
	"github.com/pgscale/pgscale/postgresql/auth"
	"github.com/pgscale/pgscale/postgresql/dbconn"
	"github.com/pgscale/pgscale/postgresql/protocol"
	"github.com/pgscale/pgscale/utils"
	"golang.org/x/sync/errgroup"
)

// https://pgpool.net/mediawiki/index.php/pgpool-II_3.5_features

const (
	QueryIdentifier         = byte('Q')
	ParseIdentifier         = byte('P')
	ReadyForQueryIdentifier = byte('Z')
	SyncIdentifier          = byte('S')
	TerminateIdentifier     = byte('X')
	BindIdentifier          = byte('B')
)

var (
	ErrGetOrCreateDMap = errors.New("failed to get or create DMap")
	ErrClientIsGone    = errors.New("client is gone")
)

var pool = bufpool.New()

type Proxy struct {
	config     *config.Config
	session    *auth.Session
	hashPrefix []byte
	client     net.Conn
	dbconn     *dbconn.Conn
	log        *flog.Logger
	dmaps      *dmaps.DMaps
	kontext    *kontext.Kontext
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewProxy(k *kontext.Kontext, client net.Conn) (*Proxy, error) {
	lg, err := utils.FlogFromKontext(k)
	if err != nil {
		return nil, err
	}

	dms, err := utils.DMapsFromKontext(k)
	if err != nil {
		return nil, err
	}

	c, err := config.FromKontext(k)
	if err != nil {
		return nil, err
	}

	dc, err := dbconn.ConnFromKontext(k)
	if err != nil {
		return nil, err
	}

	session, err := auth.SessionFromKontext(k)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Proxy{
		config:  c,
		session: session,
		client:  client,
		dbconn:  dc,
		log:     lg,
		dmaps:   dms,
		kontext: kontext.New(),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (p *Proxy) Start() error {
	var errGr errgroup.Group
	ctx, cancel := context.WithCancel(context.Background())

	errGr.Go(func() error {
		defer cancel()

		clientErr := p.clientLoop()
		if errors.Is(clientErr, ErrClientIsGone) {
			p.log.V(3).Printf("[DEBUG] Client is gone")
			clientErr = nil
		}
		if clientErr != nil {
			p.log.V(3).Printf("[ERROR] Failed to process message from client to server: %v", clientErr)
		}
		return clientErr
	})

	select {
	case <-p.ctx.Done():
		// Proxy instance is gone
	case <-ctx.Done():
		err := p.Close()
		if err != nil {
			p.log.V(3).Printf("[ERROR] Failed to close proxy: %v", err)
		}
	}

	return errGr.Wait()
}

func (p *Proxy) hashQuery(query []byte) uint64 {
	h := xxhash.New()
	_, _ = h.Write(p.hashPrefix)
	_, _ = h.Write(query)
	return h.Sum64()
}

func (p *Proxy) loadFromCache(table *config.Table, data *protocol.DataPacket) (interface{}, error) {
	dm, err := p.dmaps.GetOrCreateDMap(table.DMapName)
	if err != nil {
		p.log.V(3).Printf("[ERROR] Failed to get distributed map object: %v", err)
		// Just log this
		return nil, fmt.Errorf("%w: %v", ErrGetOrCreateDMap, err)
	}

	hquery := p.hashQuery(data.Payload)
	value, err := dm.Get(strconv.FormatUint(hquery, 10))
	if errors.Is(err, olric.ErrKeyNotFound) {
		p.kontext.Set("start", true)
		p.kontext.Set("table", table)
		p.kontext.Set("query", string(data.Payload))
		p.kontext.Set("hquery", hquery)
		buf, ok := p.kontext.Get("cache").(*bytes.Buffer)
		if ok {
			buf.Reset()
		} else {
			p.kontext.Set("cache", bytes.NewBuffer(nil))
		}

		return nil, olric.ErrKeyNotFound
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (p *Proxy) serveFromCache(value interface{}) (bool, error) {
	buf := pool.Get()
	defer pool.Put(buf)

	buf.Write(value.([]byte))
	nr, err := io.Copy(p.client, buf)
	if err != nil {
		return false, err
	}
	p.log.V(3).Printf("[DEBUG] Number of bytes served from cache: %d", nr)
	return true, nil
}

func (p *Proxy) cacheDataPacket(data *protocol.DataPacket) {
	cache := p.kontext.Get("cache").(*bytes.Buffer)
	cache.Write(data.Header)
	cache.Write(data.Payload)

	if data.Identifier != ReadyForQueryIdentifier {
		return
	}

	defer cache.Reset()

	p.kontext.Set("start", false)
	hquery := p.kontext.Get("hquery").(uint64)
	table := p.kontext.Get("table").(*config.Table)
	dm, err := p.dmaps.GetOrCreateDMap(table.DMapName)
	if err != nil {
		p.log.V(3).Printf("[ERROR] Failed to get distributed map object: %v", err)
		return
	}

	err = dm.Put(strconv.FormatUint(hquery, 10), cache.Bytes())
	if err != nil {
		p.log.V(3).Printf("[ERROR] Failed to cache query response: %v", err)
	}
}

func (p *Proxy) streamServerResponse(conn net.Conn) error {
	c, err := protocol.New(conn)
	if err != nil {
		return err
	}

	buf := pool.Get()
	defer pool.Put(buf)

	for {
		buf.Reset()

		data, err := c.Read()
		if err != nil {
			return err
		}

		start, ok := p.kontext.Get("start").(bool)
		if ok && start {
			p.cacheDataPacket(data)
		}

		_, _ = buf.Write(data.Header)
		_, _ = buf.Write(data.Payload)

		_, err = io.Copy(p.client, buf)
		if err != nil {
			return err
		}

		if data.Identifier == ReadyForQueryIdentifier {
			break
		}
	}

	return nil
}

func (p *Proxy) requestToServer(conn *pgxpool.Conn, buf io.Reader) error {
	server := conn.Conn().PgConn().Conn()
	_, err := io.Copy(server, buf)
	if err != nil {
		return err
	}

	return p.streamServerResponse(server)
}

func (p *Proxy) consumeUntilSyncMessage(r *protocol.Reader, buf io.Writer) error {
	for {
		item, err := r.Read()
		if err != nil {
			return err
		}

		_, _ = buf.Write(item.Header)
		_, _ = buf.Write(item.Payload)

		if item.Identifier == SyncIdentifier {
			break
		}
	}

	return nil
}

func (p *Proxy) readFromClient(r *protocol.Reader, buf io.Writer) (bool, error) {
	data, err := r.Read()
	if err != nil {
		return false, err
	}

	switch {
	case data.Identifier == ParseIdentifier:
		servedFromCache, err := p.handleExtendedQuery(r, data)
		if err != nil {
			return false, err
		}
		if servedFromCache {
			return true, nil
		}
	case data.Identifier == QueryIdentifier:
		servedFromCache, err := p.handleSimpleQuery(data)
		if err != nil {
			return false, err
		}
		if servedFromCache {
			return true, nil
		}
	case data.Identifier == TerminateIdentifier:
		return false, ErrClientIsGone
	}

	_, _ = buf.Write(data.Header)
	_, _ = buf.Write(data.Payload)

	if data.Identifier == ParseIdentifier || data.Identifier == BindIdentifier {
		if err := p.consumeUntilSyncMessage(r, buf); err != nil {
			return false, err
		}
	}

	return false, nil
}

func (p *Proxy) sessionPooling(r *protocol.Reader) error {
	server, err := p.dbconn.Pool.Acquire(p.ctx)
	if err != nil {
		return err
	}
	defer server.Release()

	buf := pool.Get()
	defer pool.Put(buf)

	for {
		buf.Reset()

		done, err := p.readFromClient(r, buf)
		if err != nil {
			return err
		}

		if done {
			continue
		}

		err = p.requestToServer(server, buf)
		if err != nil {
			return err
		}
	}
}

func (p *Proxy) statementPooling(r *protocol.Reader) error {
	buf := pool.Get()
	defer pool.Put(buf)

	for {
		buf.Reset()

		done, err := p.readFromClient(r, buf)
		if err != nil {
			return err
		}

		if done {
			continue
		}

		server, err := p.dbconn.Pool.Acquire(p.ctx)
		if err != nil {
			return err
		}
		err = p.requestToServer(server, buf)
		if err != nil {
			return err
		}

		server.Release()
	}
}

func (p *Proxy) clientLoop() error {
	r, err := protocol.New(p.client)
	if err != nil {
		return err
	}

	switch p.dbconn.Database.ConnectionPool.Policy {
	case config.SessionConnectionPoolPolicy:
		err = p.sessionPooling(r)
	case config.StatementConnectionPoolPolicy:
		err = p.statementPooling(r)
	default:
		return fmt.Errorf("unknown connection pool policy: %s", p.dbconn.Database.ConnectionPool.Policy)
	}

	return err
}

func (p *Proxy) Close() error {
	select {
	case <-p.ctx.Done():
		return nil
	default:
	}

	p.cancel()

	var errGr errgroup.Group
	errGr.Go(func() error {
		clientErr := p.client.Close()
		if clientErr != nil {
			p.log.V(3).Printf("[ERROR] Failed to close client conn: %v", clientErr)
		}
		return clientErr
	})

	return errGr.Wait()
}
