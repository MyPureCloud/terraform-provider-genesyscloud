package genesyscloud

import (
	"fmt"
	"os"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSmsAddressProdOrg(t *testing.T) {
	//this test as it will only pass in a prod org
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		t.Skip("This test as it will only pass in a prod org")
	}
	var (
		addressRes  = "addressRes"
		addressData = "addressData"

		name = "name-1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateRoutingSmsAddressesResource(
					addressRes,
					name,
					"street-1",
					"city-1",
					"region-1",
					"postal-code-1",
					"country-code-1",
					util.FalseValue,
				) + generateSmsAddressDataSource(
					addressData,
					name,
					"genesyscloud_routing_sms_address."+addressRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_routing_sms_address."+addressData, "id",
						"genesyscloud_routing_sms_address."+addressRes, "id",
					),
				),
			},
		},
	})
}

// If running in a prod org this test can be removed/skipped, it's only intended as a backup test for test orgs
func TestAccDataSourceSmsAddressTestOrg(t *testing.T) {
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "us-east-1" {
		t.Skip("Test intended only for test org")
	}
	var (
		addressRes  = "addressRes"
		addressData = "addressData"

		name = "name-1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateRoutingSmsAddressesResource(
					addressRes,
					name,
					"street-1",
					"city-1",
					"region-1",
					"90080",
					"US",
					util.TrueValue,
				) + generateSmsAddressDataSource(
					addressData,
					name,
					"genesyscloud_routing_sms_address."+addressRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_routing_sms_address."+addressData, "id",
						"genesyscloud_routing_sms_address."+addressRes, "id",
					),
				),
			},
		},
	})
}

func generateSmsAddressDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_routing_sms_address" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, id, name, dependsOn)
}
