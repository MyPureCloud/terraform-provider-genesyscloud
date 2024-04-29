package employeeperformance_externalmetrics_definitions

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
   The data_source_genesyscloud_employeeperformance_externalmetrics_definition.go contains the data source implementation
   for the resource.
*/

// dataSourceEmployeeperformanceExternalmetricsDefinitionRead retrieves by name the id in question
func dataSourceEmployeeperformanceExternalmetricsDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		domainOrganizationRoleId, retryable, resp, err := proxy.getEmployeeperformanceExternalmetricsDefinitionIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error searching employeeperformance externalmetrics definition %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No employeeperformance externalmetrics definition found with name %s", name), resp))
		}

		d.SetId(domainOrganizationRoleId)
		return nil
	})
}
