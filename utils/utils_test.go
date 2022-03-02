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

package utils

import (
	"crypto/rand"
	"testing"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/pkg/flog"
	"github.com/pgscale/pgscale/dmaps"
	"github.com/pgscale/pgscale/kontext"
	"github.com/stretchr/testify/require"
)

func TestUtils_StartWithSelect(t *testing.T) {
	q1 := []byte("select * from users;")
	require.True(t, StartWithSelect(q1))

	q2 := []byte("SELECT * FROM USERS;")
	require.True(t, StartWithSelect(q2))

	q3 := []byte("SET")
	require.False(t, StartWithSelect(q3))
}

func TestUtils_ByteToString(t *testing.T) {
	str := "PostgreSQL-Olric"
	data := []byte(str)
	require.Equal(t, str, ByteToString(data))
}

func TestUtils_TrimNULChar(t *testing.T) {
	data := make([]byte, 10)
	data[0] = NULByte
	_, err := rand.Read(data[1:])
	require.NoError(t, err)

	trimmed := TrimNULChar(data)
	require.NotEqual(t, trimmed[0], data[0])

	trimmed = TrimNULChar(trimmed)
	require.NotEqual(t, trimmed[0], NULByte)
}

func TestUtils_DMapsFromKontext(t *testing.T) {
	ktx := kontext.New()
	_, err := DMapsFromKontext(ktx)
	require.Equal(t, ErrDMapsNotFound, err)

	ktx.Set(kontext.DMapsKey, struct{}{})
	_, err = DMapsFromKontext(ktx)
	require.ErrorIs(t, err, kontext.ErrInvalidType)

	dms := dmaps.New(nil)
	ktx.Set(kontext.DMapsKey, dms)
	extracted, err := DMapsFromKontext(ktx)
	require.NoError(t, err)
	require.Equal(t, dms, extracted)
}

func TestUtils_FlogFromKontext(t *testing.T) {
	ktx := kontext.New()
	_, err := FlogFromKontext(ktx)
	require.Equal(t, ErrLoggerNotFound, err)

	ktx.Set(kontext.LoggerKey, struct{}{})
	_, err = FlogFromKontext(ktx)
	require.ErrorIs(t, err, kontext.ErrInvalidType)

	f := flog.New(nil)
	ktx.Set(kontext.LoggerKey, f)
	extracted, err := FlogFromKontext(ktx)
	require.NoError(t, err)
	require.Equal(t, f, extracted)
}

func TestUtils_OlricFromKontext(t *testing.T) {
	ktx := kontext.New()
	_, err := OlricFromKontext(ktx)
	require.Equal(t, ErrOlricNotFound, err)

	ktx.Set(kontext.OlricKey, struct{}{})
	_, err = OlricFromKontext(ktx)
	require.ErrorIs(t, err, kontext.ErrInvalidType)

	db := &olric.Olric{}
	ktx.Set(kontext.OlricKey, db)
	extracted, err := OlricFromKontext(ktx)
	require.NoError(t, err)
	require.Equal(t, db, extracted)
}
