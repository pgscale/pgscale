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

package protocol

import (
	"testing"

	"github.com/jackc/pgproto3/v2"
	"github.com/pgscale/pgscale/testutils"
	"github.com/stretchr/testify/require"
)

const queryIdentifier = byte('Q')

func TestProtocol_Reader(t *testing.T) {
	conn := testutils.NewConn()
	r, err := New(conn)
	require.NoError(t, err)

	q := pgproto3.Query{
		String: "select * from users;",
	}
	encoded := q.Encode(nil)
	_, err = conn.Write(encoded)
	require.NoError(t, err)

	packet, err := r.Read()
	require.NoError(t, err)
	require.Equal(t, queryIdentifier, packet.Identifier)
	require.Equal(t, encoded, packet.Encode())
}
