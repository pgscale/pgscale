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
	"os"
	"strings"
)

const (
	SessionConnectionPoolPolicy   = "session"
	StatementConnectionPoolPolicy = "statement"
)

type Dante struct {
	BindAddr   string     `hcl:"bind_addr"`
	BindPort   string     `hcl:"bind_port"`
	Auth       Auth       `hcl:"auth,block"`
	Logging    Logging    `hcl:"logging,block"`
	PostgreSQL PostgreSQL `hcl:"postgresql,block"`
}

type Logging struct {
	Perm      os.FileMode `hcl:"perm"`
	Verbosity int32       `hcl:"verbosity"`
	Level     string      `hcl:"level"`
	Output    string      `hcl:"output"`
}

type ConnectionPool struct {
	Policy            string  `hcl:"policy"`
	MaxConnIdleTime   *string `hcl:"max_conn_idle_time"`
	MaxConnLifetime   *string `hcl:"max_conn_lifetime"`
	HealthCheckPeriod *string `hcl:"health_check_period"`
	MinConns          *int    `hcl:"min_conns"`
	MaxConns          *int    `hcl:"max_conns"`
}

type Database struct {
	Dbname         string            `hcl:"dbname,label"`
	Parameters     map[string]string `hcl:"parameters"`
	ConnectionPool ConnectionPool    `hcl:"connection_pool,block"`
	LogStatements  bool              `hcl:"log_statements"`
	ResetQuery     string            `hcl:"reset_query"`
	Caches         []*Cache          `hcl:"cache,block"`
}

func (d Database) ConnString() string {
	var cs strings.Builder

	cs.WriteString(fmt.Sprintf("dbname=%s", d.Dbname))
	cs.WriteString(" ")

	for key, value := range d.Parameters {
		cs.WriteString(fmt.Sprintf("%s=%s", key, value))
		cs.WriteString(" ")
	}

	if d.ConnectionPool.MaxConnIdleTime != nil {
		cs.WriteString(fmt.Sprintf("pool_max_conn_idle_time=%s", *d.ConnectionPool.MaxConnIdleTime))
		cs.WriteString(" ")
	}

	if d.ConnectionPool.MaxConnLifetime != nil {
		cs.WriteString(fmt.Sprintf("pool_max_conn_lifetime=%s", *d.ConnectionPool.MaxConnLifetime))
		cs.WriteString(" ")
	}

	if d.ConnectionPool.HealthCheckPeriod != nil {
		cs.WriteString(fmt.Sprintf("pool_health_check_period=%s", *d.ConnectionPool.HealthCheckPeriod))
		cs.WriteString(" ")
	}

	if d.ConnectionPool.MaxConns != nil {
		cs.WriteString(fmt.Sprintf("pool_max_conns=%d", *d.ConnectionPool.MaxConns))
		cs.WriteString(" ")
	}

	if d.ConnectionPool.MinConns != nil {
		cs.WriteString(fmt.Sprintf("pool_min_conns=%d", *d.ConnectionPool.MinConns))
		cs.WriteString(" ")
	}

	return cs.String()
}

type PostgreSQL struct {
	Databases []Database `hcl:"database,block"`
}

type Table struct {
	DMapName        string
	Name            string  `hcl:"name,label"`
	MaxIdleDuration *string `hcl:"max_idle_duration"`
	TTLDuration     *string `hcl:"ttl_duration"`
	MaxKeys         *int    `hcl:"max_keys"`
	MaxInuse        *int    `hcl:"max_inuse"`
	LRUSamples      *int    `hcl:"lru_samples"`
	EvictionPolicy  *string `hcl:"eviction_policy"`
	StorageEngine   *string `hcl:"storage_engine"`
}

type Cache struct {
	Schema                      string   `hcl:"schema,label"`
	NumEvictionWorkers          *int64   `hcl:"numEvictionWorkers"`
	MaxIdleDuration             *string  `hcl:"maxIdleDuration"`
	TTLDuration                 *string  `hcl:"ttlDuration"`
	MaxKeys                     *int     `hcl:"maxKeys"`
	MaxInuse                    *int     `hcl:"maxInuse"`
	LRUSamples                  *int     `hcl:"lruSamples"`
	EvictionPolicy              *string  `hcl:"evictionPolicy"`
	StorageEngine               *string  `hcl:"storageEngine"`
	CheckEmptyFragmentsInterval *string  `hcl:"checkEmptyFragmentsInterval"`
	Tables                      []*Table `hcl:"table,block"`
}
