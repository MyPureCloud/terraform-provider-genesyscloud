package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"strconv"
	location "terraform-provider-genesyscloud/genesyscloud/location"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony_provider_edges_trunkbasesettings"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSiteOutboundRoute(t *testing.T) {

	featureToggleCheck(t)

	t.Parallel()
	var (
		outboundRouteResourceLabel1 = "outbound_route_1"
		outboundRouteResourceLabel2 = "outbound_route_2"

		// site
		siteResourceLabel = "site"
		siteName          = "site " + uuid.NewString()
		description       = "terraform description 1"
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

	site := telephonyProvidersEdgesSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		siteName,
		description,
		"genesyscloud_location."+locationResourceLabel+".id",
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
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"outboundRoute name 1",
					"outboundRoute description",
					"\"International\"",
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
					"RANDOM",
					util.FalseValue) +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel2,
						"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings2.id",
						"SEQUENTIAL",
						util.FalseValue,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "name", "outboundRoute name 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel2, "name", "outboundRoute name 2"),
				),
				Destroy: false,
			},
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel+".id",
					"outboundRoute name 1",
					"outboundRoute description",
					"\"International\"",
					"genesyscloud_telephony_providers_edges_trunkbasesettings.trunkBaseSettings1.id",
					"RANDOM",
					util.FalseValue) + generateSiteOutboundRouteDataSource(
					outboundRouteResourceLabel1,
					"outboundRoute name 1",
					strconv.Quote(""),
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "site_id", "genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "route_id", "genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "route_id"),
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
	featureToggleCheck(t)

	var (
		dataResourceLabel = "managed-site-data"
		siteName          = "PureCloud Voice - AWS"
		name              = "Default Outbound Route"
	)

	siteId, err := getSiteIdByName(siteName)
	if err != nil {
		t.Skipf("failed to retrieve ID of site '%s'", name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateSiteOutboundRouteDataSource(
					dataResourceLabel,
					name,
					strconv.Quote(siteId),
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_site_outbound_route."+dataResourceLabel, "site_id", siteId),
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_site_outbound_route."+dataResourceLabel, "name", name),
				),
			},
		},
	})
}

func generateSiteOutboundRouteDataSource(
	dataSourceLabel string,
	name string,
	siteId string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_site_outbound_route" "%s" {
		name = "%s"
		site_id = %s
		depends_on=[%s]
	}
	`, dataSourceLabel, name, siteId, dependsOnResource)
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
