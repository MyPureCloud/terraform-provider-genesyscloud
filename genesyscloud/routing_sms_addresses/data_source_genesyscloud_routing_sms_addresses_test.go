package genesyscloud

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSmsAddress(t *testing.T) {
	var (
		addressResLabel  = "address"
		addressDataLabel = "address_data"
		name             = "name-1"
		street           = "street-1"
		city             = "city-1"
		region           = "region-1"
		postalCode       = "postal-code-1"
		countryCode      = "country-code-1"
	)
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" || v == "us-west-2" {
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
					strconv.Quote(name),
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
				Destroy: false,
			},
		},
	})
}

func generateSmsAddressDataSource(dataSourceLabel, name, dependsOn string) string {
	return fmt.Sprintf(`
		data "%s" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, ResourceType, dataSourceLabel, name, dependsOn)
}
