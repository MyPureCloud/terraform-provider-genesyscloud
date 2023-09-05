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

func dataSourceIntegrationAction() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration action. Select an integration action by name",
		ReadContext: ReadWithPooledClient(dataSourceIntegrationActionRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration action",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceIntegrationActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	actionName := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			integrationAction, _, getErr := integrationAPI.GetIntegrationsActions(pageSize, pageNum, "", "", "", "", "", actionName, "", "", "")

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to get page of integration actions: %s", getErr))
			}

			if integrationAction.Entities == nil || len(*integrationAction.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no integration actions found with name: %s", actionName))
			}

			for _, action := range *integrationAction.Entities {
				if action.Name != nil && *action.Name == actionName {
					d.SetId(*action.Id)
					return nil
				}
			}

		}
	})
}
