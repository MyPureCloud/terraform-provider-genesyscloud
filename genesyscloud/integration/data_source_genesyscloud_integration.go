package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

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

// dataSourceIntegrationRead retrieves by name the integration id in question.
// If integration_type is supplied, results are further filtered by integration type.
func dataSourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)
	integrationName := d.Get("name").(string)
	integrationType := d.Get("integration_type").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		var integration *platformclientv2.Integration
		var retryable bool
		var resp *platformclientv2.APIResponse
		var err error

		if integrationType != "" {
			integration, retryable, resp, err = getIntegrationByNameAndTypeFn(ctx, ip, integrationName, integrationType)
		} else {
			integration, retryable, resp, err = ip.getIntegrationByName(ctx, integrationName)
		}

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get page of integrations: %s | error: %s", integrationName, err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get integration %s", integrationName), resp))
		}
		d.SetId(*integration.Id)
		return nil
	})
}
