package aistudio_summary_setting

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_aistudio_summary_setting.go contains the data source implementation
   for the resource.
*/

// dataSourceAistudioSummarySettingRead retrieves by name the id in question
func dataSourceAistudioSummarySettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newAistudioSummarySettingProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		summarySettingId, resp, retryable, err := proxy.getAistudioSummarySettingIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching aistudio summary setting %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No aistudio summary setting found with name %s", name), resp))
		}

		d.SetId(summarySettingId)
		return nil
	})
}
