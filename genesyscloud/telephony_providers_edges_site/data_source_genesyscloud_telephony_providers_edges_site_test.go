package telephony_providers_edges_site

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSite(t *testing.T) {
	t.Parallel()
	var (
		// site
		siteRes      = "site"
		siteDataRes  = "site-data"
		name         = "tf-site-" + uuid.NewString()
		description1 = "test site description"
		mediaModel   = "Cloud"

		// location
		locationRes = "test-location1"
	)

	emergencyNumber := "+13173124745"
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
					description1,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					"[\"us-west-2\"]",
					strconv.Quote("+19205551212"),
					strconv.Quote("Wilco plumbing")) + location + generateSiteDataSource(
					siteDataRes,
					name,
					"genesyscloud_telephony_providers_edges_site."+siteRes,
					false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site."+siteDataRes, "id", "genesyscloud_telephony_providers_edges_site."+siteRes, "id"),
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
				Config: generateSiteDataSource(
					siteDataRes,
					name,
					"",
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_site."+siteDataRes, "id", siteId),
				),
			},
		},
	})
}

func generateSiteDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
	managed bool) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_site" "%s" {
		name = "%s"
		managed = %t
		depends_on=[%s]
	}
	`, resourceID, name, managed, dependsOnResource)
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
