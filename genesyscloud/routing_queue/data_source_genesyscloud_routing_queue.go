package routing_queue

import (
	"context"
	"fmt"
	"log"
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

	if dataSourceRoutingQueueCache == nil {
		log.Printf("Instantiating the %s data source cache object", ResourceType)
		dataSourceRoutingQueueCache = rc.NewDataSourceCache(sdkConfig, hydrateRoutingQueueCacheFn, getQueueByNameFn)
	}

	queueId, err := rc.RetrieveId(dataSourceRoutingQueueCache, ResourceType, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(queueId)
	return nil
}

// hydrateRoutingQueueCacheFn for hydrating the cache with Genesys Cloud routing queues using the SDK
func hydrateRoutingQueueCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := GetRoutingQueueProxy(c.ClientConfig)

	log.Printf("Hydrating cache for data source %s", ResourceType)

	allQueues, resp, err := proxy.GetAllRoutingQueues(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get routing queues. Error: %s | API Response: %s", err.Error(), resp.String())
	}

	if allQueues == nil || len(*allQueues) == 0 {
		log.Printf("No queues found. The cache will remain empty.")
		return nil
	}

	for _, queue := range *allQueues {
		c.Cache[*queue.Name] = *queue.Id
	}

	log.Printf("Cache hydration complete for data source %s", ResourceType)
	return nil
}

// getQueueByNameFn returns the queue id (blank if not found) and diag
func getQueueByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := GetRoutingQueueProxy(c.ClientConfig)
	queueId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		queueID, resp, retryable, getErr := proxy.getRoutingQueueByName(ctx, name)
		if getErr != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting queue %s | error %s", name, getErr), resp)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)
		}

		queueId = queueID
		return nil
	})

	return queueId, diag
}
