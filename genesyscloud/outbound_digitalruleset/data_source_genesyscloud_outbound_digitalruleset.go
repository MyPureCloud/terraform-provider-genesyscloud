package outbound_digitalruleset

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
   The data_source_genesyscloud_outbound_digitalruleset.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundDigitalrulesetRead retrieves by name the id in question
func dataSourceOutboundDigitalrulesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundDigitalrulesetProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		digitalRuleSetId, retryable, err := proxy.getOutboundDigitalrulesetIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching outbound digitalruleset %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No outbound digitalruleset found with name %s", name))
		}

		d.SetId(digitalRuleSetId)
		return nil
	})
}
