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
	"github.com/mypurecloud/platform-client-sdk-go/v109/platformclientv2"
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

	name := d.Get("name").(string)
	locationId := d.Get("location_id").(string)
	mediaModel := d.Get("media_model").(string)
	description := d.Get("description").(string)
	mediaRegionsUseLatencyBased := d.Get("media_regions_use_latency_based").(bool)
	edgeAutoUpdateConfig, err := buildSdkEdgeAutoUpdateConfig(d)

	if err != nil {
		return diag.FromErr(err)
	}

	mediaRegions := lists.BuildSdkStringListFromInterfaceArray(d, "media_regions")
	callerID := d.Get("caller_id").(string)
	callerName := d.Get("caller_name").(string)

	location, err := sp.getLocation(ctx, locationId)
	if err != nil {
		return diag.FromErr(err)
	}

	err = validateMediaRegions(mediaRegions, sp, ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	siteReq := &platformclientv2.Site{
		Name: &name,
		Location: &platformclientv2.Locationdefinition{
			Id:              &locationId,
			EmergencyNumber: location.EmergencyNumber,
		},
		MediaModel:                  &mediaModel,
		MediaRegionsUseLatencyBased: &mediaRegionsUseLatencyBased,
	}

	if edgeAutoUpdateConfig != nil {
		siteReq.EdgeAutoUpdateConfig = edgeAutoUpdateConfig
	}

	if mediaRegions != nil {
		siteReq.MediaRegions = mediaRegions
	}

	if callerID != "" {
		siteReq.CallerId = &callerID
	}

	if callerName != "" {
		siteReq.CallerName = &callerName
	}

	if description != "" {
		siteReq.Description = &description
	}

	log.Printf("Creating site %s", name)
	site, err := sp.createSite(ctx, siteReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*site.Id)

	log.Printf("Creating updating site with primary/secondary:  %s", *site.Id)
	diagErr := updatePrimarySecondarySites(d, *site.Id, sp, ctx)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteNumberPlans(d, sp, ctx)
	if diagErr != nil {
		return diagErr
	}

	diagErr = gcloud.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		diagErr = updateSiteOutboundRoutes(d, sp, ctx)
		if diagErr != nil {
			return retry.RetryableError(fmt.Errorf(fmt.Sprintf("%v", diagErr), d.Id()))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created site %s", *site.Id)
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

		if retryErr := readSiteNumberPlans(d, sp, ctx); retryErr != nil {
			return retryErr
		}

		if retryErr := readSiteOutboundRoutes(d, sp, ctx); retryErr != nil {
			return retryErr
		}

		log.Printf("Read site %s %s", d.Id(), *currentSite.Name)
		return cc.CheckState()
	})
}

func updateSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

	name := d.Get("name").(string)
	locationId := d.Get("location_id").(string)
	mediaModel := d.Get("media_model").(string)
	description := d.Get("description").(string)
	mediaRegionsUseLatencyBased := d.Get("media_regions_use_latency_based").(bool)
	edgeAutoUpdateConfig, err := buildSdkEdgeAutoUpdateConfig(d)
	primarySites := lists.InterfaceListToStrings(d.Get("primary_sites").([]interface{}))
	secondarySites := lists.InterfaceListToStrings(d.Get("secondary_sites").([]interface{}))

	if err != nil {
		return diag.FromErr(err)
	}

	mediaRegions := lists.BuildSdkStringListFromInterfaceArray(d, "media_regions")
	callerID := d.Get("caller_id").(string)
	callerName := d.Get("caller_name").(string)

	location, err := sp.getLocation(ctx, locationId)
	if err != nil {
		return diag.FromErr(err)
	}

	err = validateMediaRegions(mediaRegions, sp, ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	site := &platformclientv2.Site{
		Name: &name,
		Location: &platformclientv2.Locationdefinition{
			Id:              &locationId,
			EmergencyNumber: location.EmergencyNumber,
		},
		MediaModel:                  &mediaModel,
		MediaRegionsUseLatencyBased: &mediaRegionsUseLatencyBased,
	}

	if edgeAutoUpdateConfig != nil {
		site.EdgeAutoUpdateConfig = edgeAutoUpdateConfig
	}

	if mediaRegions != nil {
		site.MediaRegions = mediaRegions
	}

	if callerID != "" {
		site.CallerId = &callerID
	}

	if callerName != "" {
		site.CallerName = &callerName
	}

	if description != "" {
		site.Description = &description
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

		log.Printf("Updating site %s", name)
		site, resp, err = sp.updateSite(ctx, d.Id(), site)
		if err != nil {
			return resp, diag.Errorf("Failed to update site %s: %s", name, err)
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteNumberPlans(d, sp, ctx)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteOutboundRoutes(d, sp, ctx)
	if diagErr != nil {
		return diagErr
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
