package outbound_callabletimeset

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The data_source_genesyscloud_outbound_callabletimeset.go contains the data source implementation for the resource.
*/

// dataSourceOutboundCallabletimesetRead retrieves by name term the id in question
func dataSourceOutboundCallabletimesetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallabletimesetProxy(sdkConfig)
	timesetName := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		timesetId, retryable, resp, err := proxy.getOutboundCallabletimesetByName(ctx, timesetName)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting callable timeset %s | error: %s", timesetName, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no callable timeset found with timesetName %s", timesetName), resp))
		}
		d.SetId(timesetId)
		return nil
	})
}
