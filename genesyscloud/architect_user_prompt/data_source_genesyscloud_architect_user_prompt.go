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
)

func dataSourceUserPromptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	name := d.Get("name").(string)
	nameArr := []string{name}

	// Query user prompt by name. Retry in case search has not yet indexed the user prompt.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {

		prompts, _, getErr, retryable := proxy.getArchitectUserPrompts(ctx, true, true, nameArr)
		if retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting user prompts %s: %s", name, getErr))
		}

		if prompts == nil || len(*prompts) == 0 {
			return retry.RetryableError(fmt.Errorf("no user prompts found with name %s", name))
		}

		prompt := (*prompts)[0]
		d.SetId(*prompt.Id)

		return nil
	})
}
