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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
func hydrateRoutingQueueCacheFn(c *rc.DataSourceCache) error {
	log.Printf("hydrating cache for data source genesyscloud_routing_queues")
	routingApi := platformclientv2.NewRoutingApiWithConfig(c.ClientConfig)
	const pageSize = 100
	queues, _, getErr := routingApi.GetRoutingQueues(1, pageSize, "", "", nil, nil, nil, "", false)

	if getErr != nil {
		return fmt.Errorf("failed to get page of skills: %v", getErr)
	}
	if queues.Entities == nil || len(*queues.Entities) == 0 {
		return nil
	}
	for _, queue := range *queues.Entities {
		c.Cache[normalizeQueueName(*queue.Name)] = *queue.Id
	}

	for pageNum := 2; pageNum <= *queues.PageCount; pageNum++ {

		queues, _, getErr := routingApi.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, "", false)
		if getErr != nil {
			return fmt.Errorf("failed to get page of queues: %v", getErr)
		}
		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}
		// Add ids to cache
		for _, queue := range *queues.Entities {
			c.Cache[normalizeQueueName(*queue.Name)] = *queue.Id
		}
	}
	log.Printf("cache hydration completed for data source genesyscloud_routing_queues")
	return nil
}

// Get queue by name.
// Returns the queue id (blank if not found) and diag
func getQueueByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	routingApi := platformclientv2.NewRoutingApiWithConfig(c.ClientConfig)
	queueId := ""
	const pageSize = 100
	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			queues, resp, getErr := routingApi.GetRoutingQueues(pageNum, pageSize, "", name, nil, nil, nil, "", false)
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting queue %s | error %s", name, getErr), resp))
			}

			if queues.Entities == nil || len(*queues.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no routing queues found with name %s", name), resp))
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
