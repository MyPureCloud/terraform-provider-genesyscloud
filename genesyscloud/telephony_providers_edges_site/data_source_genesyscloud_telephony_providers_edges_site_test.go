package telephony_providers_edges_site

import (
	"fmt"
	"strconv"
	location "terraform-provider-genesyscloud/genesyscloud/location"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSite(t *testing.T) {
	t.Parallel()
	var (
		// site
		siteResourceLabel     = "site"
		siteDataResourceLabel = "site-data"
		name                  = "tf-site-" + uuid.NewString()
		description1          = "test site description"
		mediaModel            = "Cloud"

		// location
		locationResourceLabel = "test-location1"
	)

	emergencyNumber := "+13173124745"
	if err := DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSiteResourceWithCustomAttrs(
					siteResourceLabel,
					name,
					description1,
					"genesyscloud_location."+locationResourceLabel+".id",
					mediaModel,
					false,
					util.AssignRegion(),
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing")) + locationConfig + generateSiteDataSource(
					siteDataResourceLabel,
					name,
					"genesyscloud_telephony_providers_edges_site."+siteResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site."+siteDataResourceLabel, "id", "genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
				),
			},
		},
	})
}

/*
This test expects that the org has a product called "voice" enabled on it. If the test org does not have this product on it, the test can be skipped or ignored.
*/
func TestAccDataSourceSiteManaged(t *testing.T) {
	t.Parallel()
	var (
		siteDataResourceLabel = "managed-site-data"
		name                  = "PureCloud Voice - AWS"
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
				Config: generateSiteDataSource(
					siteDataResourceLabel,
					name,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_site."+siteDataResourceLabel, "id", siteId),
				),
			},
		},
	})
}

func generateSiteDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_site" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
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
