package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	telephonyProvidersEdgesSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	tbs "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunkbasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSiteOutboundRoute(t *testing.T) {
	t.Parallel()
	var (
		outboundRouteResourceLabel1 = "outbound_route_1"
		outboundRouteResourceLabel2 = "outbound_route_2"

		// site
		siteResourceLabel = "site"
		siteName          = "tf test site " + uuid.NewString()
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

	trunkBaseSettings1Label := "trunkBaseSettings1"
	trunkBaseSettings1 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		trunkBaseSettings1Label,
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)
	trunkBaseSettings1ResourceFullPath := tbs.ResourceType + "." + trunkBaseSettings1Label

	trunkBaseSettings2Label := "trunkBaseSettings2"
	trunkBaseSettings2 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		trunkBaseSettings2Label,
		"test trunk base settings "+uuid.NewString(),
		"test description",
		"external_sip.json",
		"EXTERNAL",
		false)
	trunkBaseSettings2ResourceFullPath := tbs.ResourceType + "." + trunkBaseSettings2Label

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
	siteResourceFullPath := telephonyProvidersEdgesSite.ResourceType + "." + siteResourceLabel

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: trunkBaseSettings1 + trunkBaseSettings2 + locationConfig + site + generateSiteOutboundRoutesResource(
					outboundRouteResourceLabel1,
					siteResourceFullPath+".id",
					"outboundRoute name 1",
					"outboundRoute description",
					strconv.Quote("International"),
					trunkBaseSettings1ResourceFullPath+".id",
					"RANDOM",
					util.FalseValue) +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel2,
						siteResourceFullPath+".id",
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						trunkBaseSettings2ResourceFullPath+".id",
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
					siteResourceFullPath+".id",
					"outboundRoute name 1",
					"outboundRoute description",
					strconv.Quote("International"),
					trunkBaseSettings1ResourceFullPath+".id",
					"RANDOM",
					util.FalseValue) +
					generateSiteOutboundRoutesResource(
						outboundRouteResourceLabel2,
						siteResourceFullPath+".id",
						"outboundRoute name 2",
						"outboundRoute description",
						"\"National\"",
						trunkBaseSettings2ResourceFullPath+".id",
						"SEQUENTIAL",
						util.FalseValue,
					) +
					generateSiteOutboundRouteDataSource(
						outboundRouteResourceLabel1,
						"outboundRoute name 1",
						siteResourceFullPath+".id",
						"",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "site_id", siteResourceFullPath, "id"),
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "route_id",
						"genesyscloud_telephony_providers_edges_site_outbound_route."+outboundRouteResourceLabel1, "route_id"),
				),
			},
		},
	})
}

/*
This test expects that the org has a product called "voice" enabled on it. If the test org does not have this product on it, the test can be skipped or ignored.
*/
func TestAccDataSourceSiteManaged(t *testing.T) {
	var (
		dataResourceLabel    = "managed-site-data"
		dataResourceFullPath = "data." + ResourceType + "." + dataResourceLabel
		siteName             = "PureCloud Voice - AWS"
		name                 = "Default Outbound Route"
	)

	siteId, err := getSiteIdByName(siteName)
	if err != nil {
		t.Skipf("failed to retrieve ID of site '%s'", siteName)
	}

	// verify site has outbound routes associated with it
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
	data, _, err := api.GetTelephonyProvidersEdgesSiteOutboundroutes(siteId, 100, 1, name, "", "")
	if err != nil {
		t.Skipf("failed to read outbound routes named '%s' for site '%s'", name, siteName)
	}
	if data == nil || data.Entities == nil || len(*data.Entities) == 0 {
		t.Skipf("no outbound routes named '%s' found for site '%s'", name, siteName)
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
					resource.TestCheckResourceAttr(dataResourceFullPath, "site_id", siteId),
					resource.TestCheckResourceAttr(dataResourceFullPath, "name", name),
				),
			},
		},
	})
}

func generateSiteOutboundRouteDataSource(
	dataSourceLabel,
	name,
	siteId,
	dependsOnResource string,
) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name       = "%s"
		site_id    = %s
		depends_on = [%s]
	}
	`, ResourceType, dataSourceLabel, name, siteId, dependsOnResource)
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
