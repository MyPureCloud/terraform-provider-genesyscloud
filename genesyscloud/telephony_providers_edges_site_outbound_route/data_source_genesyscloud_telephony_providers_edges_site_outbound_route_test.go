package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"log"
	"os"
	"strconv"
	location "terraform-provider-genesyscloud/genesyscloud/location"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSiteOutboundRoute(t *testing.T) {

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

	t.Parallel()
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

	locationConfig := location.GenerateLocationResource(
		locationRes,
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
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateOutboundRoutesResource(
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
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "outbound_routes.1.name", "outboundRoute name 2"),
				),
				Destroy: false,
			},
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRouteDataSource(
					outboundRouteResource,
					name,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResource, "id", "genesyscloud_telephony_providers_edges_site."+siteRes, "id"),
				),
			},
		},
	})
}

/*
This test expects that the org has a product called "voice" enabled on it. If the test org does not have this product on it, the test can be skipped or ignored.
*/
func TestAccDataSourceSiteManaged(t *testing.T) {
	//t.Parallel()
	var (
		siteDataRes = "managed-site-data"
		name        = "PureCloud Voice - AWS"
	)

	siteId, err := getSiteIdByName(name)
	if err != nil {
		t.Skipf("failed to retrieve ID of site '%s'", name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateSiteOutboundRouteDataSource(
					siteDataRes,
					name,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_site_outbound_route."+siteDataRes, "id", siteId),
				),
			},
		},
	})
}

func generateSiteOutboundRouteDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_site_outbound_route" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}

func getSiteIdByName(name string) (string, error) {
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
	data, _, err := api.GetTelephonyProvidersEdgesSites(1, 1, "", "", name, "", true, nil)
	if err != nil {
		return "", err
	}
	if data.Entities == nil || len(*data.Entities) == 0 {
		return "", fmt.Errorf("no sites found with name %s", name)
	}
	site := (*data.Entities)[0]
	if *site.Name != name {
		return "", fmt.Errorf("no sites found with name %s", name)
	}
	return *site.Id, nil
}
