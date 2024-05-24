package integration_action

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
   The data_source_genesyscloud_integration_action.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceIntegrationActionRead retrieves by name the integration action id in question
func dataSourceIntegrationActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	actionName := d.Get("name").(string)

	// Query for integration actions by name. Retry in case new action is not yet indexed by search.
	// As action names are non-unique, fail in case of multiple results.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		actions, resp, err := iap.getIntegrationActionsByName(ctx, actionName)

		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting data action %s | error: %s", actionName, err), resp))
		}

		if len(*actions) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no data actions found with name %s", actionName), resp))
		}

		if len(*actions) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("ambiguous data action name: %s", actionName), resp))
		}
		action := (*actions)[0]
		d.SetId(*action.Id)
		return nil
	})
}
