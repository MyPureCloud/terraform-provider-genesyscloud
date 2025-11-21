package workforcemanagement_businessunits

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
   The data_source_genesyscloud_workforcemanagement_businessunits.go contains the data source implementation
   for the resource.
*/

// dataSourceWorkforcemanagementBusinessunitsRead retrieves by name the id in question
func dataSourceWorkforcemanagementBusinessunitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newWorkforceManagementBusinessUnitsProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		businessUnitResponseId, resp, retryable, err := proxy.getWorkforceManagementBusinessUnitIdByExactName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching workforce management business unit %s | error: %s", name, err), resp))
		}

		if retryable {
			return util.RetryableErrorWithRetryAfter(ctx, util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No workforce management business unit found with name %s", name), resp), resp)
		}

		d.SetId(businessUnitResponseId)
		return nil
	})
}
