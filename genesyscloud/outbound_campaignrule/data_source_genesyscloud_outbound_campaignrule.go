package outbound_campaignrule

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

func dataSourceOutboundCampaignruleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query campaign rule by name. Retry in case search has not yet indexed the campaign rule.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		campaignRules, resp, getErr := outboundAPI.GetOutboundCampaignrules(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting campaign rule %s | error: %s", name, getErr), resp))
		}

		if campaignRules.Entities == nil || len(*campaignRules.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no campaign rules found with name %s", name), resp))
		}

		campaignRule := (*campaignRules.Entities)[0]
		d.SetId(*campaignRule.Id)
		return nil
	})
}
