package routing_queue

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	dataSourceRoutingQueueCache *rc.DataSourceCache
)

func dataSourceRoutingQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig

	key := d.Get("name").(string)
	key = normalizeQueueName(key)

	if dataSourceRoutingQueueCache == nil {
		dataSourceRoutingQueueCache = rc.NewDataSourceCache(sdkConfig, hydrateRoutingQueueCacheFn, getQueueByNameFn)
	}

	queueId, err := rc.RetrieveId(dataSourceRoutingQueueCache, resourceName, key, ctx)

	if err != nil {
		return err
	}

	d.SetId(queueId)
	return nil
}

// Normalize queue name for keys in the cache
func normalizeQueueName(queueName string) string {
	return strings.ToLower(queueName)
}

// hydrateRoutingQueueCacheFn for hydrating the cache with Genesys Cloud routing queues using the SDK
func hydrateRoutingQueueCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	log.Printf("hydrating cache for data source genesyscloud_routing_queues")
	proxy := GetRoutingQueueProxy(c.ClientConfig)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	queues, _, err := proxy.GetAllRoutingQueues(ctx, "")
	if err != nil {
		return err
	}

	// Add ids to cache
	for _, queue := range *queues {
		c.Cache[normalizeQueueName(*queue.Name)] = *queue.Id
	}

	log.Printf("cache hydration completed for data source genesyscloud_routing_queues")
	return nil
}

// Get queue by name.
// Returns the queue id (blank if not found) and diag
func getQueueByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := GetRoutingQueueProxy(c.ClientConfig)
	queueId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		queue, resp, retryable, getErr := proxy.getRoutingQueueByName(ctx, name)
		if getErr != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting queue %s | error %s", name, getErr), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no routing queue found with name %s", name), resp))
		}

		queueId = queue
		return nil
	})

	return queueId, diag
}
