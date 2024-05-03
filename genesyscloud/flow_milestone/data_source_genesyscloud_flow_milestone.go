package flow_milestone

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
   The data_source_genesyscloud_flow_milestone.go contains the data source implementation
   for the resource.
*/

// dataSourceFlowMilestoneRead retrieves by name the id in question
func dataSourceFlowMilestoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newFlowMilestoneProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		flowMilestoneId, retryable, resp, err := proxy.getFlowMilestoneIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error searching flow milestone %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No flow milestone found with name %s", name), resp))
		}

		d.SetId(flowMilestoneId)
		return nil
	})
}
