package location

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLocation(t *testing.T) {
	var (
		locResourceLabel = "test-location-members"
		locDataLabel     = "location-data"
		locName          = "test-location"

		locNotes = "HQ1"
		street   = "7601 Interactive Way"
		city     = "Indianapolis"
		state    = "IN"
		country  = "US"
		zip      = "46278"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateLocationResource(
					locResourceLabel,
					locName,
					locNotes,
					[]string{}, // no paths or emergency number
					GenerateLocationAddress(street, city, state, country, zip),
				) + generateLocationDataSource(
					locDataLabel,
					locName,
					"genesyscloud_location."+locResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_location."+locDataLabel, "id", "genesyscloud_location."+locResourceLabel, "id"),
				),
			},
		},
	})
}

func generateLocationDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_location" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
