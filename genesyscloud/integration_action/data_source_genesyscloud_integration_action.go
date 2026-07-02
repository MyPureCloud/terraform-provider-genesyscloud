package integration_action

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

/*
   The data_source_genesyscloud_integration_action.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceIntegrationActionRead retrieves by name the integration action id in question.
// If integration_id is supplied (for example when looking up a static data action whose name
// may collide across integration instances), results are further filtered by integration id.
// If action_type is supplied ("static" or "custom"), results are further filtered by whether
// the action ID starts with the "static" prefix.
func dataSourceIntegrationActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	actionName := d.Get("name").(string)
	integrationId := d.Get("integration_id").(string)
	actionType := d.Get("action_type").(string)

	// Query for integration actions by name. Retry in case new action is not yet indexed by search.
	// As action names are non-unique, fail in case of multiple results unless integration_id
	// and/or action_type is supplied to disambiguate.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		actions, resp, err := iap.getIntegrationActionsByName(ctx, actionName)

		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting data action %s | error: %s", actionName, err), resp))
		}

		if len(*actions) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no data actions found with name %s", actionName), resp))
		}

		matches := *actions

		// Filter by integration_id if provided
		if integrationId != "" {
			filtered := make([]platformclientv2.Action, 0, len(matches))
			for _, action := range matches {
				if action.IntegrationId != nil && *action.IntegrationId == integrationId {
					filtered = append(filtered, action)
				}
			}
			if len(filtered) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no data actions found with name %s and integration_id %s", actionName, integrationId), resp))
			}
			matches = filtered
		}

		// Filter by action_type if provided ("static" or "custom")
		if actionType != "" {
			filtered := make([]platformclientv2.Action, 0, len(matches))
			for _, action := range matches {
				if action.Id == nil {
					continue
				}
				isStatic := strings.HasPrefix(*action.Id, staticActionIDPrefix)
				if (actionType == "static" && isStatic) || (actionType == "custom" && !isStatic) {
					filtered = append(filtered, action)
				}
			}
			if len(filtered) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no data actions found with name %s and action_type %s", actionName, actionType), resp))
			}
			matches = filtered
		}

		if len(matches) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("ambiguous data action name: %s (set integration_id and/or action_type to disambiguate)", actionName), resp))
		}
		action := matches[0]
		d.SetId(*action.Id)
		return nil
	})
}
