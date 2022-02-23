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

package dmaps

import (
	"errors"
	"sync"

	"github.com/buraksezer/olric"
)

var errDMapNotFound = errors.New("dmap not found")

// DMaps stores all available DMap instances of this PgScale node.
type DMaps struct {
	mtx sync.RWMutex

	db *olric.Olric
	m  map[string]*olric.DMap
}

// New creates a new DMaps and returns it.
func New(db *olric.Olric) *DMaps {
	return &DMaps{
		db: db,
		m:  make(map[string]*olric.DMap),
	}
}

func (d *DMaps) getDMap(name string) (*olric.DMap, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	dm, ok := d.m[name]
	if !ok {
		return nil, errDMapNotFound
	}

	return dm, nil
}

func (d *DMaps) createDMap(name string) (*olric.DMap, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	dm, ok := d.m[name]
	if ok {
		// it has been created by a previous goroutine
		return dm, nil
	}

	// Create a new DMap here
	dm, err := d.db.NewDMap(name)
	if err != nil {
		return nil, err
	}

	d.m[name] = dm
	return dm, nil
}

// GetOrCreateDMap returns a DMap instance with the given name, otherwise it creates a new DMap on the cluster.
func (d *DMaps) GetOrCreateDMap(name string) (*olric.DMap, error) {
	dm, err := d.getDMap(name)
	if err != nil {
		return d.createDMap(name)
	}

	if err != nil {
		return nil, err
	}

	return dm, nil
}
