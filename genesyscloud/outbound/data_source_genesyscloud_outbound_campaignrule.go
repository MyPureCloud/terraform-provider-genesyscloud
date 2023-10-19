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

func dataSourceOutboundCampaignRule() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Campaign Rules. Select a campaign rule by name.",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceCampaignRuleRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Campaign Rule name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceCampaignRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query campaign rule by name. Retry in case search has not yet indexed the campaign rule.
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		campaignRules, _, getErr := outboundAPI.GetOutboundCampaignrules(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting campaign rule %s: %s", name, getErr))
		}

		if campaignRules.Entities == nil || len(*campaignRules.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("no campaign rules found with name %s", name))
		}

		campaignRule := (*campaignRules.Entities)[0]
		d.SetId(*campaignRule.Id)
		return nil
	})
}
