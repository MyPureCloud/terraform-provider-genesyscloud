package webdeployments_configuration

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

func dataSourceConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	wp := getWebDeploymentConfigurationsProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		configs, err := wp.getWebDeploymentsConfiguration(ctx)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error retrieving web deployment configuration %s: %s", name, err))
		}

		for _, config := range *configs.Entities {
			if name == *config.Name {
				d.SetId(*config.Id)
				version := wp.determineLatestVersion(ctx, *config.Id)
				if version == "draft" {
					return retry.NonRetryableError(fmt.Errorf("Web deployment configuration %s has no published versions and so cannot be used", name))
				}

				d.Set("version", version)

				return nil
			}
		}

		return retry.NonRetryableError(fmt.Errorf("No web deployment configuration was found with the name %s", name))
	})
}
