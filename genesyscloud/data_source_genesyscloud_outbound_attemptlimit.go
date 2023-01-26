package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v91/platformclientv2"
)

func dataSourceOutboundAttemptLimit() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Attempt Limits. Select an attempt limit by name.",
		ReadContext: readWithPooledClient(dataSourceOutboundAttemptLimitRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Attempt Limit name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundAttemptLimitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		attemptLimits, _, getErr := outboundAPI.GetOutboundAttemptlimits(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("error requesting attempt limit %s: %s", name, getErr))
		}
		if attemptLimits.Entities == nil || len(*attemptLimits.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("no attempt limits found with name %s", name))
		}
		attemptLimit := (*attemptLimits.Entities)[0]
		d.SetId(*attemptLimit.Id)
		return nil
	})
}
