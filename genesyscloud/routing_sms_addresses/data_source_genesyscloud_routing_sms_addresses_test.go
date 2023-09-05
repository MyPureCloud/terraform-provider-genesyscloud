package genesyscloud

import (
	"fmt"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSmsAddressProdOrg(t *testing.T) {
	t.Skip("Skip this test as it will only pass in a prod org")
	var (
		addressRes  = "addressRes"
		addressData = "addressData"

		name = "Address" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateRoutingSmsAddressesResource(
					addressRes,
					name,
					"Main street",
					"New York",
					"New York",
					"AA34HH",
					"US",
					falseValue,
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
	var (
		addressRes  = "addressRes"
		addressData = "addressData"

		name = "name-1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
					trueValue,
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
