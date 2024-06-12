package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSiteoutboundRoutes(t *testing.T) {
	defer func() {
		err := os.Unsetenv(featureToggles.OutboundRoutesToggleName())
		if err != nil {
			log.Printf("%s", err)
		}
	}()

	err := os.Setenv(featureToggles.OutboundRoutesToggleName(), "enabled")
	if err != nil {
		t.Errorf("%s is not set", featureToggles.OutboundRoutesToggleName())
	}

	var (
		outboundRouteResource = "outbound_route"

		// site
		siteRes     = "site"
		name        = "site " + uuid.NewString()
		description = "terraform description 1"
		mediaModel  = "Cloud"

		// location
		locationRes = "test-location1"
	)

	emergencyNumber := "+13173124741"
	if err := telephonyProvidersEdgesSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
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

	site := telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
		siteRes,
		name,
		description,
		"genesyscloud_location."+locationRes+".id",
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
				Config: trunkBaseSettings1 + trunkBaseSettings2 + location + site + generateOutboundRoutesResource(
					outboundRouteResource,
					"genesyscloud_telephony_providers_edges_site."+siteRes+".id",
					generateSiteOutboundRoutes(
						"outboundRoute name 1",
						"outboundRoute description",
						"\"International\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
						"RANDOM",
						util.FalseValue),
					generateSiteOutboundRoutes(
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						util.FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.enabled", util.FalseValue),
				),
			},
			// Switch around the order of outbound routes which shouldn't have any effect
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + location + site + generateOutboundRoutesResource(
					outboundRouteResource,
					"genesyscloud_telephony_providers_edges_site."+siteRes+".id",
					generateSiteOutboundRoutes(
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						util.FalseValue),
					generateSiteOutboundRoutes(
						"outboundRoute name 1",
						"outboundRoute description",
						"\"International\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
						"RANDOM",
						util.FalseValue)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.classification_types.0", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.enabled", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.name", "outboundRoute name 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.description", "outboundRoute description"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.classification_types.0", "National"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.distribution", "SEQUENTIAL"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2", "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.enabled", util.FalseValue),
				),
			},
			// Remove a route and update the description, classification types, trunk base ids, distribution and enabled value of another route
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + trunkBaseSettings3 + location + site + generateOutboundRoutesResource(
					outboundRouteResource,
					"genesyscloud_telephony_providers_edges_site."+siteRes+".id",
					generateSiteOutboundRoutes(
						"outboundRoute name 1",
						"outboundRoute description updated",
						strings.Join([]string{strconv.Quote("Network"), strconv.Quote("International")}, ","),
						strings.Join([]string{"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3.id"}, ","),
						"RANDOM",
						util.TrueValue)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.description", "outboundRoute description updated"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.classification_types.0", "Network"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.classification_types.1", "International"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.distribution", "RANDOM"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.external_trunk_base_ids.0", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1", "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.0.external_trunk_base_ids.1", "genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings3", "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_site_outbound_route." + outboundRouteResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateOutboundRoutesResource(routesResource, siteId string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_site_outbound_route" "%s" {
		site_id = %s
		%s
	}`, routesResource, siteId, strings.Join(nestedBlocks, "\n"))
}

func generateSiteOutboundRoutes(
	name,
	description,
	classificationTypes,
	externalTrunkBaseIds,
	distribution,
	enabled string) string {
	return fmt.Sprintf(`outbound_routes {
		name = "%s"
		description = "%s"
		classification_types = [%s]
		external_trunk_base_ids = [%s]
		distribution = "%s"
		enabled = %s
	}
	`, name, description, classificationTypes, externalTrunkBaseIds, distribution, enabled)
}
