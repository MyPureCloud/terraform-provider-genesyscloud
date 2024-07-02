package telephony_providers_edges_site

import (
	"fmt"
	"os"
	"log"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/telephony"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceSite(t *testing.T) {
	t.Parallel()

	var (
		// site
		siteRes      = "site"
		name1        = "site " + uuid.NewString()
		name2        = "site " + uuid.NewString()
		description1 = "TestAccResourceSite description 1"
		description2 = "TestAccResourceSite description 2"
		mediaModel   = "Cloud"

		// edge_auto_update_config
		timeZone = "America/New_York"
		rrule    = "FREQ=WEEKLY;BYDAY=SU"
		start1   = "2021-08-08T08:00:00.000000"
		start2   = "2021-08-15T08:00:00.000000"
		end1     = "2021-08-08T11:00:00.000000"
		end2     = "2021-08-15T11:00:00.000000"
		// location
		locationRes = "test-location1"
	)

	emergencyNumber := "+13173124741"
	if err := DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	location := gcloud.GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name1,
					description1,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_model", mediaModel),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_regions_use_latency_based", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "location_id", "genesyscloud_location."+locationRes, "id"),
				),
			},
			// Update description, name and media_regions_use_latency_based
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name2,
					description2,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					true,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_model", mediaModel),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_regions_use_latency_based", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "location_id", "genesyscloud_location."+locationRes, "id"),
				),
			},
			// Update with EdgeAutoUpdateConfig
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name2,
					description2,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					true,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteEdgeAutoUpdateConfig(
						timeZone,
						rrule,
						start1,
						end1)) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.time_zone", timeZone),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.rrule", rrule),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.start", start1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.end", end1),
				),
			},
			// Update the EdgeAutoUpdateConfig
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name2,
					description2,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					true,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteEdgeAutoUpdateConfig(
						timeZone,
						rrule,
						start2,
						end2)) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.time_zone", timeZone),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.rrule", rrule),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.start", start2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "edge_auto_update_config.0.end", end2),
				),
			},
		},
		CheckDestroy: testVerifySitesDestroyed,
	})
}

