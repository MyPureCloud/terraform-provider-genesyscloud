package telephony_providers_edges_site

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	defaultPlans = []string{"Emergency", "Extension", "National", "International", "Network", "Suicide Prevention"}
)

func customizeSiteDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.HasChange("number_plans") {
		oldNumberPlans, newNumberPlans := diff.GetChange("number_plans")
		oldNumberPlansList := oldNumberPlans.([]interface{})
		newNumberPlansList := newNumberPlans.([]interface{})

		if len(oldNumberPlansList) <= len(newNumberPlansList) {
			return nil
		}

		sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
		edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

		siteId := diff.Id()
		if siteId == "" {
			return nil
		}

		numberPlansFromApi, resp, err := edgesAPI.GetTelephonyProvidersEdgesSiteNumberplans(siteId)
		if err != nil {
			return fmt.Errorf("failed to get number plans from site %s: %s %v", siteId, err, resp)
		}

		for _, np := range numberPlansFromApi {
			if isDefaultPlan(*np.Name) && isNumberPlanInConfig(*np.Name, oldNumberPlansList) && !isNumberPlanInConfig(*np.Name, newNumberPlansList) {
				newNumberPlansList = append(newNumberPlansList, flattenNumberPlan(&np))
			}
		}

		for i, x := range newNumberPlansList {
			log.Printf("%v: %v", i, x)
		}
		_ = diff.SetNew("number_plans", newNumberPlansList)
	}
	return nil
}

func validateMediaRegions(ctx context.Context, sp *SiteProxy, regions *[]string) error {
	telephonyRegions, _, err := sp.getTelephonyMediaregions(ctx)
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
			return fmt.Errorf("region %s is not a valid media region.  please refer to the Genesys Cloud GET /api/v2/telephony/mediaregions for list of valid regions", regions)
		}

	}

	return nil
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
func updatePrimarySecondarySites(ctx context.Context, sp *SiteProxy, d *schema.ResourceData, siteId string) diag.Diagnostics {
	primarySites := lists.InterfaceListToStrings(d.Get("primary_sites").([]interface{}))
	secondarySites := lists.InterfaceListToStrings(d.Get("secondary_sites").([]interface{}))

	site, resp, err := sp.getSiteById(ctx, siteId)
	if resp.StatusCode != 200 {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Unable to retrieve site record after site %s was created, but unable to update the primary or secondary site error: %s", siteId, err), resp)
	}
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Unable to retrieve site record after site %s was created, but unable to update the primary or secondary siteerror: %s ", siteId, err), resp)
	}

	if len(primarySites) == 0 && len(secondarySites) > 0 {
		der := platformclientv2.Domainentityref{Id: &siteId}
		derArr := make([]platformclientv2.Domainentityref, 1)
		derArr[0] = der
		site.PrimarySites = &derArr
	}

	if len(primarySites) > 0 {
		site.PrimarySites = util.BuildSdkDomainEntityRefArr(d, "primary_sites")
	}

	if len(secondarySites) > 0 {
		site.SecondarySites = util.BuildSdkDomainEntityRefArr(d, "secondary_sites")
	}

	_, resp, err = sp.updateSite(ctx, siteId, site)
	if resp.StatusCode != 200 {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Site %s was created, but unable to update the primary or secondary site. Status code %d. RespBody %s", siteId, resp.StatusCode, resp.RawBody), resp)
	}
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Site %s was created, but unable to update the primary or secondary site | error: %s", siteId, err), resp)
	}

	return nil
}

