package flow_outcome

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
   The data_source_genesyscloud_flow_outcome.go contains the data source implementation
   for the resource.
*/

// dataSourceFlowOutcomeRead retrieves by name the id in question
func dataSourceFlowOutcomeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newFlowOutcomeProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		flowOutcomeId, retryable, err := proxy.getFlowOutcomeIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching flow outcome %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No flow outcome found with name %s", name))
		}

		d.SetId(flowOutcomeId)
		return nil
	})
}
