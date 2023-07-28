package webdeployments_configuration

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

func dataSourceConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		configs, _, err := api.GetWebdeploymentsConfigurations(false)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error retrieving web deployment configuration %s: %s", name, err))
		}

		for _, config := range *configs.Entities {
			if name == *config.Name {
				d.SetId(*config.Id)
				version := determineLatestVersion(ctx, api, *config.Id)
				if version == "draft" {
					return resource.NonRetryableError(fmt.Errorf("Web deployment configuration %s has no published versions and so cannot be used", name))
				}

				d.Set("version", version)

				return nil
			}
		}

		return resource.NonRetryableError(fmt.Errorf("No web deployment configuration was found with the name %s", name))
	})
}
