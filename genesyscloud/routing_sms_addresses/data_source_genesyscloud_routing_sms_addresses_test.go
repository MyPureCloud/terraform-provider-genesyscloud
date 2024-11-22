package genesyscloud

import (
	"fmt"
	"os"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSmsAddress(t *testing.T) {

	var (
		addressResLabel  = "addressRes"
		addressDataLabel = "addressData"
		name             = "name-1"
		street           = "street-1"
		city             = "city-1"
		region           = "region-1"
		postalCode       = "postal-code-1"
		countryCode      = "country-code-1"
	)
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		postalCode = "90080"
		countryCode = "US"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateRoutingSmsAddressesResource(
					addressResLabel,
					name,
					street,
					city,
					region,
					postalCode,
					countryCode,
					util.FalseValue,
				) + generateSmsAddressDataSource(
					addressDataLabel,
					name,
					"genesyscloud_routing_sms_address."+addressResLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_routing_sms_address."+addressDataLabel, "id",
						"genesyscloud_routing_sms_address."+addressResLabel, "id",
					),
				),
			},
		},
	})
}

func generateSmsAddressDataSource(dataSourceLabel string, name string, dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_routing_sms_address" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, dataSourceLabel, name, dependsOn)
}
