package outbound_callabletimeset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
   The data_source_genesyscloud_outbound_callabletimeset.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundCallabletimesetRead retrieves by name the id in question
func dataSourceOutboundCallabletimesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundCallabletimesetProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		callableTimeSetId, retryable, err := proxy.getOutboundCallabletimesetIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error Outbound Callabletimeset %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No Outbound Callabletimeset found with name %s", name))
		}

		d.SetId(callableTimeSetId)
		return nil
	})
}
