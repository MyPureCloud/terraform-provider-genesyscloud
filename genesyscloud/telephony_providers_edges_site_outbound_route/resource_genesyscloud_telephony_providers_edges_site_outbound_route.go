package telephony_providers_edges_site_outbound_route

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
)

func getSitesOutboundRoutes(ctx context.Context, configuration *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return nil, nil
}

func createSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	siteId := d.Get("site_id").(string)

	return nil
}

func readSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	siteId := d.Get("site_id").(string)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSiteOutboundRoute(), constants.DefaultConsistencyChecks, resourceName)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		return cc.CheckState(d)
	})
}

func updateSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	siteId := d.Get("site_id").(string)

	return nil
}

func deleteSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	siteId := d.Get("site_id").(string)

	return nil
}
