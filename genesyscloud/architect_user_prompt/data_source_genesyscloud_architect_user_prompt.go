package architect_user_prompt

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func dataSourceUserPromptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	nameArr := []string{name}

	// Query user prompt by name. Retry in case search has not yet indexed the user prompt.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		prompts, _, getErr := architectApi.GetArchitectPrompts(pageNum, pageSize, nameArr, "", "", "", "", true, true, nil)
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
