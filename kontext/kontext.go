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

package kontext

import (
	"errors"
	"sync"
)

const (
	LoggerKey  = "logger"
	ConfigKey  = "config"
	OlricKey   = "olric"
	DMapsKey   = "dmaps"
	DBConnKey  = "connpool"
	SessionKey = "session"
)

var ErrInvalidType = errors.New("invalid type")

// Kontext is a simple, thread-safe key/value store to pass variables between packages
// and plugins.
type Kontext struct {
	mu sync.RWMutex
	m  map[string]interface{}
}

// New initializes and returns a new Kontext instance.
func New() *Kontext {
	return &Kontext{
		m: make(map[string]interface{}),
	}
}

// Get loads the value for a given key. If the key doesn't exist in the underlying
// map, it returns nil.
func (k *Kontext) Get(key string) interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	value, ok := k.m[key]
	if ok {
		return value
	}

	return nil
}

// Set sets a value to key. It overwrites any existing key.
func (k *Kontext) Set(key string, value interface{}) {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.m[key] = value
}

// Unset removes a particular key from Kontext.
func (k *Kontext) Unset(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	delete(k.m, key)
}

// Clear removes all the keys in this Kontext
func (k *Kontext) Clear() {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.m = make(map[string]interface{})
}

// Clone copies all keys to a new Kontext instance.
func (k *Kontext) Clone() *Kontext {
	k.mu.RLock()
	defer k.mu.RUnlock()

	f := New()
	for key, value := range k.m {
		f.Set(key, value)
	}
	return f
}
