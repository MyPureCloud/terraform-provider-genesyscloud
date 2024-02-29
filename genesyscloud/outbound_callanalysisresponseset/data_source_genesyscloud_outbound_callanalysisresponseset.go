package outbound_callanalysisresponseset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_outbound_callanalysisresponseset.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundCallanalysisresponsesetRead retrieves by name the id in question
func dataSourceOutboundCallanalysisresponsesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundCallanalysisresponsesetProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		responseSetId, retryable, err := proxy.getOutboundCallanalysisresponsesetIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching outbound callanalysisresponseset %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No outbound callanalysisresponseset found with name %s", name))
		}

		d.SetId(responseSetId)
		return nil
	})
}
