package integration_action

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

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
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	actionName := d.Get("name").(string)

	// Query for integration actions by name. Retry in case new action is not yet indexed by search.
	// As action names are non-unique, fail in case of multiple results.
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		actions, err := iap.getIntegrationActionsByName(ctx, actionName)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting data action %s: %s", actionName, err))
		}

		if len(*actions) == 0 {
			return retry.RetryableError(fmt.Errorf("no data actions found with name %s", actionName))
		}

		if len(*actions) > 1 {
			return retry.NonRetryableError(fmt.Errorf("ambiguous data action name: %s", actionName))
		}

		action := (*actions)[0]
		d.SetId(*action.Id)
		return nil
	})
}
