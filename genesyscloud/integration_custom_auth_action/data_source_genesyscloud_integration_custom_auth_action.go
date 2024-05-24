package integration_custom_auth_action

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_integration_custom_auth_action.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceIntegrationCustomAuthActionRead retrieves the custom auth action id from the integration name
func dataSourceIntegrationCustomAuthActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	integrationId := d.Get("parent_integration_id").(string)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		integration, resp, getErr := cap.getIntegrationById(ctx, integrationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read integration %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read integration %s | error: %s", d.Id(), getErr), resp))
		}

		// Get the custom auth action for the integration
		authActionId := getCustomAuthIdFromIntegration(*integration.Id)
		authAction, resp, err := cap.getCustomAuthActionById(ctx, authActionId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("cannot find custom auth action of integration %s | error: %v", *integration.Name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error deleting integration %s | error: %s", d.Id(), err), resp))
		}
		d.SetId(*authAction.Id)
		return nil
	})
}
