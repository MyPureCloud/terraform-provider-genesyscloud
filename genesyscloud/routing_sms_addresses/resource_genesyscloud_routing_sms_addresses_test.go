package genesyscloud

import (
	"fmt"
	"os"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceRoutingSmsAddresses(t *testing.T) {

	var (
		resourceName = "AD-123"
		name         = "name-1"
		street       = "street-1"
		city         = "city-1"
		region       = "region-1"
		postalCode   = "postal-code-1"
		countryCode  = "country-code-1"
		destroyValue = false //This type of org does not go out to SMS vendors. When you try and create an address in this case its trying to save it with the vendor, getting a mocked response and not storing any value. Hence cannot be deleted.

	)
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		resourceName = "sms-address1"
		name = "name-1"
		street = "street-1"
		city = "city-1"
		region = "region-1"
		postalCode = "70090"
		countryCode = "US"
		destroyValue = true
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "street", street),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "city", city),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "region", region),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "postal_code", postalCode),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "country_code", countryCode),
					resource.TestCheckResourceAttr("genesyscloud_routing_sms_address."+resourceName, "auto_correct_address", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_routing_sms_address." + resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_correct_address"},
				Destroy:                 destroyValue,
			},
		},
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
		} else if util.IsStatus404(resp) {
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
