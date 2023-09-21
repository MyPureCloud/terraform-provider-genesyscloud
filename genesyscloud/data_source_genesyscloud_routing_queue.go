package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
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
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first queue name. Retry in case new queue is not yet indexed by search
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			queues, _, getErr := routingAPI.GetRoutingQueues(pageNum, pageSize, name, "", nil, nil, nil, false)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting queue %s: %s", name, getErr))
			}

			if queues.Entities == nil || len(*queues.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No routing queues found with name %s", name))
			}

			for _, queue := range *queues.Entities {
				if queue.Name != nil && *queue.Name == name {
					d.SetId(*queue.Id)
					return nil
				}
			}
		}
	})
}
