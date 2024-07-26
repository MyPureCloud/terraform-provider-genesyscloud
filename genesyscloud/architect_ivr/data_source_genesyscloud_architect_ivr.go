package architect_ivr

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	util "terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceIvrRead retrieves the Genesys Cloud architect ivr id by name
func dataSourceIvrRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)
	name := d.Get("name").(string)

	// Query ivr by name. Retry in case search has not yet indexed the ivr.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		id, retryable, resp, err := ap.getArchitectIvrIdByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting IVR %s | error: %s", name, err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting IVR %s | error: %s", name, err), resp))
		}
		d.SetId(id)
		return nil
	})
}
