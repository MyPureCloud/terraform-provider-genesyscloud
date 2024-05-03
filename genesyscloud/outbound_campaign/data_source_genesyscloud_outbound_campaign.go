package outbound_campaign

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

/*
   The data_source_genesyscloud_outbound_campaign.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundCampaignRead retrieves by name the id in question
func dataSourceOutboundCampaignRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundCampaignProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		campaignId, retryable, resp, err := proxy.getOutboundCampaignIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error campaign %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No campaign found with name %s", name), resp))
		}

		d.SetId(campaignId)
		return nil
	})
}
