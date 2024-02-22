package outbound_sequence

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	gcloud "terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_outbound_sequence.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundSequenceRead retrieves by name the id in question
func dataSourceOutboundSequenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundSequenceProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		campaignSequenceId, retryable, err := proxy.getOutboundSequenceIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching outbound sequence %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No outbound sequence found with name %s", name))
		}

		d.SetId(campaignSequenceId)
		return nil
	})
}
