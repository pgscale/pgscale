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
	"errors"

	"github.com/buraksezer/dante/config"
	"github.com/buraksezer/dante/postgresql/matcher"
	"github.com/buraksezer/dante/postgresql/protocol"
	"github.com/buraksezer/dante/utils"
	"github.com/buraksezer/olric"
)

func (p *Proxy) cacheExtendedQuery(table *config.Table, c *protocol.Reader, data *protocol.DataPacket) (bool, error) {
	value, err := p.loadFromCache(table, data)
	if errors.Is(err, olric.ErrKeyNotFound) {
		return false, nil
	}

	if errors.Is(err, ErrGetOrCreateDMap) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	for {
		item, err := c.Read()
		if err != nil {
			return false, err
		}
		if item.Identifier == SyncIdentifier {
			break
		}
	}

	return p.serveFromCache(value)
}

func (p *Proxy) handleExtendedQuery(c *protocol.Reader, data *protocol.DataPacket) (bool, error) {
	var servedFromCache bool
	var err error

	payload := utils.TrimNULChar(data.Payload)
	if p.dbconn.Database.LogStatements {
		p.log.V(1).Printf("[INFO] Extended query statement: %s", utils.ByteToString(payload))
	}

	if !utils.StartWithSelect(payload) {
		return false, nil
	}

	query, err := matcher.Parse(payload)
	if err != nil {
		return false, err
	}

	return query.Match(p.dbconn.Database.Caches, func(table *config.Table) (bool, error) {
		servedFromCache, err = p.cacheExtendedQuery(table, c, data)
		if err != nil {
			return false, err
		}
		if servedFromCache {
			p.log.V(3).Printf("[INFO] Extended query result fetched from cache. Statement: %s", utils.ByteToString(payload))
		}
		return servedFromCache, nil
	})
}