func updateSiteNumberPlans(ctx context.Context, sp *SiteProxy, d *schema.ResourceData) diag.Diagnostics {
	if !d.HasChange("number_plans") {
		return nil
	}
	nps := d.Get("number_plans").([]interface{})
	if nps == nil {
		return nil
	}

	numberPlansFromTf := make([]platformclientv2.Numberplan, 0)
	for _, np := range nps {
		npMap := np.(map[string]interface{})
		numberPlanFromTf := platformclientv2.Numberplan{}

		resourcedata.BuildSDKStringValueIfNotNil(&numberPlanFromTf.Name, npMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&numberPlanFromTf.MatchType, npMap, "match_type")
		resourcedata.BuildSDKStringValueIfNotNil(&numberPlanFromTf.Match, npMap, "match_format")
		resourcedata.BuildSDKStringValueIfNotNil(&numberPlanFromTf.NormalizedFormat, npMap, "normalized_format")
		resourcedata.BuildSDKStringValueIfNotNil(&numberPlanFromTf.Classification, npMap, "classification")

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

	numberPlansFromAPI, resp, err := sp.getSiteNumberPlans(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get number plans for site %s error: %s", d.Id(), err), resp)
	}

	updatedNumberPlans := make([]platformclientv2.Numberplan, 0)
	namesOfOverridenDefaults := []string{}

	for _, numberPlanFromTf := range numberPlansFromTf {
		if plan, ok := nameInPlans(*numberPlanFromTf.Name, *numberPlansFromAPI); ok {
			// Update the plan
			plan.Classification = numberPlanFromTf.Classification
			plan.Numbers = numberPlanFromTf.Numbers
			plan.DigitLength = numberPlanFromTf.DigitLength
			plan.Match = numberPlanFromTf.Match
			plan.MatchType = numberPlanFromTf.MatchType
			plan.NormalizedFormat = numberPlanFromTf.NormalizedFormat

			namesOfOverridenDefaults = append(namesOfOverridenDefaults, *numberPlanFromTf.Name)
			updatedNumberPlans = append(updatedNumberPlans, *plan)
		} else {
			// Add the plan
			updatedNumberPlans = append(updatedNumberPlans, numberPlanFromTf)
		}
	}

	for _, numberPlanFromAPI := range *numberPlansFromAPI {
		// Keep the default plans which are not overriden.
		if isDefaultPlan(*numberPlanFromAPI.Name) && !lists.ItemInSlice(*numberPlanFromAPI.Name, namesOfOverridenDefaults) {
			updatedNumberPlans = append(updatedNumberPlans, numberPlanFromAPI)
		}
	}

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Updating number plans for site %s", d.Id())

		_, resp, err := sp.updateSiteNumberPlans(ctx, d.Id(), &updatedNumberPlans)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update number plans for site %s | error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}
	// Wait for the update before reading
	time.Sleep(5 * time.Second)

	return nil
}

