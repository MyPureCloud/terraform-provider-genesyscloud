package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func DataSourceRoutingWrapupcode() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Wrap-up Code. Select a wrap-up code by name",
		ReadContext: ReadWithPooledClient(dataSourceRoutingWrapupcodeRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Wrap-up code name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingWrapupcodeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			wrapCode, _, getErr := routingAPI.GetRoutingWrapupcodes(100, pageNum, "", "", name, []string{}, []string{})

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting wrap-up code %s: %s", name, getErr))
			}

			if wrapCode.Entities == nil || len(*wrapCode.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No wrap-up code found with name %s", name))
			}

			d.SetId(*(*wrapCode.Entities)[0].Id)
			return nil
		}
	})
}
