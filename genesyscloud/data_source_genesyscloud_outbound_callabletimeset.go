package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v75/platformclientv2"
	"time"
)

func dataSourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Clound Outbound Callable Timesets. Select a callable timeset by name.",
		ReadContext: readWithPooledClient(dataSourceOutboundCallabletimesetRead),
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
	sdkConfig := m.(*providerMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100

		callableTimesets, _, getErr := outboundAPI.GetOutboundCallabletimesets(pageSize, pageNum, false, "", "", "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("error requesting callable timeset %s: %s", name, getErr))
		}
		if callableTimesets.Entities == nil || len(*callableTimesets.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("no callable timeset found with name %s", name))
		}
		callableTimeset := (*callableTimesets.Entities)[0]
		d.SetId(*callableTimeset.Id)
		return nil
	})
}
