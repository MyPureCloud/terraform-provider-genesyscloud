package integration

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
   The data_source_genesyscloud_integration.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceIntegrationRead retrieves by name the integration id in question
func dataSourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	integrationName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		integration, retryable, err := ip.getIntegrationByName(ctx, integrationName)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("failed to get page of integrations: %s. %s", integrationName, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("failed to get integration %s", integrationName))
		}

		d.SetId(*integration.Id)
		return nil
	})
}
