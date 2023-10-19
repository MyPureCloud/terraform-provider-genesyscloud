package outbound_attempt_limit

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

func DataSourceOutboundAttemptLimit() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Attempt Limits. Select an attempt limit by name.",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceOutboundAttemptLimitRead),
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
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		attemptLimits, _, getErr := outboundAPI.GetOutboundAttemptlimits(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting attempt limit %s: %s", name, getErr))
		}
		if attemptLimits.Entities == nil || len(*attemptLimits.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("no attempt limits found with name %s", name))
		}
		attemptLimit := (*attemptLimits.Entities)[0]
		d.SetId(*attemptLimit.Id)
		return nil
	})
}

func GenerateOutboundAttemptLimitDataSource(id string, attemptLimitName string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_attempt_limit" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, attemptLimitName, dependsOn)
}
