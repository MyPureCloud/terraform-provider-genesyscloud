package resource_cache

import (
	"log"
	"sync"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
)

// CacheInterface defines the interface for a thread-safe resource cache
type CacheInterface[T any] interface {
	GetCache() *[]T
	GetCacheItem(id string) *T
	SetCache(id string, item T)
	DeleteCacheItem(id string)
	GetCacheSize() int
	NewResourceCache() CacheInterface[T]

	// Backward compatibility methods
	Set(key string, value T)
	Get(key string) (T, bool)
	GetAll() []T
	GetSize() int
	Delete(key string)
}

// ResourceCache is a thread-safe cache for storing resources
type ResourceCache[T any] struct {
	cache map[string]T
	mutex sync.RWMutex
}

// NewResourceCache creates a new thread-safe resource cache
func NewResourceCache[T any]() CacheInterface[T] {
	return &ResourceCache[T]{
		cache: make(map[string]T),
	}
}

// GetCache returns a copy of all cached items
func (rc *ResourceCache[T]) GetCache() *[]T {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	var items []T
	for _, item := range rc.cache {
		items = append(items, item)
	}
	return &items
}

// GetCacheItem retrieves a specific item from the cache
func (rc *ResourceCache[T]) GetCacheItem(id string) *T {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if item, exists := rc.cache[id]; exists {
		return &item
	}
	return nil
}

// SetCache stores an item in the cache
func (rc *ResourceCache[T]) SetCache(id string, item T) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.cache[id] = item
}

// DeleteCacheItem removes an item from the cache
func (rc *ResourceCache[T]) DeleteCacheItem(id string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	delete(rc.cache, id)
}

// GetCacheSize returns the number of items in the cache
func (rc *ResourceCache[T]) GetCacheSize() int {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	return len(rc.cache)
}

// NewResourceCache creates a new empty cache instance
func (rc *ResourceCache[T]) NewResourceCache() CacheInterface[T] {
	return NewResourceCache[T]()
}

// Backward compatibility methods
func (rc *ResourceCache[T]) Set(key string, value T) {
	rc.SetCache(key, value)
}

func (rc *ResourceCache[T]) Get(key string) (T, bool) {
	item := rc.GetCacheItem(key)
	if item != nil {
		return *item, true
	}
	var zero T
	return zero, false
}

func (rc *ResourceCache[T]) GetAll() []T {
	items := rc.GetCache()
	if items != nil {
		return *items
	}
	return []T{}
}

func (rc *ResourceCache[T]) GetSize() int {
	return rc.GetCacheSize()
}

func (rc *ResourceCache[T]) Delete(key string) {
	rc.DeleteCacheItem(key)
}

func SetCache[T any](cache CacheInterface[T], key string, value T) {
	if tfexporter_state.IsExporterActive() {
		cache.SetCache(key, value)
	}
}

func DeleteCacheItem[T any](cache CacheInterface[T], key string) {
	if tfexporter_state.IsExporterActive() {
		cache.DeleteCacheItem(key)
	}
}

func GetCacheItem[T any](cache CacheInterface[T], key string) *T {
	if tfexporter_state.IsExporterActive() {
		eg := cache.GetCacheItem(key)
		if eg != nil {
			return eg
		}
		log.Printf("Resource Data not present in the Cache for %v, will do API call to fetch", key)
	}
	return nil
}

func GetCache[T any](cache CacheInterface[T]) *[]T {
	if tfexporter_state.IsExporterActive() {
		items := cache.GetCache()
		if items != nil && len(*items) > 0 {
			return items
		}
		log.Print("Cache is empty, will do API calls to fetch data")
	}
	return nil
}

func GetCacheSize[T any](cache CacheInterface[T]) int {
	if tfexporter_state.IsExporterActive() {
		return cache.GetCacheSize()
	}

	return 0
}
