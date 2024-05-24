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

	// Query user prompt by name. Retry in case search has not yet indexed the user prompt.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {

		promptId, resp, getErr, retryable := proxy.getArchitectUserPromptIdByName(ctx, name)
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting user prompt by name %s | error: %s", name, getErr), resp))
		}
		if getErr != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error making user prompt request: %s", getErr), resp))
		}

		d.SetId(promptId)

		return nil
	})
}
