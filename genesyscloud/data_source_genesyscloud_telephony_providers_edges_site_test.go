package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSite(t *testing.T) {
	t.Parallel()
	var (
		// site
		siteRes      = "site"
		siteDataRes  = "site-data"
		name         = "site " + uuid.NewString()
		description1 = "test site description"
		mediaModel   = "Cloud"

		// location
		locationRes = "test-location1"
	)

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	emergencyNumber := "3173124744"
	err = deleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	location := generateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		generateLocationEmergencyNum(
			emergencyNumber,
			nullValue, // Default number type
		), generateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateSiteResourceWithCustomAttrs(
					siteRes,
					name,
					description1,
					"genesyscloud_location."+locationRes+".id",
					mediaModel,
					false,
					"[\"us-west-2\"]",
					"+19205551212",
					"Wilco plumbing") + location + generateSiteDataSource(
					siteDataRes,
					name,
					"genesyscloud_telephony_providers_edges_site."+siteRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_site."+siteDataRes, "id", "genesyscloud_telephony_providers_edges_site."+siteRes, "id"),
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
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_site" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
