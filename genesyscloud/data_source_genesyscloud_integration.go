package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"time"
)

func dataSourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration. Select an integration by name",
		ReadContext: readWithPooledClient(dataSourceIntegrationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "The name of the integration",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
		},
	}
}

func dataSourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	integrationName := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			integrations, _, getErr := integrationAPI.GetIntegrations(100, pageNum, "", make([]string, 0), "", "")

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting list of integrations: %s", getErr))
			}

			if integrations.Entities == nil || len(*integrations.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no integrations found"))
			}

			for _, integration := range *integrations.Entities {
				if  integration.Name != nil && *integration.Name ==  integrationName {
					d.SetId(*integration.Id)
					return nil
				}
			}

		}
	})

}
