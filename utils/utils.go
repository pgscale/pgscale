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

package utils

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/pkg/flog"
	"github.com/pgscale/pgscale/dmaps"
	"github.com/pgscale/pgscale/kontext"
)

const NULByte = byte(0)

var (
	ErrDMapsNotFound  = errors.New("dmaps not found")
	ErrOlricNotFound  = errors.New("olric not found")
	ErrLoggerNotFound = errors.New("logger not found")
)

func OlricFromKontext(k *kontext.Kontext) (*olric.Olric, error) {
	i := k.Get(kontext.OlricKey)
	if i == nil {
		return nil, ErrOlricNotFound
	}

	db, ok := i.(*olric.Olric)
	if !ok {
		return nil, fmt.Errorf("olric: %w", kontext.ErrInvalidType)
	}

	return db, nil
}

func FlogFromKontext(k *kontext.Kontext) (*flog.Logger, error) {
	i := k.Get(kontext.LoggerKey)
	if i == nil {
		return nil, ErrLoggerNotFound
	}

	f, ok := i.(*flog.Logger)
	if !ok {
		return nil, fmt.Errorf("logger: %w", kontext.ErrInvalidType)
	}

	return f, nil
}

func DMapsFromKontext(k *kontext.Kontext) (*dmaps.DMaps, error) {
	i := k.Get(kontext.DMapsKey)
	if i == nil {
		return nil, ErrDMapsNotFound
	}

	d, ok := i.(*dmaps.DMaps)
	if !ok {
		return nil, fmt.Errorf("dmaps: %w", kontext.ErrInvalidType)
	}

	return d, nil
}

func TrimNULChar(payload []byte) []byte {
	if len(payload) >= 1 && payload[0] == NULByte {
		return payload[1:]
	}

	return payload
}

func ByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StartWithSelect(data []byte) bool {
	if len(data) < 6 {
		return false
	}

	if data[0] == byte('S') && data[1] == byte('E') && data[2] == byte('L') {
		return true
	}

	if data[0] == byte('s') && data[1] == byte('e') && data[2] == byte('l') {
		return true
	}

	return false
}
