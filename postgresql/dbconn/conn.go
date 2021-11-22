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

package dbconn

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/buraksezer/dante/config"
	"github.com/buraksezer/dante/kontext"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrDatabaseConnNotFound = errors.New("conn not found")

type Conn struct {
	mtx sync.Mutex

	Database *config.Database
	Config   *pgxpool.Config
	Pool     *pgxpool.Pool
}

func (c *Conn) CreatePool(ctx context.Context) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.Pool != nil {
		return nil
	}

	p, err := pgxpool.ConnectConfig(ctx, c.Config)
	if err != nil {
		return err
	}
	c.Pool = p
	return nil
}

func ConnFromKontext(k *kontext.Kontext) (*Conn, error) {
	i := k.Get(kontext.DBConnKey)
	if i == nil {
		return nil, ErrDatabaseConnNotFound
	}

	dbconn, ok := i.(*Conn)
	if !ok {
		return nil, fmt.Errorf("conn: %w", kontext.ErrInvalidType)
	}

	return dbconn, nil
}
