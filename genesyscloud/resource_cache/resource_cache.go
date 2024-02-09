package resource_cache

type CacheInterface[T any] interface {
	Set(key string, value T)
	Get(key string) T
}

// NewResourceCache is a factory method to return the cache implementation. We have made this a cache so we can plugin in
func NewResourceCache[T any]() CacheInterface[T] {
	return &inMemoryCache[T]{ //This will show as a missing type in goland, but it compiles.  I think golang is have a problem resolving this
		data: make(map[string]T),
	}
}
