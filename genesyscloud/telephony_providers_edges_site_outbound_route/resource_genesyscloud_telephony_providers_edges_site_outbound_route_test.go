package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

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

	emergencyNumber := "+13173124743"
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
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
				Config: telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
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
				Config: telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
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
	})
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
