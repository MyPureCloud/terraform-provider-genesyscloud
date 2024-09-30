package auth_division

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

func dataSourceAuthDivisionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthDivisionProxy(sdkConfig)
	name := d.Get("name").(string)

	// Query division by name. Retry in case search has not yet indexed the division.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		divisionId, resp, retryable, getErr := proxy.getAuthDivisionIdByName(ctx, name)
		if getErr != nil {
			errorDetails := util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting division %s | error: %s", name, getErr), resp)
			if !retryable {
				return retry.NonRetryableError(errorDetails)
			}
			return retry.RetryableError(errorDetails)
		}

		d.SetId(divisionId)
		return nil
	})
}
