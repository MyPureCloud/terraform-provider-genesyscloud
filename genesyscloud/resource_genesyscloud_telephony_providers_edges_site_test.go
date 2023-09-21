package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
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

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	emergencyNumber := "+13173124740"
	err = DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	location := GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		GenerateLocationEmergencyNum(
			emergencyNumber,
			nullValue, // Default number type
		), GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name1,
					description1,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					"[\"us-west-2\"]",
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_model", mediaModel),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_regions_use_latency_based", falseValue),
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
					"[\"us-west-2\"]",
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing")) + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_model", mediaModel),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "media_regions_use_latency_based", trueValue),
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
					"[\"us-west-2\"]",
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
					"[\"us-west-2\"]",
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

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	emergencyNumber := "+13173124741"
	err = DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	location := GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		GenerateLocationEmergencyNum(
			emergencyNumber,
			nullValue, // Default number type
		), GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					"[\"us-west-2\"]",
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
					"[\"us-west-2\"]",
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
					"[\"us-west-2\"]",
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

func TestAccResourceSiteOutboundRoutes(t *testing.T) {
	t.Parallel()
	var (
		// site
		siteRes     = "site"
		name        = "site " + uuid.NewString()
		description = "terraform description 1"
		mediaModel  = "Cloud"

		// location
		locationRes = "test-location1"
	)

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	emergencyNumber := "+13173124742"
	err = DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	location := GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		GenerateLocationEmergencyNum(
			emergencyNumber,
			nullValue, // Default number type
		), GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	trunkBaseSettings1 := generateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings1",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings2 := generateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings2",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings3 := generateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings3",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					"[\"us-west-2\"]",
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
						false)) + trunkBaseSettings1 + trunkBaseSettings2 + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.enabled", falseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.1.enabled", falseValue),
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
					"[\"us-west-2\"]",
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing"),
					generateSiteOutboundRoutesWithCustomAttrs(
						"outboundRoute name 1",
						"outboundRoute description updated",
						strings.Join([]string{strconv.Quote("Network"), strconv.Quote("International")}, ","),
						strings.Join([]string{"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3.id"}, ","),
						"RANDOM",
						true)) + trunkBaseSettings1 + trunkBaseSettings2 + trunkBaseSettings3 + location,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.description", "outboundRoute description updated"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.classification_types.1", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site."+siteRes, "outbound_routes.0.enabled", trueValue),
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
		} else if IsStatus404(resp) {
			// site not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
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
