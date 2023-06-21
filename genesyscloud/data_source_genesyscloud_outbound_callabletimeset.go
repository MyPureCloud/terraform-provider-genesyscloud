package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

func dataSourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Clound Outbound Callable Timesets. Select a callable timeset by name.",
		ReadContext: ReadWithPooledClient(dataSourceOutboundCallabletimesetRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Callable timeset name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundCallabletimesetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	timesetName := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100

			timesets, _, getErr := outboundAPI.GetOutboundCallabletimesets(pageSize, pageNum, true, "", "", "", "")
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting callable timeset %s: %s", timesetName, getErr))
			}
			if timesets.Entities == nil || len(*timesets.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no callable timeset found with timesetName %s", timesetName))
			}

			for _, timeset := range *timesets.Entities {
				if timeset.Name != nil && *timeset.Name == timesetName {
					d.SetId(*timeset.Id)
					return nil
				}
			}
		}
	})
}
