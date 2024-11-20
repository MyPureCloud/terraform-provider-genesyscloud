package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/location"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony_provider_edges_trunkbasesettings"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSiteoutboundRoutes(t *testing.T) {

	featureToggleCheck(t)

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

	trunkBaseSettings1 := telephony_provider_edges_trunkbasesettings.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings1",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings2 := telephony_provider_edges_trunkbasesettings.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		"trunkBaseSettings2",
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)

	trunkBaseSettings3 := telephony_provider_edges_trunkbasesettings.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
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

func featureToggleCheck(t *testing.T) {
	featureEnvSet := os.Getenv(featureToggles.OutboundRoutesToggleName())
	if featureEnvSet == "" {
		err := os.Setenv(featureToggles.OutboundRoutesToggleName(), "enabled")
		if err != nil {
			t.Errorf("%s is not set", featureToggles.OutboundRoutesToggleName())
		}
		defer func() {
			err := os.Unsetenv(featureToggles.OutboundRoutesToggleName())
			if err != nil {
				log.Printf("%s", err)
			}
		}()
	}
}