func updateSiteOutboundRoutes(ctx context.Context, sp *SiteProxy, d *schema.ResourceData) diag.Diagnostics {
	if !d.HasChange("outbound_routes") {
		return nil
	}
	ors := d.Get("outbound_routes").(*schema.Set)
	if ors == nil {
		return nil
	}
	orsList := ors.List()

	outboundRoutesFromTf := make([]platformclientv2.Outboundroutebase, 0)
	for _, or := range orsList {
		orMap := or.(map[string]interface{})
		outboundRouteFromTf := platformclientv2.Outboundroutebase{}

		resourcedata.BuildSDKStringValueIfNotNil(&outboundRouteFromTf.Name, orMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&outboundRouteFromTf.Description, orMap, "description")

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
		resourcedata.BuildSDKStringValueIfNotNil(&outboundRouteFromTf.Distribution, orMap, "distribution")

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

	// Get the current outbound routes
	outboundRoutesFromAPI, resp, err := sp.getSiteOutboundRoutes(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get outbound routes for site %s error: %s", d.Id(), err), resp)
	}

	// Delete unwanted outbound roues first to free up classifications assigned to them
	for _, outboundRouteFromAPI := range *outboundRoutesFromAPI {
		// Delete route if no reference to it
		if _, ok := nameInOutboundRoutes(*outboundRouteFromAPI.Name, outboundRoutesFromTf); !ok {
			resp, err := sp.deleteSiteOutboundRoute(ctx, d.Id(), *outboundRouteFromAPI.Id)
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
	for _, outboundRouteFromTf := range outboundRoutesFromTf {
		if outboundRoute, ok := nameInOutboundRoutes(*outboundRouteFromTf.Name, *outboundRoutesFromAPI); ok {
			// Update the outbound route
			outboundRoute.Name = outboundRouteFromTf.Name
			outboundRoute.Description = outboundRouteFromTf.Description
			outboundRoute.ClassificationTypes = outboundRouteFromTf.ClassificationTypes
			outboundRoute.Enabled = outboundRouteFromTf.Enabled
			outboundRoute.Distribution = outboundRouteFromTf.Distribution
			outboundRoute.ExternalTrunkBases = outboundRouteFromTf.ExternalTrunkBases

			_, resp, err := sp.updateSiteOutboundRoute(ctx, d.Id(), *outboundRoute.Id, outboundRoute)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update outbound route with id %s for site %s error: %s", *outboundRoute.Id, d.Id(), err), resp)
			}
		} else {
			// Add the outbound route
			_, resp, err := sp.createSiteOutboundRoute(ctx, d.Id(), &outboundRouteFromTf)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to add outbound route to site %s error: %s", d.Id(), err), resp)
			}
		}
	}

	// Wait for the update before reading
	time.Sleep(5 * time.Second)

	return nil
}

func isDefaultPlan(name string) bool {
	for _, defaultPlan := range defaultPlans {
		if name == defaultPlan {
			return true
		}
	}
	return false
}

// isNumberPlanInConfig returns true if the number plan's name is in the config list
func isNumberPlanInConfig(planName string, list []interface{}) bool {
	for _, plan := range list {
		planMap := plan.(map[string]interface{})
		if planName == planMap["name"] {
			return true
		}
	}
	return false
}

func readSiteNumberPlans(ctx context.Context, sp *SiteProxy, d *schema.ResourceData) *retry.RetryError {
	numberPlans, resp, err := sp.getSiteNumberPlans(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			d.SetId("") // Site doesn't exist
			return nil
		}
		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read number plans for site %s | error: %s", d.Id(), err), resp))
	}

	dNumberPlans := make([]interface{}, 0)
	if len(*numberPlans) > 0 {
		for _, numberPlan := range *numberPlans {
			dNumberPlan := flattenNumberPlan(&numberPlan)
			dNumberPlans = append(dNumberPlans, dNumberPlan)
		}
		_ = d.Set("number_plans", dNumberPlans)
	} else {
		_ = d.Set("number_plans", nil)
	}

	return nil
}

