package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
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
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)

	name := d.Get("name").(string)
	siteId := d.Get("site_id").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		siteId, routeId, retryable, resp, err := proxy.getSiteOutboundRouteByName(ctx, name, siteId)
		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get outbound route %s", name), resp))
			}

			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting outbound route %s | error: %s", name, err), resp))
		}

		outboundRouteId := buildSiteAndOutboundRouteId(siteId, routeId)

		d.SetId(outboundRouteId)
		d.Set("site_id", siteId)
		d.Set("route_id", routeId)
		return nil
	})
}
