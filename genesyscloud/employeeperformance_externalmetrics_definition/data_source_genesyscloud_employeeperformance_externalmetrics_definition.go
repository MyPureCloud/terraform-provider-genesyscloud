package employeeperformance_externalmetrics_definition

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
   The data_source_genesyscloud_employeeperformance_externalmetrics_definition.go contains the data source implementation
   for the resource.
*/

// dataSourceEmployeeperformanceExternalmetricsDefinitionRead retrieves by name the id in question
func dataSourceEmployeeperformanceExternalmetricsDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		domainOrganizationRoleId, retryable, err := proxy.getEmployeeperformanceExternalmetricsDefinitionIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching employeeperformance externalmetrics definition %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No employeeperformance externalmetrics definition found with name %s", name))
		}

		d.SetId(domainOrganizationRoleId)
		return nil
	})
}
