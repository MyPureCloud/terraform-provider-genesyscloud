package resource_cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
)

// OrgCacheConfig holds configuration for creating an OrgCache.
type OrgCacheConfig[T any] struct {
	// Name is used in log messages to identify this cache instance.
	Name string
	// LoadFunc loads all org-level items. Optional when EnsureLoaded is called with an override loader.
	LoadFunc func(context.Context) ([]T, error)
	// KeyFunc extracts the cache key from an item.
	KeyFunc func(T) string
}

// OrgCache provides thread-safe, org-level load-once caching for export operations.
// Caches are shared at package scope and persist for the provider process lifetime.
type OrgCache[T any] struct {
	store    CacheInterface[T]
	loaded   bool
	mutex    sync.RWMutex
	loadFunc func(context.Context) ([]T, error)
	keyFunc  func(T) string
	name     string

	hits   atomic.Int64
	misses atomic.Int64
}

// NewOrgCache creates a new org-level cache manager.
func NewOrgCache[T any](cfg OrgCacheConfig[T]) *OrgCache[T] {
	name := cfg.Name
	if name == "" {
		name = "org-cache"
	}
	if cfg.KeyFunc == nil {
		panic("OrgCache KeyFunc is required")
	}

	return &OrgCache[T]{
		store:    NewResourceCache[T](),
		loadFunc: cfg.LoadFunc,
		keyFunc:  cfg.KeyFunc,
		name:     name,
	}
}

// Active returns true when export caching should be used.
func (oc *OrgCache[T]) Active() bool {
	return tfexporter_state.IsExporterActive()
}

// EnsureLoaded loads org-wide data once using double-checked locking.
// When export is not active, this is a no-op and returns nil.
// An optional loadFunc override can be supplied for loaders that need request-scoped dependencies.
func (oc *OrgCache[T]) EnsureLoaded(ctx context.Context, loadFunc ...func(context.Context) ([]T, error)) error {
	if !oc.Active() {
		return nil
	}

	loader := oc.loadFunc
	if len(loadFunc) > 0 && loadFunc[0] != nil {
		loader = loadFunc[0]
	}
	if loader == nil {
		return fmt.Errorf("%s load function is not configured", oc.name)
	}

	oc.mutex.RLock()
	if oc.loaded {
		oc.mutex.RUnlock()
		log.Printf("[%s] Org cache already loaded (%d items)", oc.name, oc.store.GetCacheSize())
		return nil
	}
	oc.mutex.RUnlock()

	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	if oc.loaded {
		return nil
	}

	log.Printf("[%s] Loading org-level cache (one-time operation)", oc.name)

	items, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", oc.name, err)
	}

	for _, item := range items {
		if key := oc.keyFunc(item); key != "" {
			oc.store.SetCache(key, item)
		}
	}

	oc.loaded = true
	log.Printf("[%s] Loaded %d items into org cache", oc.name, len(items))
	return nil
}

// Get retrieves an item from the cache. Callers should invoke EnsureLoaded before bulk lookups.
func (oc *OrgCache[T]) Get(key string) (*T, bool) {
	if !oc.Active() || key == "" {
		return nil, false
	}

	if item := oc.store.GetCacheItem(key); item != nil {
		oc.hits.Add(1)
		return item, true
	}

	oc.misses.Add(1)
	return nil, false
}

// Stats returns cache hit and miss counts.
func (oc *OrgCache[T]) Stats() (hits, misses int64) {
	return oc.hits.Load(), oc.misses.Load()
}

// LogLookupStats logs per-request cache resolution statistics.
func (oc *OrgCache[T]) LogLookupStats(contextLabel string, resolved, apiCalls int) {
	if !oc.Active() {
		return
	}
	hits, misses := oc.Stats()
	log.Printf("[%s] %s: resolved %d items (%d cache hits, %d misses, %d API calls, org cache size %d)",
		oc.name, contextLabel, resolved, hits, misses, apiCalls, oc.store.GetCacheSize())
}

// IsLoaded returns whether the cache has been successfully loaded.
func (oc *OrgCache[T]) IsLoaded() bool {
	if !oc.Active() {
		return false
	}

	oc.mutex.RLock()
	defer oc.mutex.RUnlock()
	return oc.loaded
}

// Size returns the number of cached items.
func (oc *OrgCache[T]) Size() int {
	if !oc.Active() {
		return 0
	}

	oc.mutex.RLock()
	defer oc.mutex.RUnlock()
	if !oc.loaded {
		return 0
	}
	return oc.store.GetCacheSize()
}
