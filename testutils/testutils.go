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

package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/config"
	"github.com/buraksezer/olric/pkg/flog"
	"github.com/stretchr/testify/require"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func openFixtureFile(file string) []byte {
	p := path.Join(basepath, file)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		panic(fmt.Sprintf("failed to open %s: %v", p, err))
	}
	return data
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}

func NewOlricInstance(t *testing.T) *olric.Olric {
	olricPort, err := GetFreePort()
	require.NoError(t, err)

	memberlistPort, err := GetFreePort()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	c := config.New("local")
	c.Started = func() {
		defer cancel()
	}
	c.BindPort = olricPort
	c.MemberlistConfig.BindPort = memberlistPort

	db, err := olric.New(c)
	require.NoError(t, err)

	go func() {
		require.NoError(t, db.Start())
	}()

	t.Cleanup(func() {
		require.NoError(t, db.Shutdown(context.Background()))
	})

	<-ctx.Done()

	return db
}

func CreateTmpfile(t *testing.T, pattern string, data []byte) (*os.File, error) {
	if pattern == "" {
		pattern = "pgscale.*.test"
	}
	tmpfile, err := ioutil.TempFile("", pattern)
	if err != nil {
		return nil, err
	}

	if _, err := tmpfile.Write(data); err != nil {
		_ = tmpfile.Close()
		return nil, err
	}

	t.Cleanup(func() {
		err = os.Remove(tmpfile.Name())
		if err != nil {
			t.Fatalf("Expected nil. Got: %s", err)
		}
	})

	return tmpfile, nil
}

func NewCredentialsFile(t *testing.T) (string, error) {
	data := openFixtureFile("fixtures/credentials.conf")
	f, err := CreateTmpfile(t, "", data)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

func NewPgScaleConfigYaml(t *testing.T) string {
	data := openFixtureFile("fixtures/pgscale-server.yml")
	f, err := CreateTmpfile(t, "pgscale-server.*.yml", data)
	require.NoError(t, err)
	return f.Name()
}

func NewPgScaleJSONConfig(t *testing.T) interface{} {
	data := openFixtureFile("fixtures/pgscale-server.json")
	f, err := CreateTmpfile(t, "pgscale-server.*.json", data)
	require.NoError(t, err)
	f.Name()

	d, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	var value interface{}
	err = json.Unmarshal(d, &value)
	require.NoError(t, err)
	return value
}

func NewFlogLogger() *flog.Logger {
	l := log.New(os.Stdout, "", log.LstdFlags)
	lg := flog.New(l)
	lg.SetLevel(6)
	lg.ShowLineNumber(1)
	return lg
}

func ConnStringToMap(connString string) map[string]string {
	connString = strings.TrimSpace(connString)
	result := make(map[string]string)
	parsed := strings.Split(connString, " ")
	for _, item := range parsed {
		item = strings.TrimSpace(item)
		parsedItem := strings.Split(item, "=")
		result[parsedItem[0]] = parsedItem[1]
	}
	return result
}

type Conn struct {
	net.Conn
	buf *bytes.Buffer
}

func NewConn() *Conn {
	return &Conn{
		buf: bytes.NewBuffer(nil),
	}
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.buf.Write(b)
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.buf.Read(b)
}
