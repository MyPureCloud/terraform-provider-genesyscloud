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

func dataSourceUserPrompt() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud User Prompts. Select a user prompt by name.",
		ReadContext: ReadWithPooledClient(dataSourceUserPromptRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "User Prompt name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceUserPromptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	nameArr := []string{name}

	// Query user prompt by name. Retry in case search has not yet indexed the user prompt.
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		prompts, _, getErr := architectApi.GetArchitectPrompts(pageNum, pageSize, nameArr, "", "", "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting user prompts %s: %s", name, getErr))
		}

		if prompts.Entities == nil || len(*prompts.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No user prompts found with name %s", name))
		}

		prompt := (*prompts.Entities)[0]
		d.SetId(*prompt.Id)

		return nil
	})
}