func TestAccResourceSiteoutboundRoute(t *testing.T) {
	if exists := featureToggles.OutboundRoutesToggleExists(); exists {
		// Unset outbound routes feature toggle so outbound routes will be managed by the site resource for this test
		err := os.Unsetenv(featureToggles.OutboundRoutesToggleName())
		if err != nil {
			t.Skipf("%v is set and unable to unset, skipping test", featureToggles.OutboundRoutesToggleName())
		}
	}
	var (
		// site
		siteRes     = "site"
		name        = "site " + uuid.NewString()
		description = "terraform description 1"
		mediaModel  = "Cloud"

		// location
		locationRes = "test-location1"
	)

	emergencyNumber := "+13173124743"
	if err := DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s, %v", emergencyNumber, err)
	}

	location := gcloud.GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	trunkBaseSettings1 := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings1",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings2 := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings2",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings3 := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings3",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteOutboundRoutesWithCustomAttrs(
						"outboundRoute name 1",
						"outboundRoute description",
						"\"International\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
						"RANDOM",
						false),
					generateSiteOutboundRoutesWithCustomAttrs(
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						false)+"set_as_default_site = false") + trunkBaseSettings1 + trunkBaseSettings2 + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.enabled", util.FalseValue),
				),
			},
			// Switch around the order of outbound routes which shouldn't have any effect
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteOutboundRoutesWithCustomAttrs(
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						false),
					generateSiteOutboundRoutesWithCustomAttrs(
						"outboundRoute name 1",
						"outboundRoute description",
						"\"International\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
						"RANDOM",
						false)) + trunkBaseSettings1 + trunkBaseSettings2 + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.enabled", util.FalseValue),
				),
			},
			// Remove a route and update the description, classification types, trunk base ids, distribution and enabled value of another route
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteOutboundRoutesWithCustomAttrs(
						"outboundRoute name 1",
						"outboundRoute description updated",
						strings.Join([]string{strconv.Quote("Network"), strconv.Quote("International")}, ","),
						strings.Join([]string{"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3.id"}, ","),
						"RANDOM",
						true)+"set_as_default_site = false") + trunkBaseSettings1 + trunkBaseSettings2 + trunkBaseSettings3 + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.description", "outboundRoute description updated"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.1", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.external_trunk_base_ids.1", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3", "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_site." + siteRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySitesDestroyed,
	})
}
func TestAccResourceSiteNumberPlans(t *testing.T) {
	t.Parallel()
	var (
		// site
		siteRes     = "site"
		name        = "site " + uuid.NewString()
		description = "TestAccResourceSiteNumberPlans description 1"
		mediaModel  = "Cloud"

		// location
		locationRes = "test-location1"
	)

	emergencyNumber := "+13173124742"
	if err := DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	location := gcloud.GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteNumberPlansWithCustomAttrs(
						"numberList name",
						"numberList classification",
						"",
						"numberList",
						"",
						generateSiteNumberPlansNumber("112", "113")),
					generateSiteNumberPlansWithCustomAttrs(
						"digitLength name",
						"digitLength classification",
						"",
						"digitLength",
						"",
						generateSiteNumberPlansDigitLength("4", "6")),
					generateSiteNumberPlansWithCustomAttrs(
						"intraCountryCode name",
						"intraCountryCode classification",
						"",
						"intraCountryCode",
						""),
					generateSiteNumberPlansWithCustomAttrs(
						"interCountryCode name",
						"interCountryCode classification",
						"",
						"interCountryCode",
						""),
					generateSiteNumberPlansWithCustomAttrs(
						"regex name",
						"regex classification",
						"^([^@\\\\:]+@)([^@ ]+)?$",
						"regex",
						"sip:$1$2")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.name", "numberList name"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.classification", "numberList classification"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.match_type", "numberList"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.numbers.0.start", "112"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.numbers.0.end", "113"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.4.name", "regex name"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.4.classification", "regex classification"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.4.match_type", "regex"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.4.match_format", "^([^@\\:]+@)([^@ ]+)?$"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.4.normalized_format", "sip:$1$2"),
				),
			},
			// Remove 2 number plans and update the properties of others
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteNumberPlansWithCustomAttrs(
						"numberList name",
						"numberList classification",
						"",
						"numberList",
						"",
						generateSiteNumberPlansNumber("114", "115")),
					generateSiteNumberPlansWithCustomAttrs(
						"digitLength name",
						"digitLength classification",
						"",
						"digitLength",
						"",
						generateSiteNumberPlansDigitLength("6", "8")),
					generateSiteNumberPlansWithCustomAttrs(
						"regex name",
						"regex classification",
						"^([^@\\\\:]+@)([^@ ]+)?$",
						"regex",
						"sip:$2$3")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.numbers.0.start", "114"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.numbers.0.end", "115"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.1.digit_length.0.start", "6"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.1.digit_length.0.end", "8"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.2.match_format", "^([^@\\:]+@)([^@ ]+)?$"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.2.normalized_format", "sip:$2$3"),
				),
			},
			// Add one plan back in
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteNumberPlansWithCustomAttrs(
						"numberList name",
						"numberList classification",
						"",
						"numberList",
						"",
						generateSiteNumberPlansNumber("114", "115")),
					generateSiteNumberPlansWithCustomAttrs(
						"digitLength name",
						"digitLength classification",
						"",
						"digitLength",
						"",
						generateSiteNumberPlansDigitLength("6", "8")),
					generateSiteNumberPlansWithCustomAttrs(
						"interCountryCode name",
						"interCountryCode classification",
						"",
						"interCountryCode",
						""),
					generateSiteNumberPlansWithCustomAttrs(
						"regex name",
						"regex classification",
						"^([^@\\\\:]+@)([^@ ]+)?$",
						"regex",
						"sip:$2$3")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.numbers.0.start", "114"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.0.numbers.0.end", "115"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.1.digit_length.0.start", "6"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.1.digit_length.0.end", "8"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.2.name", "interCountryCode name"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.2.classification", "interCountryCode classification"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.3.match_format", "^([^@\\:]+@)([^@ ]+)?$"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "number_plans.3.normalized_format", "sip:$2$3"),
				),
			},
		},
		CheckDestroy: testVerifySitesDestroyed,
	})
}

