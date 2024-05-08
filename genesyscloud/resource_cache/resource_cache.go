package resource_cache

import (
	"log"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
)

type CacheInterface[T any] interface {
	Set(key string, value T)
	Get(key string) (T, bool)
	GetAll() []T
	GetSize() int
	Delete(key string)
}

// NewResourceCache is a factory method to return the cache implementation. We have made this a cache so we can plugin in
func NewResourceCache[T any]() CacheInterface[T] {
	return &inMemoryCache[T]{ //This will show as a missing type in goland, but it compiles.  I think golang is have a problem resolving this
		data: make(map[string]T),
	}
}

func SetCache[T any](cache CacheInterface[T], key string, value T) {
	if tfexporter_state.IsExporterActive() {
		cache.Set(key, value)
	}
}

func DeleteCacheItem[T any](cache CacheInterface[T], key string) {
	if tfexporter_state.IsExporterActive() {
		cache.Delete(key)
	}
}

func GetCacheItem[T any](cache CacheInterface[T], key string) *T {
	if tfexporter_state.IsExporterActive() {
		eg, ok := cache.Get(key)
		if ok {
			return &eg
		}
		log.Printf("Resource Data not present in the Cache for %v, will do API call to fetch", key)
	}
	return nil
}

func GetCache[T any](cache CacheInterface[T]) *[]T {
	if tfexporter_state.IsExporterActive() {
		items := cache.GetAll()
		if items != nil && len(items) > 0 {
			return &items
		}
		log.Print("Cache is empty, will do API calls to fetch data")
	}
	return nil
}

func GetCacheSize[T any](cache CacheInterface[T]) int {
	if tfexporter_state.IsExporterActive() {
		return cache.GetSize()
	}

	return 0
}
