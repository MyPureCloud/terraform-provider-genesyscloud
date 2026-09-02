package speechandtextanalytics_settings

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
The data_source_genesyscloud_speechandtextanalytics_settings.go contains the data source implementation for the resource.
*/

func dataSourceSpeechAndTextAnalyticsSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSpeechAndTextAnalyticsSettingsProxy(sdkConfig)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		settings, resp, err := proxy.getSpeechAndTextAnalyticsSettings(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting speech and text analytics settings | error: %s", err), resp))
		}

		d.SetId(speechAndTextAnalyticsSettingsId)
		setSpeechAndTextAnalyticsSettingsToResourceData(d, settings)
		return nil
	})
}
