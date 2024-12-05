package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
	"log"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllSitesAndOutboundRoutes(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		log.Printf("cannot export %s because environment variable %s is not set", ResourceType, featureToggles.OutboundRoutesToggleName())
		return nil, nil
	}
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	var allSites []platformclientv2.Site

	// get unmanaged sites
	unmanagedSites, resp, err := proxy.siteProxy.GetAllSites(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get unmanaged sites error: %s", err), resp)
	}
	allSites = append(allSites, *unmanagedSites...)

	// get managed sites
	managedSites, resp, err := proxy.siteProxy.GetAllSites(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get managed sites error: %s", err), resp)
	}
	allSites = append(allSites, *managedSites...)

	for _, site := range allSites {
		routes, resp, err := proxy.getAllSiteOutboundRoutes(ctx, *site.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to check site %s outbound routes: %s", *site.Id, err), resp)
		}
		if routes != nil && len(*routes) > 0 {
			for _, route := range *routes {
				outboundRouteId := buildSiteAndOutboundRouteId(*site.Id, *route.Id)
				resources[outboundRouteId] = &resourceExporter.ResourceMeta{BlockLabel: *route.Name}
			}
		}
	}

	return resources, nil
}

func createSiteOutboundRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)

	siteId := d.Get("site_id").(string)

	outboundRoute := buildOutboundRoutes(d)

	newOutboundRoute, resp, err := proxy.createSiteOutboundRoute(ctx, siteId, outboundRoute)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create outbound route %s for site %s: %s", *outboundRoute.Name, siteId, err), resp)
	}

	outboundRouteId := buildSiteAndOutboundRouteId(siteId, *newOutboundRoute.Id)
	d.Set("route_id", *newOutboundRoute.Id)
	d.SetId(outboundRouteId)
	log.Printf("created outbound route %s for site %s", *newOutboundRoute.Id, siteId)
	return readSiteOutboundRoute(ctx, d, meta)
}

func readSiteOutboundRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSiteOutboundRoute(), constants.ConsistencyChecks(), ResourceType)

	siteId, outboundRouteId := splitSiteAndOutboundRoute(d.Id())

	log.Printf("Reading outbound route %s for site %s", outboundRouteId, siteId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		outboundRoute, resp, err := proxy.getSiteOutboundRouteById(ctx, siteId, outboundRouteId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read outbound route %s for site %s | error: %s", d.Id(), siteId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read outbound route %s for site %s | error: %s", d.Id(), siteId, err), resp))
		}

		_ = d.Set("site_id", siteId)
		_ = d.Set("route_id", outboundRouteId)

		if outboundRoute != nil {
			_ = d.Set("name", *outboundRoute.Name)
			resourcedata.SetNillableValue(d, "description", outboundRoute.Description)
			resourcedata.SetNillableValue(d, "enabled", outboundRoute.Enabled)
			resourcedata.SetNillableValue(d, "distribution", outboundRoute.Distribution)

			if outboundRoute.ClassificationTypes != nil {
				d.Set("classification_types", lists.StringListToInterfaceList(*outboundRoute.ClassificationTypes))
			}

			if len(*outboundRoute.ExternalTrunkBases) > 0 {
				externalTrunkBaseIds := make([]string, 0)
				for _, externalTrunkBase := range *outboundRoute.ExternalTrunkBases {
					externalTrunkBaseIds = append(externalTrunkBaseIds, *externalTrunkBase.Id)
				}
				d.Set("external_trunk_base_ids", lists.StringListToInterfaceList(externalTrunkBaseIds))
			}
		}

		log.Printf("Read outbound route %s for site %s", outboundRouteId, siteId)
		return cc.CheckState(d)
	})
}

func updateSiteOutboundRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)

	siteId, outboundRouteId := splitSiteAndOutboundRoute(d.Id())
	outboundRoute := buildOutboundRoutes(d)

	_, resp, err := proxy.updateSiteOutboundRoute(ctx, siteId, outboundRouteId, outboundRoute)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update outbound route with id %s for site %s error: %s", outboundRoute, siteId, err), resp)
	}
	// Wait for the update before reading
	time.Sleep(5 * time.Second)

	return readSiteOutboundRoute(ctx, d, meta)
}

func deleteSiteOutboundRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)

	siteId, outboundRouteId := splitSiteAndOutboundRoute(d.Id())

	// Verify parent site still exists before trying to delete outbound routes
	_, resp, err := proxy.getSite(ctx, siteId)
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Parent site %s already deleted", d.Id())
			return nil
		}

		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete outbound route %s for site %s due to error: %s", outboundRouteId, siteId, err), resp)
	}

	log.Printf("Deleting outbound route %s for site %s", outboundRouteId, siteId)

	resp, err = proxy.deleteSiteOutboundRoute(ctx, siteId, outboundRouteId)
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Outbound route %s already deleted for site %s", outboundRouteId, siteId)
			return nil
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		outboundRoute, resp, err := proxy.getSiteOutboundRouteById(ctx, siteId, outboundRouteId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted outbound route %s for site %s", outboundRouteId, siteId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to verify delete of outbound routes for site %s error: %s", d.Id(), err), resp))
		}

		if outboundRoute == nil {
			log.Printf("Deleted outbound route %s for site %s", outboundRouteId, siteId)
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("outbound route %s for site %s still exists", outboundRouteId, siteId), resp))
	})
}
