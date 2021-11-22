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
	"testing"

	"github.com/buraksezer/dante/kontext"
	"github.com/stretchr/testify/require"
)

func TestConfig_FileDoesntExist(t *testing.T) {
	_, err := New("foobar.hcl")
	require.Error(t, err)
}

func TestConfig_FromKontext(t *testing.T) {
	ktx := kontext.New()
	_, err := FromKontext(ktx)
	require.ErrorIs(t, ErrConfigNotFound, err)

	ktx.Set(kontext.ConfigKey, struct{}{})
	_, err = FromKontext(ktx)
	require.ErrorIs(t, err, kontext.ErrInvalidType)

	c := &Config{}
	ktx.Set(kontext.ConfigKey, c)
	extracted, err := FromKontext(ktx)
	require.Equal(t, c, extracted)
}
