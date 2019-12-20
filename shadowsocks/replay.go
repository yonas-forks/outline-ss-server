// Copyright 2020 Jigsaw Operations LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shadowsocks

import (
	"encoding/binary"
	"sync"
)

// Capacities in excess of 20,000 are not recommended, due to the false
// positive rate of up to 2 * capacity / 2^32 = 1 / 100,000.  If larger
// capacities are desired, the key type should be changed to uint64.
const maxCapacity = 20_000

type empty struct{}

// IVCache allows us to check whether an initialization vector was among
// the last `capacity` IVs.  It requires approximately 30*capacity bytes
// of memory (4 bytes per key plus 10.79 bytes of overhead for each set:
// https://go.googlesource.com/go/+/refs/tags/go1.13.5/src/runtime/map.go#43).
// The zero value is a cache with capacity 0, i.e. no cache.
type IVCache struct {
	mutex    sync.Mutex
	capacity int
	active   map[uint32]empty
	archive  map[uint32]empty
}

// NewIVCache returns a fresh IVCache that promises to remember at least
// the most recent `capacity` IVs.
func NewIVCache(capacity int) IVCache {
	if capacity > maxCapacity {
		panic("IVCache capacity would result in too many false positives")
	}
	return IVCache{
		capacity: capacity,
		active:   make(map[uint32]empty),
		archive:  make(map[uint32]empty),
	}
}

// Add an IV to the cache.  Returns false if the IV is already present.
func (c *IVCache) Add(iv []byte) bool {
	if c == nil || c.capacity == 0 {
		// Cache is disabled, so every IV is new.
		return true
	}
	// IVs are supposed to be random, and only authenticated IVs are added
	// to the cache.  A hostile client could produce colliding IVs, but
	// this would not impact other users.  Each map uses a new random hash
	// function, so it is not trivial for a hostile client to mount an
	// algorithmic complexity attack with nearly-colliding hashes.
	// https://dave.cheney.net/2018/05/29/how-the-go-runtime-implements-maps-efficiently-without-generics
	hash := binary.BigEndian.Uint32(iv[:4])
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.active[hash]; ok {
		// Fast replay: `iv` is already in the active set.
		return false
	}
	_, inArchive := c.archive[hash]
	c.active[hash] = empty{}
	if len(c.active) == c.capacity {
		// Discard the archive and move active to archive.
		c.archive = c.active
		c.active = make(map[uint32]empty)
	}
	return !inArchive
}
