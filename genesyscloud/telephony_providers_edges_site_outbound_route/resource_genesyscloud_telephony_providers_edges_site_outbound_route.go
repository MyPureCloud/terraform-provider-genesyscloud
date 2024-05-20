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
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

func getSitesOutboundRoutes(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return nil, util.BuildDiagnosticError(resourceName, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getSiteOutboundRouteProxy(sdkConfig)

	// get unmanaged sites
	unmanagedSites, resp, err := proxy.getAllSites(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get unmanaged sites error: %s", err), resp)
	}
	for _, unmanagedSite := range *unmanagedSites {
		resources[*unmanagedSite.Id] = &resourceExporter.ResourceMeta{Name: *unmanagedSite.Name}
	}

	// get managed sites
	managedSites, resp, err := proxy.getAllSites(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get managed sites error: %s", err), resp)
	}
	for _, managedSite := range *managedSites {
		resources[*managedSite.Id] = &resourceExporter.ResourceMeta{Name: *managedSite.Name}
	}

	return resources, nil
}

func createSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	siteId := d.Get("site_id").(string)
	log.Printf("creating outbound routes for site %s", siteId)

	outboundRoutes := buildOutboundRoutes(d.Get("outbound_routes").(*schema.Set))
	for _, outboundRoute := range *outboundRoutes {
		_, resp, err := proxy.createSiteOutboundRoute(ctx, siteId, &outboundRoute)
		if err != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to create outbound route %s for site %s: %s", *outboundRoute.Name, siteId, err), resp)
		}
	}

	d.SetId(siteId)
	log.Printf("created outbound routes for site %s", d.Id())
	return readSiteOutboundRoutes(ctx, d, meta)
}

func readSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSiteOutboundRoute(), constants.DefaultConsistencyChecks, resourceName)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		outboundRoutes, resp, err := proxy.getSiteOutboundRoutes(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read outbound routes for site %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read outbound routes for site %s | error: %s", d.Id(), err), resp))
		}

		_ = d.Set("site_id", d.Id())

		if outboundRoutes != nil && len(*outboundRoutes) > 0 {
			outboundRoutesSet := schema.NewSet(schema.HashResource(outboundRouteSchema), []interface{}{})
			for _, outboundRoute := range *outboundRoutes {
				dOutboundRoute := make(map[string]interface{})
				dOutboundRoute["name"] = *outboundRoute.Name

				resourcedata.SetMapValueIfNotNil(dOutboundRoute, "description", outboundRoute.Description)
				resourcedata.SetMapValueIfNotNil(dOutboundRoute, "enabled", outboundRoute.Enabled)
				resourcedata.SetMapValueIfNotNil(dOutboundRoute, "distribution", outboundRoute.Distribution)

				if outboundRoute.ClassificationTypes != nil {
					dOutboundRoute["classification_types"] = lists.StringListToInterfaceList(*outboundRoute.ClassificationTypes)
				}

				if len(*outboundRoute.ExternalTrunkBases) > 0 {
					externalTrunkBaseIds := make([]string, 0)
					for _, externalTrunkBase := range *outboundRoute.ExternalTrunkBases {
						externalTrunkBaseIds = append(externalTrunkBaseIds, *externalTrunkBase.Id)
					}
					dOutboundRoute["external_trunk_base_ids"] = lists.StringListToInterfaceList(externalTrunkBaseIds)
				}

				outboundRoutesSet.Add(dOutboundRoute)
			}

			_ = d.Set("outbound_routes", outboundRoutesSet)
		} else {
			_ = d.Set("outbound_routes", nil)
		}

		return cc.CheckState(d)
	})
}

func updateSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OutboundRoutesToggleExists(); !exists {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Environment variable %s not set", featureToggles.OutboundRoutesToggleName()), fmt.Errorf("environment variable %s not set", featureToggles.OutboundRoutesToggleName()))
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)

	outboundRoutes := buildOutboundRoutes(d.Get("outbound_routes").(*schema.Set))

	// Get the current outbound routes
	outboundRoutesAPI, resp, err := proxy.getSiteOutboundRoutes(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get outbound routes for site %s error: %s", d.Id(), err), resp)
	}

	// Delete unwanted outbound routes first to free up classifications assigned to them
	for _, outboundRouteFromAPI := range *outboundRoutesAPI {
		// Delete route if no reference to it
		if _, ok := nameInOutboundRoutes(*outboundRouteFromAPI.Name, *outboundRoutes); !ok {
			resp, err := proxy.deleteSiteOutboundRoute(ctx, d.Id(), *outboundRouteFromAPI.Id)
			if err != nil {
				if util.IsStatus404(resp) {
					return nil
				}
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete outbound route from site %s error: %s", d.Id(), err), resp)
			}
		}
	}
	time.Sleep(2 * time.Second)

	// Update the outbound routes
	for _, outboundRoute := range *outboundRoutes {
		if outboundRouteUpdate, ok := nameInOutboundRoutes(*outboundRoute.Name, *outboundRoutesAPI); ok {
			// Update the outbound route
			outboundRouteUpdate.Name = outboundRoute.Name
			outboundRouteUpdate.Description = outboundRoute.Description
			outboundRouteUpdate.ClassificationTypes = outboundRoute.ClassificationTypes
			outboundRouteUpdate.Enabled = outboundRoute.Enabled
			outboundRouteUpdate.Distribution = outboundRoute.Distribution
			outboundRouteUpdate.ExternalTrunkBases = outboundRoute.ExternalTrunkBases

			_, resp, err := proxy.updateSiteOutboundRoute(ctx, d.Id(), *outboundRouteUpdate.Id, outboundRouteUpdate)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update outbound route with id %s for site %s error: %s", *outboundRouteUpdate.Id, d.Id(), err), resp)
			}
		} else {
			// Add the outbound route if not already present
			_, resp, err := proxy.createSiteOutboundRoute(ctx, d.Id(), &outboundRoute)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to add outbound route to site %s error: %s", d.Id(), err), resp)
			}
		}
	}

	// Wait for the update before reading
	time.Sleep(5 * time.Second)

	return readSiteOutboundRoutes(ctx, d, meta)
}

func deleteSiteOutboundRoutes(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSiteOutboundRouteProxy(sdkConfig)
	managedRoutes := buildOutboundRoutes(d.Get("outbound_routes").(*schema.Set))

	// Verify parent site still exists before trying to delete outbound routes
	_, resp, err := proxy.getSite(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Parent site %s already deleted", d.Id())
			return nil
		}

		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete outbound routes for site %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Deleting outbound routes for site %s", d.Id())
	apiRoutes, resp, err := proxy.getSiteOutboundRoutes(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get outbound routes for site %s for delete. Error: %s", d.Id(), err), resp)
	}

	for _, managedRoute := range *managedRoutes {
		if route, ok := nameInOutboundRoutes(*managedRoute.Name, *apiRoutes); ok {
			resp, err := proxy.deleteSiteOutboundRoute(ctx, d.Id(), *route.Id)
			if err != nil {
				if util.IsStatus404(resp) {
					log.Printf("Outbound route already deleted for site %s", d.Id())
					return nil
				}
			}
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		outboundRoutes, resp, err := proxy.getSiteOutboundRoutes(ctx, d.Id())
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to verify delete of outbound routes for site %s error: %s", d.Id(), err), resp))
		}

		if outboundRoutes == nil || len(*outboundRoutes) == 0 {
			log.Printf("Deleted all outbound routes for site %s", d.Id())
			return nil
		}

		// Verify the managed routes have been deleted
		for _, managedRoute := range *managedRoutes {
			if _, present := nameInOutboundRoutes(*managedRoute.Name, *outboundRoutes); present {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("outbound route %s for site %s still exists", *managedRoute.Name, d.Id()), resp))
			}
		}

		log.Printf("Deleted managed outbound routes for site %s", d.Id())
		return nil
	})
}
