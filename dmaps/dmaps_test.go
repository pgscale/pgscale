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

package dmaps

import (
	"testing"

	"github.com/pgscale/pgscale/testutils"
	"github.com/stretchr/testify/require"
)

func TestDMaps_GetOrCreateDMap(t *testing.T) {
	db := testutils.NewOlricInstance(t)

	dmaps := New(db)
	dm, err := dmaps.GetOrCreateDMap("mydmap")
	require.NoError(t, err)

	require.Equal(t, "mydmap", dm.Name())
	require.Len(t, dmaps.m, 1)

	received, err := dmaps.GetOrCreateDMap("mydmap")
	require.NoError(t, err)
	require.Equal(t, dm, received)
	require.Len(t, dmaps.m, 1)
}
