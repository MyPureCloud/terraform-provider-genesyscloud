package resource_cache

import "sync"

type inMemoryCache[T any] struct {
	lock sync.Mutex
	data map[string]T
}

// Set stores a value in the in-memory cache
func (c *inMemoryCache[T]) Set(key string, value T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data[key] = value
}

// Get retrieves a value from the in-memory cache
func (c *inMemoryCache[T]) Get(key string) T {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.data[key]
}
