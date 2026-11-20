package analytics_dataretention_settings

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_analytics_dataretention_settings.go contains the data source implementation for the resource.

   Note: This code should contain no code for doing the actual lookup in Genesys Cloud. Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

func dataSourceAnalyticsDataretentionSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getAnalyticsDataretentionSettingsProxy(sdkConfig)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		settings, resp, err := proxy.getAnalyticsDataretentionSettings(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting analytics data retention settings | error: %s", err), resp))
		}

		// The organization analytics data retention setting is a singleton; assign the fixed ID.
		d.SetId(analyticsDataretentionSettingsId)
		setAnalyticsDataretentionSettingsToResourceData(d, settings)
		return nil
	})
}
