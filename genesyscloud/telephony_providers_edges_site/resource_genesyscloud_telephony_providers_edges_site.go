package telephony_providers_edges_site

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getSites(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	sp := getSiteProxy(sdkConfig)

	unmanagedSites, err := sp.getAllUnmanagedSites(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	for _, unmanagedSite := range *unmanagedSites {
		resources[*unmanagedSite.Id] = &resourceExporter.ResourceMeta{Name: *unmanagedSite.Name}
	}

	managedSites, err := sp.getAllManagedSites(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	for _, managedSite := range *managedSites {
		resources[*managedSite.Id] = &resourceExporter.ResourceMeta{Name: *managedSite.Name}
	}

	return resources, nil
}

func createSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

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
	location, err := sp.getLocation(ctx, locationId)
	if err != nil {
		return diag.FromErr(err)
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
	site, err := sp.createSite(ctx, siteReq)
	if err != nil {
		return diag.FromErr(err)
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

	diagErr = gcloud.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		diagErr = updateSiteOutboundRoutes(ctx, sp, d)
		if diagErr != nil {
			return retry.RetryableError(fmt.Errorf(fmt.Sprintf("%v", diagErr), d.Id()))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created site %s", *site.Id)

	// Default site
	if d.Get("set_as_default_site").(bool) {
		log.Printf("Setting default site to %s", *site.Id)
		err := sp.setDefaultSite(ctx, *site.Id)
		if err != nil {
			return diag.Errorf("unable to set default site to %s. err: %v", *site.Id, err)
		}
	}

	return readSite(ctx, d, meta)
}

func readSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

	log.Printf("Reading site %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentSite, resp, err := sp.getSiteById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read site %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read site %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSite())
		d.Set("name", *currentSite.Name)
		d.Set("location_id", nil)
		if currentSite.Location != nil {
			d.Set("location_id", *currentSite.Location.Id)
		}
		d.Set("media_model", *currentSite.MediaModel)
		d.Set("media_regions_use_latency_based", *currentSite.MediaRegionsUseLatencyBased)

		resourcedata.SetNillableValue(d, "description", currentSite.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "edge_auto_update_config", currentSite.EdgeAutoUpdateConfig, flattenSdkEdgeAutoUpdateConfig)
		resourcedata.SetNillableValue(d, "media_regions", currentSite.MediaRegions)

		d.Set("caller_id", currentSite.CallerId)
		d.Set("caller_name", currentSite.CallerName)

		if currentSite.PrimarySites != nil {
			d.Set("primary_sites", gcloud.SdkDomainEntityRefArrToList(*currentSite.PrimarySites))
		}

		if currentSite.SecondarySites != nil {
			d.Set("secondary_sites", gcloud.SdkDomainEntityRefArrToList(*currentSite.SecondarySites))
		}

		if retryErr := readSiteNumberPlans(ctx, sp, d); retryErr != nil {
			return retryErr
		}

		if retryErr := readSiteOutboundRoutes(ctx, sp, d); retryErr != nil {
			return retryErr
		}

		defaultSiteId, err := sp.getDefaultSiteId(ctx)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to get default site id: %v", err))
		}
		d.Set("set_as_default_site", defaultSiteId == *currentSite.Id)

		log.Printf("Read site %s %s", d.Id(), *currentSite.Name)
		return cc.CheckState()
	})
}

func updateSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

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

	location, err := sp.getLocation(ctx, locationId)
	if err != nil {
		return diag.FromErr(err)
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
		site.PrimarySites = gcloud.BuildSdkDomainEntityRefArr(d, "primary_sites")
	}

	if len(secondarySites) > 0 {
		site.SecondarySites = gcloud.BuildSdkDomainEntityRefArr(d, "secondary_sites")
	}

	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current site version
		currentSite, resp, err := sp.getSiteById(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to read site %s: %s", d.Id(), err)
		}
		site.Version = currentSite.Version

		log.Printf("Updating site %s", *site.Name)
		site, resp, err = sp.updateSite(ctx, d.Id(), site)
		if err != nil {
			return resp, diag.Errorf("Failed to update site %s: %s", *site.Name, err)
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

	diagErr = updateSiteOutboundRoutes(ctx, sp, d)
	if diagErr != nil {
		return diagErr
	}

	if d.Get("set_as_default_site").(bool) {
		log.Printf("Setting default site to %s", *site.Id)
		err := sp.setDefaultSite(ctx, *site.Id)
		if err != nil {
			return diag.Errorf("unable to set default site to %s. err: %v", *site.Id, err)
		}
	}

	log.Printf("Updated site %s", *site.Id)
	time.Sleep(5 * time.Second)
	return readSite(ctx, d, meta)
}

func deleteSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

	log.Printf("Deleting site")
	resp, err := sp.deleteSite(ctx, d.Id())
	if err != nil {
		if gcloud.IsStatus404(resp) {
			log.Printf("Site already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("failed to delete site: %s %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		site, resp, err := sp.getSiteById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Site deleted
				log.Printf("Deleted site %s", d.Id())
				// Need to sleep here because if terraform deletes the dependent location straight away
				// the API will think it's still in use
				time.Sleep(8 * time.Second)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting site %s: %s", d.Id(), err))
		}

		if site.State != nil && *site.State == "deleted" {
			// Site deleted
			log.Printf("Deleted site %s", d.Id())
			// Need to sleep here because if terraform deletes the dependent location straight away
			// the API will think it's still in use
			time.Sleep(8 * time.Second)
			return nil
		}

		return retry.RetryableError(fmt.Errorf("site %s still exists", d.Id()))
	})
}
