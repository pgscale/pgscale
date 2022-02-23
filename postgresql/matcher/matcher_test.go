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

package matcher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatcher_Parse_Hierarchy(t *testing.T) {
	data := "SELECT * FROM users;"
	q, err := Parse([]byte(data))
	require.NoError(t, err)
	expected := map[string]map[string]struct{}{
		"public": {"users": {}},
	}
	require.Equal(t, expected, q.hierarchy)
}

func TestMatcher_Match(t *testing.T) {
	data := "SELECT * FROM users;"
	q, err := Parse([]byte(data))
	require.NoError(t, err)
	fmt.Println(q)
}
