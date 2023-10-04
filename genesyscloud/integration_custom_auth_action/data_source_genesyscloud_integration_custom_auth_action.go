package integration_custom_auth_action

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
   The data_source_genesyscloud_integration_custom_auth_action.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceIntegrationCustomAuthActionRead retrieves the custom auth action id from the integration name
func dataSourceIntegrationCustomAuthActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	integrationName := d.Get("integration_name").(string)

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		// Get the integration by name
		integration, retryable, err := cap.getIntegrationByName(ctx, integrationName)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("failed to get page of integrations: %s. %s", integrationName, err))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("failed to get integration %s", integrationName))
		}

		// Get the custom auth action for the integration
		authActionId := getCustomAuthIdFromIntegration(*integration.Id)
		authAction, resp, err := cap.getCustomAuthActionById(ctx, authActionId)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("cannot find custom auth action of integration %s: %v", *integration.Name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting integration %s: %s", d.Id(), err))
		}

		d.SetId(*authAction.Id)
		return nil
	})
}
