package telephony_providers_edges_site

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllSites(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	sp := GetSiteProxy(sdkConfig)

	// get unmanaged sites
	unmanagedSites, resp, err := sp.GetAllSites(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get unmanaged sites error: %s", err), resp)
	}
	for _, unmanagedSite := range *unmanagedSites {
		resources[*unmanagedSite.Id] = &resourceExporter.ResourceMeta{BlockLabel: *unmanagedSite.Name}
	}

	// get managed sites
	managedSites, resp, err := sp.GetAllSites(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get managed sites error: %s", err), resp)
	}
	for _, managedSite := range *managedSites {
		resources[*managedSite.Id] = &resourceExporter.ResourceMeta{BlockLabel: *managedSite.Name}
		// When exporting managed sites, they must automatically be exported as data source
		// Managed sites are added to the ExportAsData []string in resource_exporter
		if tfexporter_state.IsExporterActive() {
			resourceExporter.AddDataSourceItems(ResourceType, *managedSite.Name)
		}
	}
	return resources, nil
}

func createSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	sp := GetSiteProxy(sdkConfig)

	siteReq := &platformclientv2.Site{
		Name:                        platformclientv2.String(d.Get("name").(string)),
		CallerId:                    platformclientv2.String(d.Get("caller_id").(string)),
		CallerName:                  platformclientv2.String(d.Get("caller_name").(string)),
		MediaModel:                  platformclientv2.String(d.Get("media_model").(string)),
		Description:                 platformclientv2.String(d.Get("description").(string)),
		MediaRegionsUseLatencyBased: platformclientv2.Bool(d.Get("media_regions_use_latency_based").(bool)),
	}

	edgeAutoUpdateConfig, err := buildSdkEdgeAutoUpdateConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	mediaRegions := lists.BuildSdkStringListFromInterfaceArray(d, "media_regions")

	locationId := d.Get("location_id").(string)
	location, resp, err := sp.getLocation(ctx, locationId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get location %s error: %s", locationId, err), resp)
	}

	err = validateMediaRegions(ctx, sp, mediaRegions)
	if err != nil {
		return diag.FromErr(err)
	}

	siteReq.Location = &platformclientv2.Locationdefinition{
		Id:              platformclientv2.String(locationId),
		EmergencyNumber: location.EmergencyNumber,
	}

	if edgeAutoUpdateConfig != nil {
		siteReq.EdgeAutoUpdateConfig = edgeAutoUpdateConfig
	}

	if mediaRegions != nil {
		siteReq.MediaRegions = mediaRegions
	}

	log.Printf("Creating site %s", *siteReq.Name)
	site, resp, err := sp.createSite(ctx, siteReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create site %s error: %s", *siteReq.Name, err), resp)
	}

	d.SetId(*site.Id)

	log.Printf("Creating updating site with primary/secondary:  %s", *site.Id)
	diagErr := updatePrimarySecondarySites(ctx, sp, d, *site.Id)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteNumberPlans(ctx, sp, d)
	if diagErr != nil {
		return diagErr
	}

	if !featureToggles.OutboundRoutesToggleExists() {
		diagErr = util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
			diagErr = updateSiteOutboundRoutes(ctx, sp, d)
			if diagErr != nil {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to create site %s | error: %v", d.Id(), diagErr), nil))
			}
			return nil
		})
		if diagErr != nil {
			return diagErr
		}
	} else {
		log.Printf("%s is set, not managing outbound_routes attribute in site %s resource", featureToggles.OutboundRoutesToggleName(), d.Id())
	}

	log.Printf("Created site %s", *site.Id)

	// Default site
	if d.Get("set_as_default_site").(bool) {
		log.Printf("Setting default site to %s", *site.Id)
		resp, err := sp.setDefaultSite(ctx, *site.Id)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("unable to set default site to %s error: %s", *site.Id, err), resp)
		}
	}

	return readSite(ctx, d, meta)
}

func readSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	sp := GetSiteProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSite(), constants.ConsistencyChecks(), ResourceType)
	utilE164 := util.NewUtilE164Service()

	log.Printf("Reading site %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentSite, resp, err := sp.getSiteById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read site %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read site %s | error: %s", d.Id(), err), resp))
		}

		_ = d.Set("name", *currentSite.Name)
		_ = d.Set("location_id", nil)
		if currentSite.Location != nil {
			_ = d.Set("location_id", *currentSite.Location.Id)
		}
		_ = d.Set("media_model", *currentSite.MediaModel)
		_ = d.Set("media_regions_use_latency_based", *currentSite.MediaRegionsUseLatencyBased)

		resourcedata.SetNillableValue(d, "description", currentSite.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "edge_auto_update_config", currentSite.EdgeAutoUpdateConfig, flattenSdkEdgeAutoUpdateConfig)
		resourcedata.SetNillableValue(d, "media_regions", currentSite.MediaRegions)

		d.Set("caller_id", nil)
		if currentSite.CallerId != nil && *currentSite.CallerId != "" {
			_ = d.Set("caller_id", utilE164.FormatAsCalculatedE164Number(*currentSite.CallerId))
		}
		_ = d.Set("caller_name", currentSite.CallerName)

		if currentSite.PrimarySites != nil {
			_ = d.Set("primary_sites", util.SdkDomainEntityRefArrToList(*currentSite.PrimarySites))
		}

		if currentSite.SecondarySites != nil {
			_ = d.Set("secondary_sites", util.SdkDomainEntityRefArrToList(*currentSite.SecondarySites))
		}

		if retryErr := readSiteNumberPlans(ctx, sp, d); retryErr != nil {
			return retryErr
		}

		if !featureToggles.OutboundRoutesToggleExists() {
			if retryErr := readSiteOutboundRoutes(ctx, sp, d); retryErr != nil {
				return retryErr
			}
		} else {
			log.Printf("%s is set, not managing outbound_routes attribute in site %s resource", featureToggles.OutboundRoutesToggleName(), d.Id())
		}

		defaultSiteId, resp, err := sp.getDefaultSiteId(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get default site id: %v", err), resp))
		}
		_ = d.Set("set_as_default_site", defaultSiteId == *currentSite.Id)

		log.Printf("Read site %s %s", d.Id(), *currentSite.Name)
		return cc.CheckState(d)
	})
}

func updateSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	sp := GetSiteProxy(sdkConfig)

	site := &platformclientv2.Site{
		Name:                        platformclientv2.String(d.Get("name").(string)),
		CallerId:                    platformclientv2.String(d.Get("caller_id").(string)),
		CallerName:                  platformclientv2.String(d.Get("caller_name").(string)),
		MediaModel:                  platformclientv2.String(d.Get("media_model").(string)),
		Description:                 platformclientv2.String(d.Get("description").(string)),
		MediaRegionsUseLatencyBased: platformclientv2.Bool(d.Get("media_regions_use_latency_based").(bool)),
	}

	locationId := d.Get("location_id").(string)
	edgeAutoUpdateConfig, err := buildSdkEdgeAutoUpdateConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	primarySites := lists.InterfaceListToStrings(d.Get("primary_sites").([]interface{}))
	secondarySites := lists.InterfaceListToStrings(d.Get("secondary_sites").([]interface{}))

	mediaRegions := lists.BuildSdkStringListFromInterfaceArray(d, "media_regions")

	location, resp, err := sp.getLocation(ctx, locationId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get location %s error: %s", locationId, err), resp)
	}
	site.Location = &platformclientv2.Locationdefinition{
		Id:              &locationId,
		EmergencyNumber: location.EmergencyNumber,
	}

	err = validateMediaRegions(ctx, sp, mediaRegions)
	if err != nil {
		return diag.FromErr(err)
	}

	if edgeAutoUpdateConfig != nil {
		site.EdgeAutoUpdateConfig = edgeAutoUpdateConfig
	}

	if mediaRegions != nil {
		site.MediaRegions = mediaRegions
	}

	if len(primarySites) > 0 {
		site.PrimarySites = util.BuildSdkDomainEntityRefArr(d, "primary_sites")
	}

	if len(secondarySites) > 0 {
		site.SecondarySites = util.BuildSdkDomainEntityRefArr(d, "secondary_sites")
	}

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current site version
		currentSite, resp, err := sp.getSiteById(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read site %s error: %s", d.Id(), err), resp)
		}
		site.Version = currentSite.Version

		log.Printf("Updating site %s", *site.Name)
		site, resp, err = sp.updateSite(ctx, d.Id(), site)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update site %s error: %s", *site.Name, err), resp)
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteNumberPlans(ctx, sp, d)
	if diagErr != nil {
		return diagErr
	}

	if !featureToggles.OutboundRoutesToggleExists() {
		diagErr = updateSiteOutboundRoutes(ctx, sp, d)
		if diagErr != nil {
			return diagErr
		}
	} else {
		log.Printf("%s is set, not managing outbound_routes attribute in site %s resource", featureToggles.OutboundRoutesToggleName(), d.Id())
	}

	if d.Get("set_as_default_site").(bool) {
		log.Printf("Setting default site to %s", *site.Id)
		resp, err := sp.setDefaultSite(ctx, *site.Id)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to to set default site to %s error: %s", *site.Id, err), resp)
		}
	}

	log.Printf("Updated site %s", *site.Id)
	time.Sleep(5 * time.Second)
	return readSite(ctx, d, meta)
}

func deleteSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	sp := GetSiteProxy(sdkConfig)

	// A site linked to a trunk will not be able to be deleted until that trunk is deleted. Retrying here to make sure it is cleared properly.
	log.Printf("Deleting site %s", d.Id())
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting site %s", d.Id())
		resp, err := sp.deleteSite(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Site already deleted %s", d.Id())
				return resp, nil
			}
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete site %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		site, resp, err := sp.getSiteById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Site deleted
				log.Printf("Deleted site %s", d.Id())
				// Need to sleep here because if terraform deletes the dependent location straight away
				// the API will think it's still in use
				time.Sleep(8 * time.Second)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting site %s | error: %s", d.Id(), err), resp))
		}

		if site.State != nil && *site.State == "deleted" {
			// Site deleted
			log.Printf("Deleted site %s", d.Id())
			// Need to sleep here because if terraform deletes the dependent location straight away
			// the API will think it's still in use
			time.Sleep(8 * time.Second)
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("site %s still exists", d.Id()), resp))
	})
}
