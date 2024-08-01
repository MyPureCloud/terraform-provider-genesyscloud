package resource_cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// Cache for Data Sources
type DataSourceCache struct {
	Cache            map[string]string
	mutex            sync.RWMutex
	ClientConfig     *platformclientv2.Configuration
	HydrateCacheFunc func(*DataSourceCache) error
	getApiFunc       func(*DataSourceCache, string, context.Context) (string, diag.Diagnostics)
}

// NewDataSourceCache creates a new data source cache
func NewDataSourceCache(clientConfig *platformclientv2.Configuration, hydrateFn func(*DataSourceCache) error, getFn func(*DataSourceCache, string, context.Context) (string, diag.Diagnostics)) *DataSourceCache {

	return &DataSourceCache{
		Cache:            make(map[string]string),
		ClientConfig:     clientConfig,
		HydrateCacheFunc: hydrateFn,
		getApiFunc:       getFn,
	}
}

func (c *DataSourceCache) HydrateCacheIfEmpty() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isEmpty() {
		if err := c.hydrateCache(); err != nil {
			return err
		}
	}
	return nil
}

// Hydrate the cache with updated values.
func (c *DataSourceCache) hydrateCache() error {
	return c.HydrateCacheFunc(c)
}

// Adds or updates a cache entry
func (c *DataSourceCache) UpdateCacheEntry(key string, val string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Cache == nil {
		return fmt.Errorf("cache is not initialized")
	}
	c.Cache[key] = val

	log.Printf("updated cache entry [%v] to value: %v", key, val)

	return nil
}

// Returns true if the cache is empty
func (c *DataSourceCache) isEmpty() bool {
	return len(c.Cache) <= 0
}

// Get value (resource id) from cache by key string
// If value is not found return empty string and `false`
func (c *DataSourceCache) Get(key string) (val string, isFound bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isEmpty() {
		log.Printf("cache is empty. Hydrate it first with values")
		return "", false
	}

	id, ok := c.Cache[key]
	if !ok {
		log.Printf("cache miss. cannot find key %s", key)
		return "", false
	}

	log.Printf("cache hit. found key %v in cache with value %v", key, id)
	return id, true
}

func RetrieveId(cache *DataSourceCache,
	resourceName, key string, ctx context.Context) (string, diag.Diagnostics) {

	if err := cache.HydrateCacheIfEmpty(); err != nil {
		return "", diag.FromErr(err)
	}

	// Get id from cache
	id, ok := cache.Get(key)
	if !ok {
		// If not found in cache, try to obtain through SDK call
		log.Printf("could not find the resource %v in cache. Will try API to find value", key)
		idFromApi, diagErr := cache.getApiFunc(cache, key, ctx)
		if diagErr != nil {
			return "", diagErr
		}

		if err := cache.UpdateCacheEntry(key, idFromApi); err != nil {
			return "", util.BuildDiagnosticError(resourceName, fmt.Sprintf("error updating cache"), err)
		}
		// id gets reset to empty string at the updateCacheEntry method.
		id = idFromApi
	}
	log.Printf(" id identified %v from cache", id)
	return id, nil
}
