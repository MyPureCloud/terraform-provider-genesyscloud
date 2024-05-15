package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
)

func getSitesOutboundRoutes(ctx context.Context, configuration *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return nil, nil
}

func createSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}

	siteId := d.Get("site_id").(string)
	log.Printf("creating outbound routes for site %s", siteId)
	d.SetId(siteId)

	return updateSiteOutboundRoutes(ctx, d, meta)
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
