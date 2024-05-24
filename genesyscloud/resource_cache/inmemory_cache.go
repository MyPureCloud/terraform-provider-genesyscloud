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

func (c *inMemoryCache[T]) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.data, key)
}

// Get retrieves a value from the in-memory cache
func (c *inMemoryCache[T]) Get(key string) (T, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	value, ok := c.data[key]
	return value, ok
}

// GetAll retrieves all the values from the in-memory cache
func (c *inMemoryCache[T]) GetAll() []T {
	c.lock.Lock()
	defer c.lock.Unlock()

	var items []T
	for _, item := range c.data {
		items = append(items, item)
	}

	return items
}

// GetSize retrieves the size of the in-memory cache
func (c *inMemoryCache[T]) GetSize() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return len(c.data)
}
