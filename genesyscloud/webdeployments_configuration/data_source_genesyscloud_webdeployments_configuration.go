package webdeployments_configuration

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		configs, resp, err := wp.getWebDeploymentsConfiguration(ctx)

		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error retrieving web deployment configuration %s | error: %s", name, err), resp))
		}

		for _, config := range *configs.Entities {
			if name == *config.Name {
				d.SetId(*config.Id)
				version := wp.determineLatestVersion(ctx, *config.Id)
				if version == "draft" {
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Web deployment configuration %s has no published versions and so cannot be used", name), resp))
				}

				_ = d.Set("version", version)

				return nil
			}
		}

		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No web deployment configuration was found with the name %s", name), resp))
	})
}
