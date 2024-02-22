package outbound_campaign

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	gcloud "terraform-provider-genesyscloud/genesyscloud/util"
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

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		campaignId, retryable, err := proxy.getOutboundCampaignIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error campaign %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No campaign found with name %s", name))
		}

		d.SetId(campaignId)
		return nil
	})
}
