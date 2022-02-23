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
	"encoding/json"
	"testing"

	"github.com/pgscale/pgscale/testutils"
	"github.com/stretchr/testify/require"
)

func TestConfig_PgScale_ConnString(t *testing.T) {
	cfg := testutils.NewPgScaleConfig(t)
	c, err := New(cfg)
	require.NoError(t, err)

	expected := map[int]string{
		0: "dbname=postgres user=postgres host=localhost port=5432 " +
			"pool_max_conn_idle_time=15m pool_max_conn_lifetime=1h " +
			"pool_health_check_period=1m pool_max_conns=50 pool_min_conns=0 ",
		1: "dbname=somedatabase port=5432 user=postgres host=localhost " +
			"pool_max_conn_idle_time=15m pool_max_conn_lifetime=1h " +
			"pool_health_check_period=1m pool_max_conns=50 pool_min_conns=0 ",
	}
	for i, db := range c.PgScale.PostgreSQL.Databases {
		require.Equal(t,
			testutils.ConnStringToMap(expected[i]),
			testutils.ConnStringToMap(db.ConnString()),
		)
	}
}

func TestConfig_PgScale(t *testing.T) {
	cfg := testutils.NewPgScaleConfig(t)
	c, err := New(cfg)
	require.NoError(t, err)

	data, err := json.Marshal(c.PgScale)
	require.NoError(t, err)

	var tmp interface{}
	err = json.Unmarshal(data, &tmp)
	require.NoError(t, err)
	require.Equal(t, testutils.NewPgScaleJSONConfig(t), tmp)
}
