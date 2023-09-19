package integration_credential

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_integration_credential.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceIntegrationCredentialRead retrieves by name the integration action id in question
func dataSourceIntegrationCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	credName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		credential, retryable, err := ip.getIntegrationCredByName(ctx, credName)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("failed to get integration credential: %s. %s", credential, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no integration credential found: %s", credName))
		}

		d.SetId(*credential.Id)
		return nil
	})
}