func TestAccResourceSiteDefaultSite(t *testing.T) {
	var (
		// site
		siteRes      = "site"
		name1        = "site " + uuid.NewString()
		description1 = "TestAccResourceSite description 1"
		mediaModel   = "Cloud"

		// location
		locationRes = "test-location1"
	)

	originalSiteId, err := GetOrganizationDefaultSiteId(sdkConfig)
	if err != nil {
		t.Fatal(err)
	}

	emergencyNumber := "+13173124744"
	if err = DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s, %v", emergencyNumber, err)
	}

	location := gcloud.GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Store the original default site, so it can be restored later
				PreConfig: func() {
					originalSiteId, err = GetOrganizationDefaultSiteId(sdkConfig)
					if err != nil {
						t.Fatalf("error setting original default site ID %s", originalSiteId)
					}
				},
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name1,
					description1,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					"set_as_default_site = true") + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_model", mediaModel),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_regions_use_latency_based", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "location_id", "genesyscloud_location."+locationRes, "id"),
					testDefaultSite("genesyscloud_telephony_providers_edges_site."+siteRes),
				),
			},
			{
				// Restore the old default site before cleaning up after the test.
				PreConfig: func() {
					if err := setDefaultSite(originalSiteId); err != nil {
						t.Fatalf("cannot restore default site back to %s", originalSiteId)
					}
					time.Sleep(5 * time.Second) // Wait or test case will error trying to delete the created default site
				},
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name1,
					description1,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					"set_as_default_site = false") + location + gcloud.GenerateOrganizationMe(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_organizations_me.me", "default_site_id", originalSiteId),
				),
			},
		},
		CheckDestroy: testVerifySitesDestroyed,
	})
}

func testVerifySitesDestroyed(state *terraform.State) error {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_site" {
			continue
		}

		site, resp, err := edgesAPI.GetTelephonyProvidersEdgesSite(rs.Primary.ID)
		if site != nil {
			if site.State != nil && *site.State == "deleted" {
				// site deleted
				continue
			}
			return fmt.Errorf("site (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// site not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All sites destroyed
	return nil
}

func generateSiteEdgeAutoUpdateConfig(timeZone, rrule, start, end string) string {
	return fmt.Sprintf(`edge_auto_update_config {
		time_zone = "%s"
        rrule = "%s"
        start = "%s"
        end = "%s"
	}
	`, timeZone, rrule, start, end)
}

func generateSiteNumberPlansWithCustomAttrs(
	name,
	classification,
	matchFormat,
	matchType,
	normalizedFormat string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`number_plans {
		name = "%s"
		classification = "%s"
		match_format = "%s"
		match_type = "%s"
		normalized_format = "%s"
		%s
	}
	`,
		name,
		classification,
		matchFormat,
		matchType,
		normalizedFormat,
		strings.Join(otherAttrs, "\n"))
}

func generateSiteNumberPlansNumber(start, end string) string {
	return fmt.Sprintf(`numbers {
        start = "%s"
        end = "%s"
	}
	`, start, end)
}

func generateSiteNumberPlansDigitLength(start, end string) string {
	return fmt.Sprintf(`digit_length {
        start = "%s"
        end = "%s"
	}
	`, start, end)
}

func generateSiteOutboundRoutesWithCustomAttrs(
	name,
	description,
	classificationTypes,
	externalTrunkBaseIds,
	distribution string,
	enabled bool,
	otherAttrs ...string) string {
	return fmt.Sprintf(`outbound_routes {
		name = "%s"
		description = "%s"
		classification_types = [%s]
		external_trunk_base_ids = [%s]
		distribution = "%s"
		enabled = %v
		%s
	}
	`,
		name,
		description,
		classificationTypes,
		externalTrunkBaseIds,
		distribution,
		enabled,
		strings.Join(otherAttrs, "\n"))
}

// getOrganizationDefaultSite is a test utiliy function to set the default site of the org
func setDefaultSite(siteId string) error {
	sdkConfig := platformclientv2.GetDefaultConfiguration()
	organizationApi := platformclientv2.NewOrganizationApiWithConfig(sdkConfig)

	org, _, err := organizationApi.GetOrganizationsMe()
	if err != nil {
		return err
	}

	// Update org details
	*org.DefaultSiteId = siteId

	_, _, err = organizationApi.PutOrganizationsMe(*org)
	if err != nil {
		return err
	}

	log.Printf("set default site to %s", siteId)

	return nil
}

// Verify if the provided resource site is the default site
func testDefaultSite(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		defaultSiteId, err := GetOrganizationDefaultSiteId(sdkConfig)
		if err != nil {
			return fmt.Errorf("failed to get default site id: %v", err)
		}

		r := state.RootModule().Resources[resource]
		if r == nil {
			return fmt.Errorf("%s not found in state", resource)
		}

		if r.Primary.ID != defaultSiteId {
			return fmt.Errorf("default site is expected to be %s. Instead got %s", r.Primary.ID, defaultSiteId)
		}

		return nil
	}
}
