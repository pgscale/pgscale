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
	"os"
	"strings"
)

const (
	SessionConnectionPoolPolicy   = "session"
	StatementConnectionPoolPolicy = "statement"
)

type PgScale struct {
	BindAddr   string     `yaml:"bindAddr"`
	BindPort   string     `yaml:"bindPort"`
	Auth       Auth       `yaml:"auth"`
	Logging    Logging    `yaml:"logging"`
	PostgreSQL PostgreSQL `yaml:"postgresql"`
}

type Logging struct {
	Perm      os.FileMode `yaml:"perm"`
	Verbosity int32       `yaml:"verbosity"`
	Level     string      `yaml:"level"`
	Output    string      `yaml:"output"`
}

type ConnectionPool struct {
	Policy            string  `yaml:"policy"`
	MaxConnIdleTime   *string `yaml:"maxConnIdleTime"`
	MaxConnLifetime   *string `yaml:"maxConnLifeTime"`
	HealthCheckPeriod *string `yaml:"healthCheckPeriod"`
	MinConns          *int    `yaml:"minConns"`
	MaxConns          *int    `yaml:"maxConns"`
}

type Table struct {
	DMapName        string
	Name            string  `yaml:"name"`
	MaxIdleDuration *string `yaml:"maxIdleDuration"`
	TTLDuration     *string `yaml:"ttlDuration"`
	MaxKeys         *int    `yaml:"maxKeys"`
	MaxInuse        *int    `yaml:"maxInuse"`
	LRUSamples      *int    `yaml:"lruSamples"`
	EvictionPolicy  *string `yaml:"evictionPolicy"`
	StorageEngine   *string `yaml:"storageEngine"`
}

type Schema struct {
	Schema                      string            `yaml:"schema"`
	NumEvictionWorkers          *int64            `yaml:"numEvictionWorkers"`
	MaxIdleDuration             *string           `yaml:"maxIdleDuration"`
	TTLDuration                 *string           `yaml:"ttlDuration"`
	MaxKeys                     *int              `yaml:"maxKeys"`
	MaxInuse                    *int              `yaml:"maxInuse"`
	LRUSamples                  *int              `yaml:"lruSamples"`
	EvictionPolicy              *string           `yaml:"evictionPolicy"`
	StorageEngine               *string           `yaml:"storageEngine"`
	CheckEmptyFragmentsInterval *string           `yaml:"checkEmptyFragmentsInterval"`
	Tables                      map[string]*Table `yaml:"tables"`
}

type Database struct {
	Dbname         string
	Parameters     map[string]string  `yaml:"parameters"`
	ConnectionPool ConnectionPool     `yaml:"connectionPool"`
	LogStatements  bool               `yaml:"logStatements"`
	ResetQuery     string             `yaml:"resetQuery"`
	Schemas        map[string]*Schema `yaml:"schemas"`
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
	Databases map[string]*Database `yaml:"databases"`
}

func (p *PostgreSQL) Sanitize() error {
	for dbname, db := range p.Databases {
		db.Dbname = dbname
	}

	return nil
}
