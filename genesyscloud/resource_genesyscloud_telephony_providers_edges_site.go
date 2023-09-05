package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/leekchan/timeutil"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func ResourceSite() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Site",

		CreateContext: CreateWithPooledClient(createSite),
		ReadContext:   ReadWithPooledClient(readSite),
		UpdateContext: UpdateWithPooledClient(updateSite),
		DeleteContext: DeleteWithPooledClient(deleteSite),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location_id": {
				Description: "Site location ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"media_model": {
				Description:  "Media model for the site Valid Values: Premises, Cloud. Changing the media_model attribute will cause the site object to be dropped and created with a new ID.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Premises", "Cloud"}, false),
				ForceNew:     true,
			},
			"media_regions_use_latency_based": {
				Description: "Latency based on media region",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"media_regions": {
				Description: "The ordered list of AWS regions through which media can stream. A full list of available media regions can be found at the GET /api/v2/telephony/mediaregions endpoint",
				Type:        schema.TypeList, //This has to be a list because it must be ordered
				Optional:    true,
				Computed:    true, //This needs to be a computed field because the sites API automatically adds the home region to whatever regions you add add.
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"caller_id": {
				Description:      "The caller ID value for the site. The callerID must be a valid E.164 formatted phone number",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidatePhoneNumber,
			},
			"caller_name": {
				Description: "The caller name for the site",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"edge_auto_update_config": {
				Description: "Recurrence rule, time zone, and start/end settings for automatic edge updates for this site",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"time_zone": {
							Description: "The timezone of the window in which any updates to the edges assigned to the site can be applied. The minimum size of the window is 2 hours.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"rrule": {
							Description: "The recurrence rule for updating the Edges assigned to the site. The only supported frequencies are daily and weekly. Weekly frequencies require a day list with at least oneday specified. All other configurations are not supported.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"start": {
							Description: "Date time is represented as an ISO-8601 string without a timezone. For example: yyyy-MM-ddTHH:mm:ss.SSS",
							Type:        schema.TypeString,
							Required:    true,
						},
						"end": {
							Description: "Date time is represented as an ISO-8601 string without a timezone. For example: yyyy-MM-ddTHH:mm:ss.SSS",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"number_plans": {
				Description: "Number plans for the site. The order of the plans in the resource file determines the priority of the plans. Specifying number plans will not result in the default plans being overwritten.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the entity.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"match_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"digitLength", "e164NumberList", "interCountryCode", "intraCountryCode", "numberList", "regex"}, false),
						},
						"normalized_format": {
							Description: "Use regular expression capture groups to build the normalized number",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"match_format": {
							Description: "Use regular expression capture groups to build the normalized number",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"numbers": {
							Description: "Numbers must be 2-9 digits long. Numbers within ranges must be the same length. (e.g. 888, 888-999, 55555-77777, 800).",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"end": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"digit_length": {
							Description: "Allowed values are between 1-20 digits.",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"end": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"classification": {
							Description: "Used to classify this number plan",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"outbound_routes": {
				Description: "Outbound Routes for the site. The default outbound route will not be delete if routes are specified",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the entity.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"description": {
							Description: "The resource's description.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"classification_types": {
							Description: "Used to classify this outbound route.",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"enabled": {
							Description: "Enable or disable the outbound route",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"distribution": {
							Description:  "Valid values: SEQUENTIAL, RANDOM.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "SEQUENTIAL",
							ValidateFunc: validation.StringInSlice([]string{"SEQUENTIAL", "RANDOM"}, false),
						},
						"external_trunk_base_ids": {
							Description: "Trunk base settings of trunkType \"EXTERNAL\". This base must also be set on an edge logical interface for correct routing. The order of the IDs determines the distribution if \"distribution\" is set to \"SEQUENTIAL\"",
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"primary_sites": {
				Description: `Used for primary phone edge assignment on physical edges only.  List of primary sites the phones can be assigned to. If no primary_sites are defined, the site id for this site will be used as the primary site id.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"secondary_sites": {
				Description: `Used for secondary phone edge assignment on physical edges only.  List of secondary sites the phones can be assigned to.  If no primary_sites or secondary_sites are defined then the current site will defined as primary and secondary. `,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func getSites(_ context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	// Unmanaged
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sites, _, getErr := edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", "", "", false)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of sites: %v", getErr)
		}

		if sites.Entities == nil || len(*sites.Entities) == 0 {
			break
		}

		for _, site := range *sites.Entities {
			if site.State != nil && *site.State != "deleted" {
				resources[*site.Id] = &resourceExporter.ResourceMeta{Name: *site.Name}
			}
		}
	}

	// Managed
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sites, _, getErr := edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", "", "", true)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of sites: %v", getErr)
		}

		if sites.Entities == nil || len(*sites.Entities) == 0 {
			break
		}

		for _, site := range *sites.Entities {
			if site.State != nil && *site.State != "deleted" {
				resources[*site.Id] = &resourceExporter.ResourceMeta{Name: *site.Name}
			}
		}
	}

	return resources, nil
}

func SiteExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getSites),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"location_id": {RefType: "genesyscloud_location"},
			"outbound_routes.external_trunk_base_ids": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
			"primary_sites":   {RefType: "genesyscloud_telephony_providers_edges_site"},
			"secondary_sites": {RefType: "genesyscloud_telephony_providers_edges_site"},
		},
	}
}

func validateMediaRegions(regions *[]string, sdkConfig *platformclientv2.Configuration) error {

	telephonyAPI := platformclientv2.NewTelephonyApiWithConfig(sdkConfig)
	telephonyRegions, _, err := telephonyAPI.GetTelephonyMediaregions()

	if err != nil {
		return err
	}

	homeRegion := telephonyRegions.AwsHomeRegion
	coreRegions := telephonyRegions.AwsCoreRegions
	satRegions := telephonyRegions.AwsSatelliteRegions

	for _, region := range *regions {
		if region != *homeRegion &&
			!lists.ItemInSlice(region, *coreRegions) &&
			!lists.ItemInSlice(region, *satRegions) {
			return fmt.Errorf("region %s is not a valid media region.  please refer to the Genesys Cloud GET /api/v2/telephony/mediaregions for list of valid regions.", regions)
		}

	}

	return nil
}

func createSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	locationAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)
	location, _, err := locationAPI.GetLocation(locationId, nil)
	if err != nil {
		return diag.Errorf("Error fetching location with id %v: %v", locationId, err)
	}
	if location.EmergencyNumber == nil {
		return diag.Errorf("Location with id %v does not have an emergency number", locationId)
	}

	err = validateMediaRegions(mediaRegions, sdkConfig)
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

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Creating site %s", name)
	site, _, err = edgesAPI.PostTelephonyProvidersEdgesSites(*site)
	if err != nil {
		return diag.Errorf("Failed to create site %s: %s", name, err)
	}

	d.SetId(*site.Id)

	log.Printf("Creating updating site with primary/secondary:  %s", *site.Id)
	diagErr := updatePrimarySecondarySites(d, *site.Id, edgesAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteNumberPlans(d, edgesAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		diagErr = updateSiteOutboundRoutes(d, edgesAPI)
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading site %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentSite, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesSite(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read site %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read site %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSite())
		d.Set("name", *currentSite.Name)
		d.Set("location_id", nil)
		if currentSite.Location != nil {
			d.Set("location_id", *currentSite.Location.Id)
		}
		d.Set("media_model", *currentSite.MediaModel)
		d.Set("description", nil)
		if currentSite.Description != nil {
			d.Set("description", *currentSite.Description)
		}
		d.Set("media_regions_use_latency_based", *currentSite.MediaRegionsUseLatencyBased)

		d.Set("edge_auto_update_config", nil)
		if currentSite.EdgeAutoUpdateConfig != nil {
			d.Set("edge_auto_update_config", flattenSdkEdgeAutoUpdateConfig(currentSite.EdgeAutoUpdateConfig))
		}

		d.Set("media_regions", nil)
		if currentSite.MediaRegions != nil {
			d.Set("media_regions", *currentSite.MediaRegions)
		}

		d.Set("caller_id", currentSite.CallerId)
		d.Set("caller_name", currentSite.CallerName)

		if currentSite.PrimarySites != nil {
			d.Set("primary_sites", SdkDomainEntityRefArrToList(*currentSite.PrimarySites))
		}

		if currentSite.SecondarySites != nil {
			d.Set("secondary_sites", SdkDomainEntityRefArrToList(*currentSite.SecondarySites))
		}

		if retryErr := readSiteNumberPlans(d, edgesAPI); retryErr != nil {
			return retryErr
		}

		if retryErr := readSiteOutboundRoutes(d, edgesAPI); retryErr != nil {
			return retryErr
		}

		log.Printf("Read site %s %s", d.Id(), *currentSite.Name)
		return cc.CheckState()
	})
}

func updateSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	locationAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)
	location, _, err := locationAPI.GetLocation(locationId, nil)
	if err != nil {
		return diag.Errorf("Error fetching location with id %v: %v", locationId, err)
	}
	if location.EmergencyNumber == nil {
		return diag.Errorf("Location with id %v does not have an emergency number", locationId)
	}

	err = validateMediaRegions(mediaRegions, sdkConfig)
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
		site.PrimarySites = BuildSdkDomainEntityRefArr(d, "primary_sites")
	}

	if len(secondarySites) > 0 {
		site.SecondarySites = BuildSdkDomainEntityRefArr(d, "secondary_sites")
	}

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current site version
		currentSite, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesSite(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read site %s: %s", d.Id(), getErr)
		}
		site.Version = currentSite.Version

		log.Printf("Updating site %s", name)
		site, resp, err = edgesAPI.PutTelephonyProvidersEdgesSite(d.Id(), *site)
		if err != nil {
			return resp, diag.Errorf("Failed to update site %s: %s", name, err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteNumberPlans(d, edgesAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateSiteOutboundRoutes(d, edgesAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated site %s", *site.Id)
	time.Sleep(5 * time.Second)
	return readSite(ctx, d, meta)
}

func deleteSite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting site")
	resp, err := edgesAPI.DeleteTelephonyProvidersEdgesSite(d.Id())
	if err != nil {
		if IsStatus404(resp) {
			log.Printf("Site already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete site: %s %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		site, resp, err := edgesAPI.GetTelephonyProvidersEdgesSite(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Site deleted
				log.Printf("Deleted site %s", d.Id())
				// Need to sleep here because if terraform deletes the dependent location straight away
				// the API will think it's still in use
				time.Sleep(8 * time.Second)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting site %s: %s", d.Id(), err))
		}

		if site.State != nil && *site.State == "deleted" {
			// Site deleted
			log.Printf("Deleted site %s", d.Id())
			// Need to sleep here because if terraform deletes the dependent location straight away
			// the API will think it's still in use
			time.Sleep(8 * time.Second)
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Site %s still exists", d.Id()))
	})
}

func nameInPlans(name string, plans []platformclientv2.Numberplan) (*platformclientv2.Numberplan, bool) {
	for _, plan := range plans {
		if name == *plan.Name {
			return &plan, true
		}
	}

	return nil, false
}

func nameInOutboundRoutes(name string, outboundRoutes []platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, bool) {
	for _, outboundRoute := range outboundRoutes {
		if name == *outboundRoute.Name {
			return &outboundRoute, true
		}
	}

	return nil, false
}

// Contains the logic to determine if a primary or secondary site need to be updated.
func updatePrimarySecondarySites(d *schema.ResourceData, siteId string, edgesAPI *platformclientv2.TelephonyProvidersEdgeApi) diag.Diagnostics {
	primarySites := lists.InterfaceListToStrings(d.Get("primary_sites").([]interface{}))
	secondarySites := lists.InterfaceListToStrings(d.Get("secondary_sites").([]interface{}))

	site, resp, err := edgesAPI.GetTelephonyProvidersEdgesSite(siteId)

	if resp.StatusCode != 200 {
		return diag.Errorf("Unable to retrieve site record after site %s was created, but unable to update the primary or secondary site.  Status code %d. RespBody %s", siteId, resp.StatusCode, resp.RawBody)
	}

	if err != nil {
		return diag.Errorf("Unable to retrieve site record after site %s was created, but unable to update the primary or secondary site.  Err: %s", siteId, err)
	}

	if len(primarySites) == 0 && len(secondarySites) > 0 {
		der := platformclientv2.Domainentityref{Id: &siteId}
		derArr := make([]platformclientv2.Domainentityref, 1)
		derArr[0] = der
		site.PrimarySites = &derArr
	}

	if len(primarySites) > 0 {
		site.PrimarySites = BuildSdkDomainEntityRefArr(d, "primary_sites")
	}

	if len(secondarySites) > 0 {
		site.SecondarySites = BuildSdkDomainEntityRefArr(d, "secondary_sites")
	}

	_, resp, err = edgesAPI.PutTelephonyProvidersEdgesSite(siteId, *site)

	if resp.StatusCode != 200 {
		return diag.Errorf("Site %s was created, but unable to update the primary or secondary site.  Status code %d. RespBody %s", siteId, resp.StatusCode, resp.RawBody)
	}

	if err != nil {
		return diag.Errorf("[Site %s was created, but unable to update the primary or secondary site.  Err: %s", siteId, err)
	}

	return nil
}

func updateSiteNumberPlans(d *schema.ResourceData, edgesAPI *platformclientv2.TelephonyProvidersEdgeApi) diag.Diagnostics {
	if d.HasChange("number_plans") {
		if nps := d.Get("number_plans").([]interface{}); nps != nil {
			numberPlansFromTf := make([]platformclientv2.Numberplan, 0)
			for _, np := range nps {
				npMap := np.(map[string]interface{})
				numberPlanFromTf := platformclientv2.Numberplan{}

				if name := npMap["name"].(string); name != "" {
					numberPlanFromTf.Name = &name
				}

				if matchType := npMap["match_type"].(string); matchType != "" {
					numberPlanFromTf.MatchType = &matchType
				}

				if matchFormat := npMap["match_format"].(string); matchFormat != "" {
					numberPlanFromTf.Match = &matchFormat
				}

				if normalizedFormat := npMap["normalized_format"].(string); normalizedFormat != "" {
					numberPlanFromTf.NormalizedFormat = &normalizedFormat
				}

				if classification := npMap["classification"].(string); classification != "" {
					numberPlanFromTf.Classification = &classification
				}

				if numbers, ok := npMap["numbers"].([]interface{}); ok && len(numbers) > 0 {
					sdkNumbers := make([]platformclientv2.Number, 0)
					for _, number := range numbers {
						numberMap := number.(map[string]interface{})
						sdkNumber := platformclientv2.Number{}
						if start, ok := numberMap["start"].(string); ok {
							sdkNumber.Start = &start
						}
						if end, ok := numberMap["end"].(string); ok {
							sdkNumber.End = &end
						}
						sdkNumbers = append(sdkNumbers, sdkNumber)
					}
					numberPlanFromTf.Numbers = &sdkNumbers
				}

				if digitLength, ok := npMap["digit_length"].([]interface{}); ok && len(digitLength) > 0 {
					sdkDigitlengthMap := digitLength[0].(map[string]interface{})
					sdkDigitlength := platformclientv2.Digitlength{}
					if start, ok := sdkDigitlengthMap["start"].(string); ok {
						sdkDigitlength.Start = &start
					}
					if end, ok := sdkDigitlengthMap["end"].(string); ok {
						sdkDigitlength.End = &end
					}
					numberPlanFromTf.DigitLength = &sdkDigitlength
				}

				numberPlansFromTf = append(numberPlansFromTf, numberPlanFromTf)
			}

			// The default plans won't be assigned yet if there isn't a wait
			time.Sleep(5 * time.Second)

			numberPlansFromAPI, _, err := edgesAPI.GetTelephonyProvidersEdgesSiteNumberplans(d.Id())
			if err != nil {
				return diag.Errorf("Failed to get number plans for site %s: %s", d.Id(), err)
			}

			updatedNumberPlans := make([]platformclientv2.Numberplan, 0)

			for _, numberPlanFromTf := range numberPlansFromTf {
				if plan, ok := nameInPlans(*numberPlanFromTf.Name, numberPlansFromAPI); ok {
					// Update the plan
					plan.Classification = numberPlanFromTf.Classification
					plan.Numbers = numberPlanFromTf.Numbers
					plan.DigitLength = numberPlanFromTf.DigitLength
					plan.Match = numberPlanFromTf.Match
					plan.MatchType = numberPlanFromTf.MatchType
					plan.NormalizedFormat = numberPlanFromTf.NormalizedFormat
					updatedNumberPlans = append(updatedNumberPlans, *plan)
				} else {
					// Add the plan
					updatedNumberPlans = append(updatedNumberPlans, numberPlanFromTf)
				}
			}

			for _, numberPlanFromAPI := range numberPlansFromAPI {
				// Keep the default plans assigned.
				if isDefaultPlan(*numberPlanFromAPI.Name) {
					updatedNumberPlans = append(updatedNumberPlans, numberPlanFromAPI)
				}
			}

			diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				log.Printf("Updating number plans for site %s", d.Id())
				_, resp, err := edgesAPI.PutTelephonyProvidersEdgesSiteNumberplans(d.Id(), updatedNumberPlans)
				if err != nil {
					respString := ""
					if resp != nil {
						respString = resp.String()
					}
					return resp, diag.Errorf("Failed to update number plans for site %s: %s %s", d.Id(), err, respString)
				}
				return resp, nil
			})
			if diagErr != nil {
				return diagErr
			}
			// Wait for the update before reading
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func updateSiteOutboundRoutes(d *schema.ResourceData, edgesAPI *platformclientv2.TelephonyProvidersEdgeApi) diag.Diagnostics {
	if d.HasChange("outbound_routes") {
		if ors := d.Get("outbound_routes").([]interface{}); ors != nil {
			outboundRoutesFromTf := make([]platformclientv2.Outboundroutebase, 0)
			for _, or := range ors {
				orMap := or.(map[string]interface{})
				outboundRouteFromTf := platformclientv2.Outboundroutebase{}

				if name := orMap["name"].(string); name != "" {
					outboundRouteFromTf.Name = &name
				}
				if description := orMap["description"].(string); description != "" {
					outboundRouteFromTf.Description = &description
				}
				if classificationTypes, ok := orMap["classification_types"].([]interface{}); ok && len(classificationTypes) > 0 {
					cts := make([]string, 0)
					for _, classificationType := range classificationTypes {
						cts = append(cts, classificationType.(string))
					}
					outboundRouteFromTf.ClassificationTypes = &cts
				}
				if enabled, ok := orMap["enabled"].(bool); ok {
					outboundRouteFromTf.Enabled = &enabled
				}
				if distribution := orMap["distribution"].(string); distribution != "" {
					outboundRouteFromTf.Distribution = &distribution
				}
				if externalTrunkBaseIds, ok := orMap["external_trunk_base_ids"].([]interface{}); ok && len(externalTrunkBaseIds) > 0 {
					ids := make([]platformclientv2.Domainentityref, 0)
					for _, externalTrunkBaseId := range externalTrunkBaseIds {
						externalTrunkBaseIdStr := externalTrunkBaseId.(string)
						ids = append(ids, platformclientv2.Domainentityref{Id: &externalTrunkBaseIdStr})
					}
					outboundRouteFromTf.ExternalTrunkBases = &ids
				}

				outboundRoutesFromTf = append(outboundRoutesFromTf, outboundRouteFromTf)
			}

			// The default outbound routes won't be assigned yet if there isn't a wait
			time.Sleep(5 * time.Second)

			outboundRoutesFromAPI := make([]platformclientv2.Outboundroutebase, 0)
			for pageNum := 1; ; pageNum++ {
				const pageSize = 100
				outboundRoutes, _, err := edgesAPI.GetTelephonyProvidersEdgesSiteOutboundroutes(d.Id(), pageSize, pageNum, "", "", "")
				if err != nil {
					return diag.Errorf("Failed to get outbound routes for site %s: %s", d.Id(), err)
				}
				if outboundRoutes.Entities == nil || len(*outboundRoutes.Entities) == 0 {
					break
				}
				outboundRoutesFromAPI = append(outboundRoutesFromAPI, *outboundRoutes.Entities...)
			}

			for _, outboundRouteFromTf := range outboundRoutesFromTf {
				if outboundRoute, ok := nameInOutboundRoutes(*outboundRouteFromTf.Name, outboundRoutesFromAPI); ok {
					// Update the outbound route
					outboundRoute.Name = outboundRouteFromTf.Name
					outboundRoute.Description = outboundRouteFromTf.Description
					outboundRoute.ClassificationTypes = outboundRouteFromTf.ClassificationTypes
					outboundRoute.Enabled = outboundRouteFromTf.Enabled
					outboundRoute.Distribution = outboundRouteFromTf.Distribution
					outboundRoute.ExternalTrunkBases = outboundRouteFromTf.ExternalTrunkBases
					_, _, err := edgesAPI.PutTelephonyProvidersEdgesSiteOutboundroute(d.Id(), *outboundRoute.Id, *outboundRoute)
					if err != nil {
						return diag.Errorf("Failed to update outbound route with id %s for site %s: %s", *outboundRoute.Id, d.Id(), err)
					}
				} else {
					// Add the outbound route
					_, _, err := edgesAPI.PostTelephonyProvidersEdgesSiteOutboundroutes(d.Id(), outboundRouteFromTf)
					if err != nil {
						return diag.Errorf("Failed to add outbound route to site %s: %s", d.Id(), err)
					}
				}
			}

			for _, outboundRouteFromAPI := range outboundRoutesFromAPI {
				// Delete route if no reference to it
				if _, ok := nameInOutboundRoutes(*outboundRouteFromAPI.Name, outboundRoutesFromTf); !ok {
					resp, err := edgesAPI.DeleteTelephonyProvidersEdgesSiteOutboundroute(d.Id(), *outboundRouteFromAPI.Id)
					if err != nil {
						if IsStatus404(resp) {
							return nil
						}
						return diag.Errorf("Failed to delete outbound route from site %s: %s", d.Id(), err)
					}
				}
			}

			// Wait for the update before reading
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}

func isDefaultPlan(name string) bool {
	defaultPlans := []string{"Emergency", "Extension", "National", "International", "Network", "Suicide Prevent"}
	for _, defaultPlan := range defaultPlans {
		if name == defaultPlan {
			return true
		}
	}
	return false
}

func readSiteNumberPlans(d *schema.ResourceData, edgesAPI *platformclientv2.TelephonyProvidersEdgeApi) *retry.RetryError {
	numberPlans, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesSiteNumberplans(d.Id())
	if getErr != nil {
		if IsStatus404(resp) {
			d.SetId("") // Site doesn't exist
			return nil
		}
		return retry.NonRetryableError(fmt.Errorf("Failed to read number plans for site %s: %s", d.Id(), getErr))
	}

	dNumberPlans := make([]interface{}, 0)
	if len(numberPlans) > 0 {
		for _, numberPlan := range numberPlans {
			if isDefaultPlan(*numberPlan.Name) {
				continue
			}

			dNumberPlan := make(map[string]interface{})
			dNumberPlan["name"] = *numberPlan.Name

			if numberPlan.Match != nil {
				dNumberPlan["match_format"] = *numberPlan.Match
			}
			if numberPlan.NormalizedFormat != nil {
				dNumberPlan["normalized_format"] = *numberPlan.NormalizedFormat
			}
			if numberPlan.Classification != nil {
				dNumberPlan["classification"] = *numberPlan.Classification
			}
			if numberPlan.MatchType != nil {
				dNumberPlan["match_type"] = *numberPlan.MatchType
			}

			if numberPlan.Numbers != nil {
				numbers := make([]interface{}, 0)
				for _, number := range *numberPlan.Numbers {
					numberMap := make(map[string]interface{})
					if number.Start != nil {
						numberMap["start"] = *number.Start
					}
					if number.End != nil {
						numberMap["end"] = *number.End
					}
					numbers = append(numbers, numberMap)
				}
				dNumberPlan["numbers"] = numbers
			}
			if numberPlan.DigitLength != nil {
				digitLength := make([]interface{}, 0)
				digitLengthMap := make(map[string]interface{})
				if numberPlan.DigitLength.Start != nil {
					digitLengthMap["start"] = *numberPlan.DigitLength.Start
				}
				if numberPlan.DigitLength.End != nil {
					digitLengthMap["end"] = *numberPlan.DigitLength.End
				}
				digitLength = append(digitLength, digitLengthMap)
				dNumberPlan["digit_length"] = digitLength
			}

			dNumberPlans = append(dNumberPlans, dNumberPlan)
		}
		d.Set("number_plans", dNumberPlans)
	} else {
		d.Set("number_plans", nil)
	}

	return nil
}

func readSiteOutboundRoutes(d *schema.ResourceData, edgesAPI *platformclientv2.TelephonyProvidersEdgeApi) *retry.RetryError {
	outboundRoutes := make([]platformclientv2.Outboundroutebase, 0)
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		outboundRouteEntityListing, _, err := edgesAPI.GetTelephonyProvidersEdgesSiteOutboundroutes(d.Id(), pageSize, pageNum, "", "", "")
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get outbound routes for site %s: %s", d.Id(), err))
		}
		if outboundRouteEntityListing.Entities == nil || len(*outboundRouteEntityListing.Entities) == 0 {
			break
		}
		outboundRoutes = append(outboundRoutes, *outboundRouteEntityListing.Entities...)
	}

	dOutboundRoutes := make([]interface{}, 0)

	if len(outboundRoutes) > 0 {
		for _, outboundRoute := range outboundRoutes {
			dOutboundRoute := make(map[string]interface{})
			dOutboundRoute["name"] = *outboundRoute.Name

			if outboundRoute.Description != nil {
				dOutboundRoute["description"] = *outboundRoute.Description
			}

			if outboundRoute.ClassificationTypes != nil {
				dOutboundRoute["classification_types"] = *outboundRoute.ClassificationTypes
			}

			if outboundRoute.Enabled != nil {
				dOutboundRoute["enabled"] = *outboundRoute.Enabled
			}

			if outboundRoute.Distribution != nil {
				dOutboundRoute["distribution"] = *outboundRoute.Distribution
			}

			if len(*outboundRoute.ExternalTrunkBases) > 0 {
				externalTrunkBaseIds := make([]string, 0)
				for _, externalTrunkBase := range *outboundRoute.ExternalTrunkBases {
					externalTrunkBaseIds = append(externalTrunkBaseIds, *externalTrunkBase.Id)
				}
				dOutboundRoute["external_trunk_base_ids"] = externalTrunkBaseIds
			}

			dOutboundRoutes = append(dOutboundRoutes, dOutboundRoute)
		}
		d.Set("outbound_routes", dOutboundRoutes)
	} else {
		d.Set("outbound_routes", nil)
	}

	return nil
}

func flattenSdkEdgeAutoUpdateConfig(edgeAutoUpdateConfig *platformclientv2.Edgeautoupdateconfig) []interface{} {
	if edgeAutoUpdateConfig == nil {
		return nil
	}

	edgeAutoUpdateConfigMap := make(map[string]interface{})
	edgeAutoUpdateConfigMap["time_zone"] = *edgeAutoUpdateConfig.TimeZone
	edgeAutoUpdateConfigMap["rrule"] = *edgeAutoUpdateConfig.Rrule
	edgeAutoUpdateConfigMap["start"] = timeutil.Strftime(edgeAutoUpdateConfig.Start, "%Y-%m-%dT%H:%M:%S.%f")
	edgeAutoUpdateConfigMap["end"] = timeutil.Strftime(edgeAutoUpdateConfig.End, "%Y-%m-%dT%H:%M:%S.%f")

	return []interface{}{edgeAutoUpdateConfigMap}
}

func buildSdkEdgeAutoUpdateConfig(d *schema.ResourceData) (*platformclientv2.Edgeautoupdateconfig, error) {
	if edgeAutoUpdateConfig := d.Get("edge_auto_update_config"); edgeAutoUpdateConfig != nil {
		if edgeAutoUpdateConfigList := edgeAutoUpdateConfig.([]interface{}); len(edgeAutoUpdateConfigList) > 0 {
			edgeAutoUpdateConfigMap := edgeAutoUpdateConfigList[0].(map[string]interface{})

			timeZone := edgeAutoUpdateConfigMap["time_zone"].(string)
			rrule := edgeAutoUpdateConfigMap["rrule"].(string)
			startStr := edgeAutoUpdateConfigMap["start"].(string)
			endStr := edgeAutoUpdateConfigMap["end"].(string)

			start, err := time.Parse("2006-01-02T15:04:05.000000", startStr)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse date %s: %s", startStr, err)
			}

			end, err := time.Parse("2006-01-02T15:04:05.000000", endStr)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse date %s: %s", end, err)
			}

			return &platformclientv2.Edgeautoupdateconfig{
				TimeZone: &timeZone,
				Rrule:    &rrule,
				Start:    &start,
				End:      &end,
			}, nil
		}
	}

	return nil, nil
}

func GenerateSiteResourceWithCustomAttrs(
	siteRes,
	name,
	description,
	locationId,
	mediaModel string,
	mediaRegionsUseLatencyBased bool,
	mediaRegions string,
	callerId string,
	callerName string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_site" "%s" {
		name = "%s"
		description = "%s"
		location_id = %s
		media_model = "%s"
		media_regions_use_latency_based = %v
		media_regions= %s
		caller_id = %s
		caller_name = %s
		%s
	}
	`, siteRes, name, description, locationId, mediaModel, mediaRegionsUseLatencyBased, mediaRegions, callerId, callerName, strings.Join(otherAttrs, "\n"))
}

func DeleteLocationWithNumber(emergencyNumber string) error {
	sdkConfig := platformclientv2.GetDefaultConfiguration()
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		locations, _, getErr := locationsAPI.GetLocations(pageSize, pageNum, nil, "")
		if getErr != nil {
			return getErr
		}

		if locations.Entities == nil || len(*locations.Entities) == 0 {
			break
		}

		for _, location := range *locations.Entities {
			if location.EmergencyNumber != nil {
				if strings.Contains(*location.EmergencyNumber.E164, emergencyNumber) {
					err := deleteSiteWithLocationId(*location.Id)
					if err != nil {
						return err
					}
					_, err = locationsAPI.DeleteLocation(*location.Id)
					time.Sleep(30 * time.Second)
					return err
				}
			}
		}
	}

	return nil
}

func deleteSiteWithLocationId(locationId string) error {
	sdkConfig := platformclientv2.GetDefaultConfiguration()
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sites, _, getErr := edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", "", "", false)
		if getErr != nil {
			return getErr
		}

		if sites.Entities == nil || len(*sites.Entities) == 0 {
			return nil
		}

		for _, site := range *sites.Entities {
			if site.Location != nil && *site.Location.Id == locationId {
				_, err := edgesAPI.DeleteTelephonyProvidersEdgesSite(*site.Id)
				if err != nil {
					return err
				}
				time.Sleep(8 * time.Second)
				break
			}
		}
	}
}
