package analytics_dataretention_settings

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

/*
The resource_genesyscloud_analytics_dataretention_settings.go contains all the methods that perform the core logic for the resource.
*/

// analyticsDataretentionSettingsId is the fixed ID assigned to this singleton resource
const analyticsDataretentionSettingsId = "analytics_dataretention_settings"

// getAllAnalyticsDataretentionSettings fetches the organization analytics data retention setting for the exporter.
func getAllAnalyticsDataretentionSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource has only a single instance, we attempt to fetch the data from
	// the API in order to verify the user's permission to access this resource's API endpoint.
	proxy := getAnalyticsDataretentionSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getAnalyticsDataretentionSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// The setting has never been configured for the org, so there is nothing to export
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s resource due to error: %s", ResourceType, err), resp)
	}

	resources[analyticsDataretentionSettingsId] = &resourceExporter.ResourceMeta{BlockLabel: ResourceType}
	return resources, nil
}

// createAnalyticsDataretentionSettings is used by the analytics_dataretention_settings resource to set the Genesys Cloud analytics data retention setting.
// Since the API has no create operation, we assign a fixed ID and delegate to update.
func createAnalyticsDataretentionSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(analyticsDataretentionSettingsId)
	return updateAnalyticsDataretentionSettings(ctx, d, meta)
}

// readAnalyticsDataretentionSettings is used by the analytics_dataretention_settings resource to read the analytics data retention setting from Genesys Cloud
func readAnalyticsDataretentionSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAnalyticsDataretentionSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAnalyticsDataretentionSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading analytics data retention settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getAnalyticsDataretentionSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read analytics data retention settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read analytics data retention settings %s | error: %s", d.Id(), getErr), resp))
		}

		setAnalyticsDataretentionSettingsToResourceData(d, settings)

		log.Printf("Read analytics data retention settings %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateAnalyticsDataretentionSettings is used by the analytics_dataretention_settings resource to update the analytics data retention setting in Genesys Cloud
func updateAnalyticsDataretentionSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAnalyticsDataretentionSettingsProxy(sdkConfig)
	request := getAnalyticsDataretentionSettingsFromResourceData(d)

	log.Printf("Updating analytics data retention settings %s", d.Id())

	// The API returns a 409 when existing recording media retention policies, recordings or the snippet
	// recording settings exceed the requested retention_days; the API error message is surfaced as-is.
	_, resp, err := proxy.updateAnalyticsDataretentionSettings(ctx, &request)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update analytics data retention settings: %s", err), resp)
	}

	log.Printf("Updated analytics data retention settings %s", d.Id())
	return readAnalyticsDataretentionSettings(ctx, d, meta)
}

// deleteAnalyticsDataretentionSettings is a no-op for the analytics_dataretention_settings resource. The setting cannot
// be cleared once it has been configured (the API has no delete operation); Terraform simply removes it from state.
func deleteAnalyticsDataretentionSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Deleting analytics data retention settings %s is a no-op; removing from state only", d.Id())
	return nil
}
