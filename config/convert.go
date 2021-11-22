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
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	olricConfig "github.com/buraksezer/olric/config"
	"github.com/buraksezer/olric/hasher"
	"github.com/buraksezer/olric/serializer"
	"github.com/pkg/errors"
)

func durationCondition(name string) bool {
	return strings.HasSuffix(name, "Duration") ||
		strings.HasSuffix(name, "Timeout") ||
		strings.HasSuffix(name, "Period")
}

func keepaliveCondition(name string, field reflect.Value) bool {
	return strings.EqualFold(name, "KEEPALIVE") && field.Kind() == reflect.Int64
}

// mapYamlToConfig maps a parsed yaml to related configuration struct.
// TODO: Use this to create Olric and memberlist config from yaml file.
func mapYamlToConfig(rawDst, rawSrc interface{}) error {
	dst := reflect.ValueOf(rawDst).Elem()
	src := reflect.ValueOf(rawSrc).Elem()
	for j := 0; j < src.NumField(); j++ {
		for i := 0; i < dst.NumField(); i++ {
			if src.Type().Field(j).Name == dst.Type().Field(i).Name {
				if src.Field(j).Kind() == dst.Field(i).Kind() {
					dst.Field(i).Set(src.Field(j))
					continue
				}
				// Special cases
				name := src.Type().Field(j).Name
				if src.Field(j).Kind() == reflect.String && !src.Field(j).IsZero() {
					if durationCondition(name) || keepaliveCondition(name, dst.Field(i)) {
						value, err := time.ParseDuration(src.Field(j).String())
						if err != nil {
							return err
						}
						dst.Field(i).Set(reflect.ValueOf(value))
					}
				}
			}
		}
	}
	return nil
}

// MakeOlricConfig reads and loads Olric configuration.
func MakeOlricConfig(c *Config) (*olricConfig.Config, error) {
	var logOutput io.Writer
	switch {
	case c.Olric.Logging.Output == "stderr":
		logOutput = os.Stderr
	case c.Olric.Logging.Output == "stdout":
		logOutput = os.Stdout
	default:
		logOutput = os.Stderr
	}

	if c.Olric.Logging.Level == "" {
		c.Olric.Logging.Level = olricConfig.DefaultLogLevel
	}

	// Default serializer is Gob serializer, just set nil or use gob keyword to use it.
	var s serializer.Serializer
	switch {
	case c.Olric.Serializer == "json":
		s = serializer.NewJSONSerializer()
	case c.Olric.Serializer == "msgpack":
		s = serializer.NewMsgpackSerializer()
	case c.Olric.Serializer == "gob":
		s = serializer.NewGobSerializer()
	default:
		return nil, fmt.Errorf("invalid serializer: %s", c.Olric.Serializer)
	}

	rawMc, err := olricConfig.NewMemberlistConfig(c.Olric.Memberlist.Environment)
	if err != nil {
		return nil, err
	}

	memberlistConfig, err := prepareMemberlistConfig(c.Olric, rawMc)
	if err != nil {
		return nil, err
	}

	var joinRetryInterval, keepAlivePeriod, bootstrapTimeout, routingTablePushInterval time.Duration
	if c.Olric.KeepAlivePeriod != "" {
		keepAlivePeriod, err = time.ParseDuration(c.Olric.KeepAlivePeriod)
		if err != nil {
			return nil, errors.WithMessage(err,
				fmt.Sprintf("failed to parse olricd.keepAlivePeriod: '%s'", c.Olric.KeepAlivePeriod))
		}
	}
	if c.Olric.BootstrapTimeout != "" {
		bootstrapTimeout, err = time.ParseDuration(c.Olric.BootstrapTimeout)
		if err != nil {
			return nil, errors.WithMessage(err,
				fmt.Sprintf("failed to parse olricd.bootstrapTimeout: '%s'", c.Olric.BootstrapTimeout))
		}
	}
	if c.Olric.Memberlist.JoinRetryInterval != "" {
		joinRetryInterval, err = time.ParseDuration(c.Olric.Memberlist.JoinRetryInterval)
		if err != nil {
			return nil, errors.WithMessage(err,
				fmt.Sprintf("failed to parse memberlist.joinRetryInterval: '%s'",
					c.Olric.Memberlist.JoinRetryInterval))
		}
	}
	if c.Olric.RoutingTablePushInterval != "" {
		routingTablePushInterval, err = time.ParseDuration(c.Olric.RoutingTablePushInterval)
		if err != nil {
			return nil, errors.WithMessage(err,
				fmt.Sprintf("failed to parse olricd.routingTablePushInterval: '%s'", c.Olric.RoutingTablePushInterval))
		}
	}

	clientConfig := olricConfig.Client{}
	err = mapYamlToConfig(&clientConfig, &c.Olric.Client)
	if err != nil {
		return nil, err
	}

	storageEngines := olricConfig.NewStorageEngine()
	if c.Olric.StorageEngines.Plugins != nil {
		storageEngines.Plugins = append(storageEngines.Plugins, *c.Olric.StorageEngines.Plugins...)
	}
	storageEngines.Config = map[string]map[string]interface{}{
		"kvstore": {
			"tableSize": 1235532,
		},
	}

	ds, err := prepareDMapConfig(c)
	if err != nil {
		return nil, err
	}

	cfg := &olricConfig.Config{
		BindAddr:                 c.Olric.BindAddr,
		BindPort:                 c.Olric.BindPort,
		Serializer:               s,
		MemberlistConfig:         memberlistConfig,
		Client:                   &clientConfig,
		LogLevel:                 c.Olric.Logging.Level,
		JoinRetryInterval:        joinRetryInterval,
		RoutingTablePushInterval: routingTablePushInterval,
		MaxJoinAttempts:          c.Olric.Memberlist.MaxJoinAttempts,
		Peers:                    c.Olric.Memberlist.Peers,
		PartitionCount:           c.Olric.PartitionCount,
		ReplicaCount:             c.Olric.ReplicaCount,
		WriteQuorum:              c.Olric.WriteQuorum,
		ReadQuorum:               c.Olric.ReadQuorum,
		ReplicationMode:          c.Olric.ReplicationMode,
		ReadRepair:               c.Olric.ReadRepair,
		MemberCountQuorum:        c.Olric.MemberCountQuorum,
		Logger:                   log.New(logOutput, "", log.LstdFlags),
		LogOutput:                logOutput,
		LogVerbosity:             c.Olric.Logging.Verbosity,
		Hasher:                   hasher.NewDefaultHasher(),
		KeepAlivePeriod:          keepAlivePeriod,
		BootstrapTimeout:         bootstrapTimeout,
		DMaps:                    ds,
		StorageEngines:           storageEngines,
	}
	if c.Olric.Interface != nil {
		cfg.Interface = *c.Olric.Interface
	}

	if c.Olric.Memberlist.Interface != nil {
		cfg.MemberlistInterface = *c.Olric.Memberlist.Interface
	}

	if c.Olric.LoadFactor != nil {
		cfg.LoadFactor = *c.Olric.LoadFactor
	}

	if err := cfg.Sanitize(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
