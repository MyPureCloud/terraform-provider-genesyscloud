package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// Cache for Data Sources
type DataSourceCache struct {
	cache        map[string]string
	mutex        sync.RWMutex
	clientConfig *platformclientv2.Configuration

	hydrateCacheFunc func(*DataSourceCache) error
}

var (
	dataSourceRoutingQueueCache *DataSourceCache
)

func DataSourceRoutingQueue() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Queues. Select a queue by name.",
		ReadContext: ReadWithPooledClient(dataSourceRoutingQueueRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Queue name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// Create a cache for the queues
	if dataSourceRoutingQueueCache == nil {
		dataSourceRoutingQueueCache = NewDataSourceCache(sdkConfig, hydrateRoutingQueueCacheFn)
	}

	if err := dataSourceRoutingQueueCache.hydrateCacheIfEmpty(); err != nil {
		return diag.FromErr(err)
	}

	// Get id from cache
	name := d.Get("name").(string)
	queueId, ok := dataSourceRoutingQueueCache.get(normalizeQueueName(name))
	if !ok {
		// If not found in cache, try to obtain through SDK call
		log.Printf("could not find routing queue %v in cache. Will try API to find value", name)
		queueId, diagErr := getQueueByName(ctx, routingApi, name)
		if diagErr != nil {
			return diagErr
		}

		d.SetId(queueId)
		if err := dataSourceRoutingQueueCache.updateCacheEntry(name, queueId); err != nil {
			return diag.Errorf("error updating cache: %v", err)
		}
		return nil
	}

	log.Printf("found queue %v from cache", name)
	d.SetId(queueId)
	return nil
}

func (c *DataSourceCache) hydrateCacheIfEmpty() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isEmpty() {
		if err := c.hydrateCache(); err != nil {
			return err
		}
	}
	return nil
}

// Normalize queue name for keys in the cache
func normalizeQueueName(queueName string) string {
	return strings.ToLower(queueName)
}

// NewDataSourceCache creates a new data source cache
func NewDataSourceCache(clientConfig *platformclientv2.Configuration, hydrateFn func(*DataSourceCache) error) *DataSourceCache {
	return &DataSourceCache{
		cache:            make(map[string]string),
		clientConfig:     clientConfig,
		hydrateCacheFunc: hydrateFn,
	}
}

// hydrateRoutingQueueCacheFn for hydrating the cache with Genesys Cloud routing queues using the SDK
func hydrateRoutingQueueCacheFn(c *DataSourceCache) error {
	log.Printf("hydrating cache for data source genesyscloud_routing_queues")

	routingApi := platformclientv2.NewRoutingApiWithConfig(c.clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		queues, _, getErr := routingApi.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, false)
		if getErr != nil {
			return fmt.Errorf("failed to get page of queues: %v", getErr)
		}

		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}

		// Add ids to cache
		for _, queue := range *queues.Entities {
			c.cache[normalizeQueueName(*queue.Name)] = *queue.Id
		}
	}

	log.Printf("cache hydration completed for data source genesyscloud_routing_queues")

	return nil
}

// Get queue by name.
// Returns the queue id (blank if not found) and diag
func getQueueByName(ctx context.Context, routingApi *platformclientv2.RoutingApi, name string) (string, diag.Diagnostics) {
	queueId := ""
	diag := WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			queues, _, getErr := routingApi.GetRoutingQueues(pageNum, pageSize, "", name, nil, nil, nil, false)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("error requesting queue %s: %s", name, getErr))
			}

			if queues.Entities == nil || len(*queues.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no routing queues found with name %s", name))
			}

			for _, queue := range *queues.Entities {
				if queue.Name != nil && normalizeQueueName(*queue.Name) == normalizeQueueName(name) {
					queueId = *queue.Id
					return nil
				}
			}
		}
	})

	return queueId, diag
}

// Hydrate the cache with updated values.
func (c *DataSourceCache) hydrateCache() error {
	return c.hydrateCacheFunc(c)
}

// Adds or updates a cache entry
func (c *DataSourceCache) updateCacheEntry(key string, val string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache == nil {
		return fmt.Errorf("cache is not initialized")
	}
	c.cache[key] = val

	log.Printf("updated cache entry [%v] to value: %v", key, val)

	return nil
}

// Returns true if the cache is empty
func (c *DataSourceCache) isEmpty() bool {
	return len(c.cache) <= 0
}

// Get value (resource id) from cache by key string
// If value is not found return empty string and `false`
func (c *DataSourceCache) get(key string) (val string, isFound bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isEmpty() {
		log.Printf("cache is empty. Hydrate it first with values")
		return "", false
	}

	queueId, ok := c.cache[key]
	if !ok {
		log.Printf("cache miss. cannot find key %s", key)
		return "", false
	}

	log.Printf("cache hit. found key %v in cache with value %v", key, queueId)
	return queueId, true
}
