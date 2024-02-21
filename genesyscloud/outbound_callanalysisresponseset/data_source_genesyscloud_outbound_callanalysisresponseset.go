package outbound_callanalysisresponseset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
   The data_source_genesyscloud_outbound_callanalysisresponseset.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundCallanalysisresponsesetRead retrieves by name the id in question
func dataSourceOutboundCallanalysisresponsesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundCallanalysisresponsesetProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
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
