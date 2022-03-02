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

package pgscale

import (
	"context"
	"testing"
	"time"

	"github.com/buraksezer/olric"
	"github.com/pgscale/pgscale/config"
	"github.com/pgscale/pgscale/kontext"
	"github.com/pgscale/pgscale/testutils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestPgScale_ListenAndServe(t *testing.T) {
	configFile := testutils.NewPgScaleConfig(t)
	c, err := config.New(configFile)

	olricConfig, err := config.MakeOlricConfig(c)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	callback := func() {
		defer cancel()
	}
	olricConfig.Started = callback

	db, err := olric.New(olricConfig)
	require.NoError(t, err)

	var errGr errgroup.Group
	errGr.Go(func() error {
		return db.Start()
	})

	<-ctx.Done()

	k := kontext.New()
	k.Set(kontext.OlricKey, db)
	k.Set(kontext.ConfigKey, c)

	pg, err := New(k)
	require.NoError(t, err)

	errGr.Go(func() error {
		return pg.ListenAndServe()
	})

	<-time.After(250 * time.Millisecond)

	errGr.Go(func() error {
		return pg.Shutdown()
	})

	errGr.Go(func() error {
		return db.Shutdown(context.Background())
	})

	require.NoError(t, errGr.Wait())
}
