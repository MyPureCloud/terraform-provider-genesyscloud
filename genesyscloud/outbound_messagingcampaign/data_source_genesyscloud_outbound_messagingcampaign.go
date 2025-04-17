package outbound_messagingcampaign

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_outbound_messagingcampaign.go contains the data source implementation
   for the resource.
*/

var obMessagingCampaignDataSourceCache *rc.DataSourceCache

// dataSourceOutboundMessagingcampaignRead retrieves by name the id in question
func dataSourceOutboundMessagingcampaignRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	key := d.Get("name").(string)

	if obMessagingCampaignDataSourceCache == nil {
		log.Printf("Instantiating the %s data source cache object", ResourceType)
		obMessagingCampaignDataSourceCache = rc.NewDataSourceCache(sdkConfig, hydrateObMessagingCampaignCacheFn, getObMessagingCampaignByNameFn)
	}

	campaignId, err := rc.RetrieveId(obMessagingCampaignDataSourceCache, ResourceType, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(campaignId)
	return nil
}

// hydrateObMessagingCampaignCacheFn for hydrating the cache with Genesys Cloud Outbound Messaging Campaigns using the SDK
func hydrateObMessagingCampaignCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	log.Printf("Hydrating %s data source cache object", ResourceType)
	proxy := getOutboundMessagingcampaignProxy(c.ClientConfig)

	campaigns, resp, err := proxy.getAllOutboundMessagingcampaign(ctx)
	if err != nil {
		return fmt.Errorf("failed to get outbound messagingcampaigns: %v | API Response: %s", err.Error(), resp)
	}

	if campaigns == nil || len(*campaigns) == 0 {
		log.Printf("No outbound messagingcampaigns returned. Cache will remain empty")
		return nil
	}

	for _, campaign := range *campaigns {
		c.Cache[*campaign.Name] = *campaign.Id
	}

	log.Printf("Cache hydration complete for %s data source cache object", ResourceType)

	return nil
}

// getObMessagingCampaignByNameFn returns the campaign Id (blank if not found) and diag
func getObMessagingCampaignByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	log.Printf("Retrieving outbound messagingcampaign by name %s", name)
	proxy := getOutboundMessagingcampaignProxy(c.ClientConfig)
	campaignId := ""
	// Find first non-deleted campaigns by name. Retry in case new campaigns is not yet indexed by search
	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		campaignID, retryable, resp, err := proxy.getOutboundMessagingcampaignIdByName(ctx, name)
		if err != nil {
			diagnosticErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting outbound messagingcampaign %s: %s", name, err), resp)
			if !retryable {
				return retry.NonRetryableError(diagnosticErr)
			}
			return retry.RetryableError(diagnosticErr)
		}

		campaignId = campaignID
		return nil
	})

	return campaignId, diag
}
