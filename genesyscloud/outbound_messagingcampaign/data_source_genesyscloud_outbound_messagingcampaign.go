package outbound_messagingcampaign

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_outbound_messagingcampaign.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundMessagingcampaignRead retrieves by name the id in question
func dataSourceOutboundMessagingcampaignRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundMessagingcampaignProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		messagingCampaignId, retryable, resp, err := proxy.getOutboundMessagingcampaignIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching outbound messagingcampaign %s: %s", name, err), resp))
		}

		if retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No outbound messagingcampaign found with name %s", name), resp))

		}

		d.SetId(messagingCampaignId)
		return nil
	})
}
