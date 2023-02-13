package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func dataSourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration. Select an integration by name",
		ReadContext: readWithPooledClient(dataSourceIntegrationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	integrationName := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			integrations, _, getErr := integrationAPI.GetIntegrations(pageSize, pageNum, "", nil, "", "")

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("failed to get page of integrations: %s", getErr))
			}

			if integrations.Entities == nil || len(*integrations.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no integrations found with name: %s", integrationName))
			}

			for _, integration := range *integrations.Entities {
				if integration.Name != nil && *integration.Name == integrationName {
					d.SetId(*integration.Id)
					return nil
				}
			}

		}
	})

}
