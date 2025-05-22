package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	telephonyProvidersEdgesSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	tbs "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunkbasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSiteoutboundRoutes(t *testing.T) {
	var (
		outboundRouteResourceLabel1 = "outbound_route_1"
		outboundRouteResourceLabel2 = "outbound_route_2"

		// site
		siteResourceLabel = "site"
		siteName          = "tf test site " + uuid.NewString()
		siteDescription   = "terraform description 1"
		mediaModel        = "Cloud"

		// location
		locationResourceLabel = "test-location1"
	)

	emergencyNumber := "+13173124741"
	if err := telephonyProvidersEdgesSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s, %v", emergencyNumber, err)
	}

	locationConfig := location.GenerateLocationResource(
		locationResourceLabel,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), location.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	trunkBaseSettings1 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings1",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings2 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings2",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings3 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings3",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	site := telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		siteName,
		siteDescription,
		"genesyscloud_location."+locationResourceLabel+".id",
		mediaModel,
		false,
		util.AssignRegion(),
		strconv.Quote("+19205551212"),
		strconv.Quote("Wilco plumbing"),
		"set_as_default_site = false")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"outboundRoute name 1",
					"outboundRoute description 1",
					strings.Join([]string{strconv.Quote("National"), strconv.Quote("International")}, ","),
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
					"RANDOM",
					util.FalseValue) +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel2,
						"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
						"outboundRoute name 2",
						"outboundRoute description 2",
						"\"Network\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						util.FalseValue,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.1", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "description", "outboundRoute description 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "enabled", util.FalseValue),
				),
			},
			// Switch around the order of outbound routes which shouldn't have any effect
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel2,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"outboundRoute name 2",
					"outboundRoute description 2",
					"\"Network\"",
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
					"SEQUENTIAL",
					util.FalseValue) +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel1,
						"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
						"outboundRoute name 1",
						"outboundRoute description 1",
						"\"International\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
						"RANDOM",
						util.FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "description", "outboundRoute description 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "enabled", util.FalseValue),
				),
			},
			// Remove a route and update the description, classification types, trunk base ids, distribution and enabled value of another route
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + trunkBaseSettings3 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"outboundRoute name 1",
					"outboundRoute description updated",
					"\"International\"",
					strings.Join([]string{"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3.id"}, ","),
					"RANDOM",
					util.TrueValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description updated"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.1", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3", "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_site_outbound_route." + outboundRouteResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSiteoutboundRoutesDefaultOutboundRoute(t *testing.T) {
	var (
		outboundRouteResourceLabel1 = "outbound_route_1"
		outboundRouteResourceLabel2 = "outbound_route_2"

		// site
		siteResourceLabel = "site"
		siteName          = "site " + uuid.NewString()
		siteDescription   = "terraform description 1"
		mediaModel        = "Cloud"

		// location
		locationResourceLabel = "test-location1"
	)

	emergencyNumber := "+13173124741"
	if err := telephonyProvidersEdgesSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s, %v", emergencyNumber, err)
	}

	locationConfig := location.GenerateLocationResource(
		locationResourceLabel,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), location.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	trunkBaseSettings1 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings1",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings2 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings2",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	site := telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		siteName,
		siteDescription,
		"genesyscloud_location."+locationResourceLabel+".id",
		mediaModel,
		false,
		util.AssignRegion(),
		strconv.Quote("+19205551212"),
		strconv.Quote("Wilco plumbing"),
		"set_as_default_site = false")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Check Default Outbound Route gets created when a Site is created
				Config: trunkBaseSettings1 + locationConfig + site + generateSiteOutboundRouteDataSource(
					outboundRouteResourceLabel1,
					"Default Outbound Route",
					testrunner.GenerateFullPathId(telephonyProvidersEdgesSite.ResourceType, siteResourceLabel),
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "site_id", "genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
				),
			},
			// Check that the Default Outbound Route can be updated
			{
				Config: trunkBaseSettings1 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"Default Outbound Route",
					"outboundRoute description 1",
					strings.Join([]string{strconv.Quote("National"), strconv.Quote("International")}, ","),
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
					"RANDOM",
					util.FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "Default Outbound Route"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.1", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.FalseValue),
				),
			},
			// Check that a new outbound route can be added to the site
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"Default Outbound Route",
					"outboundRoute description 1",
					"\"International\"",
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
					"RANDOM",
					util.FalseValue) + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel2,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"outboundRoute name 2",
					"outboundRoute description 2",
					"\"Network\"",
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
					"SEQUENTIAL",
					util.FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "Default Outbound Route"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "description", "outboundRoute description 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "enabled", util.FalseValue),
				),
			},
			// Remove a route and update the description, classification types, trunk base ids, distribution and enabled value of another route
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"Default Outbound Route",
					"outboundRoute description updated",
					"\"International\"",
					strings.Join([]string{"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id"}, ","),
					"RANDOM",
					util.TrueValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "Default Outbound Route"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description updated"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.1", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_site_outbound_route." + outboundRouteResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Test to confirm that upon destroy if the site has already been removed, the site outbound route resource can be cleanly deleted too
