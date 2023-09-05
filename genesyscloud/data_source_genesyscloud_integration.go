package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func dataSourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration. Select an integration by name",
		ReadContext: ReadWithPooledClient(dataSourceIntegrationRead),
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

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			integrations, _, getErr := integrationAPI.GetIntegrations(pageSize, pageNum, "", nil, "", "")

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to get page of integrations: %s", getErr))
			}

			if integrations.Entities == nil || len(*integrations.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no integrations found with name: %s", integrationName))
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
