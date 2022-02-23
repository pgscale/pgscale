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

package config

import (
	"fmt"
	"time"

	olricConfig "github.com/buraksezer/olric/config"
)

func prepareDMapConfig(c *Config) (*olricConfig.DMaps, error) {
	ds := olricConfig.DMaps{
		Custom: make(map[string]olricConfig.DMap),
	}

	for _, db := range c.PgScale.PostgreSQL.Databases {
		for _, cache := range db.Caches {
			df := olricConfig.DMaps{}
			if cache.NumEvictionWorkers != nil {
				df.NumEvictionWorkers = *cache.NumEvictionWorkers
			}

			if cache.MaxIdleDuration != nil {
				maxIdleDuration, err := time.ParseDuration(*cache.MaxIdleDuration)
				if err != nil {
					return nil, err
				}
				df.MaxIdleDuration = maxIdleDuration
			}

			if cache.TTLDuration != nil {
				ttlDuration, err := time.ParseDuration(*cache.TTLDuration)
				if err != nil {
					return nil, err
				}
				df.TTLDuration = ttlDuration
			}

			if cache.MaxKeys != nil {
				df.MaxKeys = *cache.MaxKeys
			}

			if cache.MaxInuse != nil {
				df.MaxInuse = *cache.MaxInuse
			}

			if cache.LRUSamples != nil {
				df.LRUSamples = *cache.LRUSamples
			}

			if cache.EvictionPolicy != nil {
				df.EvictionPolicy = olricConfig.EvictionPolicy(*cache.EvictionPolicy)
			}

			if cache.StorageEngine != nil {
				df.StorageEngine = *cache.StorageEngine
			}

			for _, table := range cache.Tables {
				dm := olricConfig.DMap{}
				if table.MaxIdleDuration != nil {
					maxIdleDuration, err := time.ParseDuration(*table.MaxIdleDuration)
					if err != nil {
						return nil, err
					}
					dm.MaxIdleDuration = maxIdleDuration
				} else {
					dm.MaxIdleDuration = df.MaxIdleDuration
				}

				if table.TTLDuration != nil {
					ttlDuration, err := time.ParseDuration(*table.TTLDuration)
					if err != nil {
						return nil, err
					}
					dm.TTLDuration = ttlDuration
				} else {
					dm.TTLDuration = df.TTLDuration
				}

				if table.MaxKeys != nil {
					dm.MaxKeys = *table.MaxKeys
				} else {
					dm.MaxKeys = df.MaxKeys
				}

				if table.MaxInuse != nil {
					dm.MaxInuse = *table.MaxInuse
				} else {
					dm.MaxInuse = df.MaxInuse
				}

				if table.LRUSamples != nil {
					dm.LRUSamples = *table.LRUSamples
				} else {
					dm.LRUSamples = df.LRUSamples
				}

				if table.EvictionPolicy != nil {
					dm.EvictionPolicy = olricConfig.EvictionPolicy(*table.EvictionPolicy)
				} else {
					dm.EvictionPolicy = df.EvictionPolicy
				}

				table.DMapName = fmt.Sprintf("%s.%s.%s", db.Dbname, cache.Schema, table.Name)
				ds.Custom[table.DMapName] = dm
			}
		}
	}

	return &ds, nil
}
