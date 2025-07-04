package guide

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

func dataSourceGuideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideProxy(sdkConfig)

	nameStr := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		guideId, retryable, resp, err := proxy.getGuideByName(ctx, nameStr)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting guide %s | error: %s", nameStr, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting guide %s | error: %s", nameStr, err), resp))
		}

		d.SetId(guideId)
		return nil
	})
}
