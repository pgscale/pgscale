// Copyright 2021-2022 Burak Sezer
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
		for _, schema := range db.Schemas {
			globalDMapConfig := olricConfig.DMaps{}
			if schema.NumEvictionWorkers != nil {
				globalDMapConfig.NumEvictionWorkers = *schema.NumEvictionWorkers
			}

			if schema.MaxIdleDuration != nil {
				maxIdleDuration, err := time.ParseDuration(*schema.MaxIdleDuration)
				if err != nil {
					return nil, err
				}
				globalDMapConfig.MaxIdleDuration = maxIdleDuration
			}

			if schema.TTLDuration != nil {
				ttlDuration, err := time.ParseDuration(*schema.TTLDuration)
				if err != nil {
					return nil, err
				}
				globalDMapConfig.TTLDuration = ttlDuration
			}

			if schema.MaxKeys != nil {
				globalDMapConfig.MaxKeys = *schema.MaxKeys
			}

			if schema.MaxInuse != nil {
				globalDMapConfig.MaxInuse = *schema.MaxInuse
			}

			if schema.LRUSamples != nil {
				globalDMapConfig.LRUSamples = *schema.LRUSamples
			}

			if schema.EvictionPolicy != nil {
				globalDMapConfig.EvictionPolicy = olricConfig.EvictionPolicy(*schema.EvictionPolicy)
			}

			if schema.StorageEngine != nil {
				globalDMapConfig.StorageEngine = *schema.StorageEngine
			}

			for _, table := range schema.Tables {
				dm := olricConfig.DMap{}
				if table.MaxIdleDuration != nil {
					maxIdleDuration, err := time.ParseDuration(*table.MaxIdleDuration)
					if err != nil {
						return nil, err
					}
					dm.MaxIdleDuration = maxIdleDuration
				} else {
					dm.MaxIdleDuration = globalDMapConfig.MaxIdleDuration
				}

				if table.TTLDuration != nil {
					ttlDuration, err := time.ParseDuration(*table.TTLDuration)
					if err != nil {
						return nil, err
					}
					dm.TTLDuration = ttlDuration
				} else {
					dm.TTLDuration = globalDMapConfig.TTLDuration
				}

				if table.MaxKeys != nil {
					dm.MaxKeys = *table.MaxKeys
				} else {
					dm.MaxKeys = globalDMapConfig.MaxKeys
				}

				if table.MaxInuse != nil {
					dm.MaxInuse = *table.MaxInuse
				} else {
					dm.MaxInuse = globalDMapConfig.MaxInuse
				}

				if table.LRUSamples != nil {
					dm.LRUSamples = *table.LRUSamples
				} else {
					dm.LRUSamples = globalDMapConfig.LRUSamples
				}

				if table.EvictionPolicy != nil {
					dm.EvictionPolicy = olricConfig.EvictionPolicy(*table.EvictionPolicy)
				} else {
					dm.EvictionPolicy = globalDMapConfig.EvictionPolicy
				}

				table.DMapName = fmt.Sprintf("%s.%s.%s", db.Dbname, schema.Schema, table.Name)
				ds.Custom[table.DMapName] = dm
			}
		}
	}

	return &ds, nil
}
