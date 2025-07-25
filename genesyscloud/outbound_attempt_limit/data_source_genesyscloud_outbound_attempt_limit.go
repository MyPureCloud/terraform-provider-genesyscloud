package outbound_attempt_limit

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
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
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting attempt limit %s | error: %s", name, getErr), resp))
		}
		if attemptLimits.Entities == nil || len(*attemptLimits.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no attempt limits found with name %s", name), resp))
		}
		attemptLimit := (*attemptLimits.Entities)[0]
		d.SetId(*attemptLimit.Id)
		return nil
	})
}

func GenerateOutboundAttemptLimitDataSource(dataSourceLabel string, attemptLimitName string, dependsOn string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, ResourceType, dataSourceLabel, attemptLimitName, dependsOn)
}