func TestAccResourceSiteoutboundRoutesWhenSiteResourceIsDeleted(t *testing.T) {
	var (
		outboundRouteResourceLabel1 = "outbound_route_1"
		outboundRouteResourceLabel2 = "outbound_route_2"

		// site
		siteResourceLabel = "site"
		siteName          = "site " + uuid.NewString()
		siteDescription   = "terraform description 1"
		mediaModel        = "Cloud"

		// location
		locationResourceLabel = "test-location1"
	)

	emergencyNumber := "+13173124741"
	if err := telephonyProvidersEdgesSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s, %v", emergencyNumber, err)
	}

	locationConfig := location.GenerateLocationResource(
		locationResourceLabel,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), location.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	trunkBaseSettings1 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings1",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings2 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings2",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	site := telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		siteName,
		siteDescription,
		"genesyscloud_location."+locationResourceLabel+".id",
		mediaModel,
		false,
		util.AssignRegion(),
		strconv.Quote("+19205551212"),
		strconv.Quote("Wilco plumbing"),
		"set_as_default_site = false")
	var siteId string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			// Step 1: Create and verify site
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
					func(state *terraform.State) error {
						// Get the site ID from the terraform state for the next step
						siteResource := state.RootModule().Resources["genesyscloud_telephony_providers_edges_site."+siteResourceLabel]
						if siteResource == nil {
							log.Printf("Unable to find site resource in state")
							return nil
						}
						siteId = siteResource.Primary.ID
						return nil
					},
				),
			},
			// Step 2: Create outbound routes using the site
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel1,
						"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
						"outboundRoute name 1",
						"outboundRoute description 1",
						strings.Join([]string{strconv.Quote("National"), strconv.Quote("International")}, ","),
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
						"RANDOM",
						util.FalseValue) +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel2,
						"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
						"outboundRoute name 2",
						"outboundRoute description 2",
						"\"Network\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						util.FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "description", "outboundRoute description 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "classification_types.1", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "description", "outboundRoute description 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "enabled", util.FalseValue),
				),
			},
			// Step 3: Delete site on the API and verify outbound route resources can be destroyed gracefully when no site exists
			{
				PreConfig: func() {
					if siteId == "" {
						log.Printf("Unable to find site ID")
						return
					}
					ctx := context.Background()
					proxy := telephonyProvidersEdgesSite.GetSiteProxy(internalProxy.clientConfig)
					_, err := proxy.DeleteSite(ctx, siteId)
					if err != nil {
						log.Printf("Unable to delete site with ID %s, %v", siteId, err)
					}
				},
				Config: " ", // Empty config to trigger destroy (the single whitespace is necessary for the test harness)
				Check: resource.ComposeTestCheckFunc(
					util.TestCheckNoResourceInState("genesyscloud_telephony_providers_edges_site."+siteResourceLabel),
					util.TestCheckNoResourceInState("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1),
					util.TestCheckNoResourceInState("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2),
				),
			},
		},
	})
}

func generateSiteOutboundRoutesResource(
	routesResourceLabel,
	siteId string,
	name,
	description,
	classificationTypes,
	externalTrunkBaseIds,
	distribution,
	enabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_site_outbound_route" "%s" {
		site_id = %s
		name = "%s"
		description = "%s"
		classification_types = [%s]
		external_trunk_base_ids = [%s]
		distribution = "%s"
		enabled = %s
	}
	`, routesResourceLabel, siteId, name, description, classificationTypes, externalTrunkBaseIds, distribution, enabled)
}
