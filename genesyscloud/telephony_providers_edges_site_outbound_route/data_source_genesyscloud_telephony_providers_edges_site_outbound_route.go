package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSiteOutboundRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	sp := getSiteOutboundRouteProxy(sdkConfig)

	name := d.Get("name").(string)
	name = strings.TrimSuffix(name, "-outbound-routes")

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		siteId, retryable, resp, err := sp.siteProxy.GetSiteIdByName(ctx, name)
		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to get site %s", name), resp))
			}

			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting site %s | error: %s", name, err), resp))
		}

		d.SetId(siteId)
		return nil
	})
}
