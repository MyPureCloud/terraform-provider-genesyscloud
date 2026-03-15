package process_automation_trigger

import (
	"context"
	"fmt"
	"time"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

type ProcessAutomationTriggers struct {
	Entities *[]ProcessAutomationTrigger `json:"entities,omitempty"`
	NextUri  *string                     `json:"nextUri,omitempty"`
}

func dataSourceProcessAutomationTrigger() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud process automation trigger. Select a trigger by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceProcessAutomationTriggerRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the trigger",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceProcessAutomationTriggerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig

	triggerName := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		relativePath := "/api/v2/processAutomation/triggers"

		for {
			processAutomationTriggers, resp, getErr := getAllProcessAutomationTriggers(ctx, sdkConfig, relativePath)

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get page of process automation triggers: %s", getErr), resp))
			}

			if processAutomationTriggers.Entities == nil || len(*processAutomationTriggers.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no process automation triggers found with name: %s", triggerName), resp))
			}

			for _, trigger := range *processAutomationTriggers.Entities {
				if trigger.Name != nil && *trigger.Name == triggerName {
					d.SetId(*trigger.Id)
					return nil
				}
			}

			if processAutomationTriggers.NextUri == nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no process automation triggers found with name: %s", getErr), resp))
			}

			relativePath = *processAutomationTriggers.NextUri
		}
	})
}

func getAllProcessAutomationTriggers(ctx context.Context, config *platformclientv2.Configuration, relativePath string) (*ProcessAutomationTriggers, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(config, ResourceType)
	return customapi.Do[ProcessAutomationTriggers](ctx, c, customapi.MethodGet, relativePath, nil, nil)
}
