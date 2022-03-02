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
	"errors"
	"fmt"
	"os"

	"github.com/pgscale/pgscale/kontext"
	"gopkg.in/yaml.v3"
)

var ErrConfigNotFound = errors.New("no config key found")

const (
	DebugLog = "DEBUG"
	WarnLog  = "WARN"
	ErrorLog = "ERROR"
	InfoLog  = "INFO"
)
const (
	Stdout = "stdout"
	Stderr = "stderr"
)

type Config struct {
	PgScale PgScale `yaml:"pgscale"`
	Olric   Olric   `yaml:"olric"`
}

func (c *Config) Sanitize() error {
	if err := c.PgScale.PostgreSQL.Sanitize(); err != nil {
		return err
	}
	return nil
}

func FromKontext(k *kontext.Kontext) (*Config, error) {
	i := k.Get(kontext.ConfigKey)
	if i == nil {
		return nil, ErrConfigNotFound
	}

	c, ok := i.(*Config)
	if !ok {
		return nil, fmt.Errorf("config: %w", kontext.ErrInvalidType)
	}
	return c, nil
}

func New(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file does not exist: %s", filename)
	}
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open configuration file: %s: %w", filename, err)
	}

	var c Config
	err = yaml.NewDecoder(r).Decode(&c)
	if err != nil {
		return nil, err
	}

	if err = c.Sanitize(); err != nil {
		return nil, err
	}

	return &c, nil
}
