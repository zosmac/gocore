// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"sync"
)

type (
	// cache defines a type for caching values by key.
	cache[K comparable, V any, F func(K) (V, error)] struct {
		sync.RWMutex
		lookup F
		values map[K]V
	}
)

// newCache creates a cache for values by key.
func newCache[K comparable, V any, F func(K) (V, error)](lookup F) *cache[K, V, F] {
	cache := &cache[K, V, F]{
		values: map[K]V{},
	}
	cache.lookup = func(key K) (V, error) {
		cache.RLock()
		value, ok := cache.values[key]
		cache.RUnlock()
		var err error
		if !ok {
			cache.Lock()
			value, err = lookup(key)
			cache.values[key] = value
			cache.Unlock()
		}

		return value, err
	}
	return cache
}
