package config

import (
	"encoding/json"
	"testing"

	"github.com/pgscale/pgscale/testutils"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	filePath := testutils.NewPgScaleConfigYaml(t)
	c, err := New(filePath)
	require.NoError(t, err)

	data, err := json.Marshal(c)
	require.NoError(t, err)

	var tmp interface{}
	err = json.Unmarshal(data, &tmp)
	require.NoError(t, err)
	require.Equal(t, testutils.NewPgScaleJSONConfig(t), tmp)
}

func TestConfig_PgScale_ConnString(t *testing.T) {
	cfg := testutils.NewPgScaleConfigYaml(t)
	c, err := New(cfg)
	require.NoError(t, err)

	expected := map[string]string{
		"database-name": "dbname=database-name user=postgres host=localhost port=5432 " +
			"pool_max_conn_idle_time=15m pool_max_conn_lifetime=1h " +
			"pool_health_check_period=1m pool_max_conns=50 pool_min_conns=0 ",
		"another-database": "dbname=another-database port=5432 user=postgres host=localhost " +
			"pool_max_conn_idle_time=15m pool_max_conn_lifetime=1h " +
			"pool_health_check_period=1m pool_max_conns=50 pool_min_conns=0 ",
	}
	for dbname, db := range c.PgScale.PostgreSQL.Databases {
		require.Equal(t,
			testutils.ConnStringToMap(expected[dbname]),
			testutils.ConnStringToMap(db.ConnString()),
		)
	}
}
