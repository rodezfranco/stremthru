package cache

import (
	"sync"
	"time"

	"github.com/elastic/go-freelru"
	"github.com/zeebo/xxh3"
)

type LRUCache[V any] struct {
	c    *freelru.LRU[string, V]
	name string
	m    sync.Mutex
}

func (cache *LRUCache[V]) GetName() string {
	return cache.name
}

func (cache *LRUCache[V]) Add(key string, value V) error {
	cache.m.Lock()
	defer cache.m.Unlock()

	cache.c.Add(key, value)
	return nil
}

func (cache *LRUCache[V]) AddWithLifetime(key string, value V, lifetime time.Duration) error {
	cache.m.Lock()
	defer cache.m.Unlock()

	cache.c.AddWithLifetime(key, value, lifetime)
	return nil
}

func (cache *LRUCache[V]) Get(key string, value *V) bool {
	cache.m.Lock()
	defer cache.m.Unlock()

	val, ok := cache.c.Get(key)
	*value = val
	return ok
}

func (cache *LRUCache[V]) Remove(key string) {
	cache.m.Lock()
	defer cache.m.Unlock()

	cache.c.Remove(key)
}

func CacheHashKeyString(key string) uint32 {
	return uint32(xxh3.HashString(key))
}

func NewLRUCache[V any](config *CacheConfig) *LRUCache[V] {
	if config.LocalCapacity == 0 {
		config.LocalCapacity = 1024
	}

	lru, err := freelru.New[string, V](config.LocalCapacity, CacheHashKeyString)
	if err != nil {
		errMsg := "failed to create cache"
		if config.Name != "" {
			errMsg += ": " + config.Name
		}
		panic(errMsg)
	}
	if config.Lifetime != 0 {
		lru.SetLifetime(config.Lifetime)
	}
	cache := &LRUCache[V]{c: lru, name: config.Name}
	return cache
}
