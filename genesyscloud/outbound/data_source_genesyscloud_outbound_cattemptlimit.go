package outbound

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func DataSourceOutboundAttemptLimit() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Attempt Limits. Select an attempt limit by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundAttemptLimitRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		attemptLimits, resp, getErr := outboundAPI.GetOutboundAttemptlimits(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting attempt limit %s | error: %s", name, getErr), resp))
		}
		if attemptLimits.Entities == nil || len(*attemptLimits.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no attempt limits found with name %s", name), resp))
		}
		attemptLimit := (*attemptLimits.Entities)[0]
		d.SetId(*attemptLimit.Id)
		return nil
	})
}
