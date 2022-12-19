package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceSmsAddress(t *testing.T) {
	var (
		addressRes  = "addressRes"
		addressData = "addressData"

		name = "Address" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
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

func generateSmsAddressDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_routing_sms_address" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, id, name, dependsOn)
}