func readSiteOutboundRoutes(ctx context.Context, sp *SiteProxy, d *schema.ResourceData) *retry.RetryError {
	outboundRoutes, resp, err := sp.getSiteOutboundRoutes(ctx, d.Id())
	if err != nil {
		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to get outbound routes for site %s | error: %s", d.Id(), err), resp))
	}

	dOutboundRoutes := schema.NewSet(schema.HashResource(outboundRouteSchema), []interface{}{})

	if len(*outboundRoutes) > 0 {
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

			dOutboundRoutes.Add(dOutboundRoute)
		}
		_ = d.Set("outbound_routes", dOutboundRoutes)
	} else {
		_ = d.Set("outbound_routes", nil)
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

func flattenNumberPlan(numberPlan *platformclientv2.Numberplan) interface{} {
	dNumberPlan := make(map[string]interface{})
	dNumberPlan["name"] = *numberPlan.Name

	resourcedata.SetMapValueIfNotNil(dNumberPlan, "match_format", numberPlan.Match)
	resourcedata.SetMapValueIfNotNil(dNumberPlan, "normalized_format", numberPlan.NormalizedFormat)
	resourcedata.SetMapValueIfNotNil(dNumberPlan, "classification", numberPlan.Classification)
	resourcedata.SetMapValueIfNotNil(dNumberPlan, "match_type", numberPlan.MatchType)

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
	return dNumberPlan
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
				return nil, fmt.Errorf("failed to parse date %s: %s", startStr, err)
			}

			end, err := time.Parse("2006-01-02T15:04:05.000000", endStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse date %s: %s", end, err)
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

func CheckForDefaultSite(siteName string) error {
	var (
		sdk, _   = provider.AuthorizeSdk()
		siteAPI  = platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdk)
		pageSize = 100
	)

	sites, _, getErr := siteAPI.GetTelephonyProvidersEdgesSites(pageSize, 1, "", "", siteName, "", true, []string{})
	if getErr != nil {
		return getErr
	}
	if sites == nil {
		return fmt.Errorf("no default site found with name %s", siteName)
	}
	return nil
}

// DeleteLocationWithNumber is a test utility function to delete site and location with the provided emergency number
func DeleteLocationWithNumber(emergencyNumber string, config *platformclientv2.Configuration) error {
	var (
		locationsAPI = platformclientv2.NewLocationsApiWithConfig(config)
		pageCount    int
	)
	const pageSize = 100

	log.Printf("Reading locations")
	locations, _, getErr := locationsAPI.GetLocations(pageSize, 1, []string{}, "")
	if getErr != nil {
		return getErr
	}
	log.Printf("Read locations")
	if locations.Entities == nil || len(*locations.Entities) == 0 {
		return nil
	}

	pageCount = *locations.PageCount

	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		log.Printf("Reading locations")
		locations, _, getErr := locationsAPI.GetLocations(pageSize, pageNum, []string{}, "")
		if getErr != nil {
			return getErr
		}
		log.Printf("Read locations")

		if locations.Entities == nil || len(*locations.Entities) == 0 {
			break
		}

		for _, location := range *locations.Entities {
			if location.EmergencyNumber != nil {
				if location.EmergencyNumber.E164 == nil {
					continue
				}
				if strings.Contains(*location.EmergencyNumber.E164, emergencyNumber) {
					err := deleteSiteWithLocationId(*location.Id, config)
					if err != nil {
						return err
					}
					log.Printf("Deleting location %s", *location.Id)
					if _, err = locationsAPI.DeleteLocation(*location.Id); err != nil {
						return err
					}
					log.Printf("Deleted location %s", *location.Id)
					time.Sleep(30 * time.Second)
					return nil
				}
			}
		}
	}

	return nil
}

// deleteSiteWithLocationId is a test utility function that will
// delete a site with the provided location id
func deleteSiteWithLocationId(locationId string, config *platformclientv2.Configuration) error {
	const pageSize = 100
	var (
		pageCount int
		edgesAPI  = platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(config)
	)

	log.Printf("Reading telephony providers edges sites with location ID %s", locationId)
	sites, _, getErr := edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, 1, "", "", "", locationId, false, nil)
	if getErr != nil {
		return getErr
	}
	log.Printf("Read telephony providers edges sites with location ID %s", locationId)
	if sites.Entities == nil || len(*sites.Entities) == 0 {
		return nil
	}

	pageCount = *sites.PageCount

	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		log.Printf("Reading telephony providers edges site with location ID %s", locationId)
		sites, _, getErr = edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", "", locationId, false, nil)
		if getErr != nil {
			return getErr
		}
		log.Printf("Read telephony providers edges sites with location ID %s", locationId)

		if sites.Entities == nil || len(*sites.Entities) == 0 {
			return nil
		}

		for _, site := range *sites.Entities {
			if site.Location != nil && *site.Location.Id == locationId {
				log.Printf("Deleting telephony providers edges site %s", *site.Id)
				if _, err := edgesAPI.DeleteTelephonyProvidersEdgesSite(*site.Id); err != nil {
					return err
				}
				log.Printf("Deleted telephony providers edges site %s", *site.Id)
				time.Sleep(8 * time.Second)
			}
		}
	}
	return nil
}

// GetOrganizationDefaultSiteId is a test utiliy function to get the default site ID of the org
func GetOrganizationDefaultSiteId(config *platformclientv2.Configuration) (siteId string, err error) {
	organizationApi := platformclientv2.NewOrganizationApiWithConfig(config)

	org, _, err := organizationApi.GetOrganizationsMe()
	if err != nil {
		return "", err
	}
	if org.DefaultSiteId == nil {
		return "", nil
	}

	return *org.DefaultSiteId, nil
}
