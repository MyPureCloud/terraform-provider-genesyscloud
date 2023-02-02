package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v91/platformclientv2"
	"testing"
)

func TestAccResourceRoutingSmsAddresses(t *testing.T) {

	var (
		resourceName = "sms-address1"
		name         = "address name-" + uuid.NewString()
		street       = "Main street"
		city         = "Galway"
		region       = "Galway"
		postalCode   = "H91DZ48"
		countryCode  = "US"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingSmsAddressesResource(
					resourceName,
					name,
					street,
					city,
					region,
					postalCode,
					countryCode,
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "street", street),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "city", city),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "region", region),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "postal_code", postalCode),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "country_code", countryCode),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "auto_correct_address", falseValue),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_routing_sms_address." + resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_correct_address"},
			},
		},
		CheckDestroy: testVerifySmsAddressDestroyed,
	})
}

func testVerifySmsAddressDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_sms_address" {
			continue
		}
		address, resp, err := routingAPI.GetRoutingSmsAddress(rs.Primary.ID)
		if address != nil {
			return fmt.Errorf("address (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// Address not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	//Success. All addresses destroyed
	return nil
}

func generateRoutingSmsAddressesResource(
	resourceId string,
	name string,
	street string,
	city string,
	region string,
	postalCode string,
	countryCode string,
	autoCorrectAddress string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_routing_sms_address" "%s" {
			name = "%s"
			street = "%s"
			city = "%s"
			region = "%s"
			postal_code = "%s"
			country_code = "%s"
			auto_correct_address = %s
		}
	`, resourceId, name, street, city, region, postalCode, countryCode, autoCorrectAddress)
}
