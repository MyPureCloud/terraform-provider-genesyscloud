package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Clound Outbound Callable Timesets. Select a callable timeset by name.",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceOutboundCallabletimesetRead),
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
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	timesetName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100

			timesets, _, getErr := outboundAPI.GetOutboundCallabletimesets(pageSize, pageNum, true, "", "", "", "")
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("error requesting callable timeset %s: %s", timesetName, getErr))
			}
			if timesets.Entities == nil || len(*timesets.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no callable timeset found with timesetName %s", timesetName))
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
