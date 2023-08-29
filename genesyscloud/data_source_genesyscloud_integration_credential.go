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

func dataSourceIntegrationCredential() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration credential. Select an integration credential by name",
		ReadContext: ReadWithPooledClient(dataSourceIntegrationCredentialRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration credential",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceIntegrationCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	credName := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			integrationCredentials, _, getErr := integrationAPI.GetIntegrationsCredentials(pageNum, pageSize)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to get page of integration credentials: %s", getErr))
			}

			if integrationCredentials.Entities == nil || len(*integrationCredentials.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no integration credentials found with name: %s", credName))
			}

			for _, credential := range *integrationCredentials.Entities {
				if credential.Name != nil && *credential.Name == credName {
					d.SetId(*credential.Id)
					return nil
				}
			}

		}
	})

}
