package routing_language

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

func dataSourceRoutingLanguageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingLanguageProxy(sdkConfig)
	name := d.Get("name").(string)

	// Find first non-deleted language by name. Retry in case new language is not yet indexed by search
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		languageId, resp, retryable, err := proxy.getRoutingLanguageIdByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting language %s | error: %s", name, err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting language %s | error: %s", name, err), resp))
		}

		d.SetId(languageId)
		return nil
	})
}
