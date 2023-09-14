package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v107/platformclientv2"
)

func dataSourceRoutingSmsAddress() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Routing Sms Address. Select a Routing Sms Address by name.`,

		ReadContext: ReadWithPooledClient(dataSourceRoutingSmsAddressRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Routing Sms Address name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingSmsAddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdksmsaddressentitylisting, _, getErr := routingApi.GetRoutingSmsAddresses(pageSize, pageNum)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting Routing Sms Address %s: %s", name, getErr))
			}

			if sdksmsaddressentitylisting.Entities == nil || len(*sdksmsaddressentitylisting.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No Routing Sms Address found with name %s", name))
			}

			for _, entity := range *sdksmsaddressentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
