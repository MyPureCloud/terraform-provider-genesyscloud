package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceWebDeploymentsConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Web Deployments Configurations. Select a configuration by name.",
		ReadContext: ReadWithPooledClient(dataSourceConfigurationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the configuration",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Description: "The version of the configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		configs, _, err := api.GetWebdeploymentsConfigurations(false)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error retrieving web deployment configuration %s: %s", name, err))
		}

		for _, config := range *configs.Entities {
			if name == *config.Name {
				d.SetId(*config.Id)
				version := determineLatestVersion(ctx, api, *config.Id)
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
