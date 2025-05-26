package genesyscloud

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func TestAccResourceRoutingSmsAddresses(t *testing.T) {

	var (
		resourceLabel = "AD-123"
		fullPath      = ResourceType + "." + resourceLabel
		name          = "name-1"
		street        = "Strasse 66"
		city          = "Berlin"
		region        = "Berlin"
		postalCode    = "280990"
		countryCode   = "GR"
		destroyValue  = false //This type of org does not go out to SMS vendors. When you try and create an address in this case its trying to save it with the vendor, getting a mocked response and not storing any value. Hence cannot be deleted.
	)

	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		resourceLabel = "sms-address1"
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
				Config: generateRoutingSmsAddressesResource(
					resourceLabel,
					util.NullValue, // Optional
					street,
					city,
					region,
					postalCode,
					countryCode,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullPath, "name", ""),
					resource.TestCheckResourceAttr(fullPath, "street", street),
					resource.TestCheckResourceAttr(fullPath, "city", city),
					resource.TestCheckResourceAttr(fullPath, "region", region),
					resource.TestCheckResourceAttr(fullPath, "postal_code", postalCode),
					resource.TestCheckResourceAttr(fullPath, "country_code", countryCode),
					resource.TestCheckResourceAttr(fullPath, "auto_correct_address", util.FalseValue),
				),
			},
			{
				// Create
				Config: generateRoutingSmsAddressesResource(
					resourceLabel,
					strconv.Quote(name),
					street,
					city,
					region,
					postalCode,
					countryCode,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullPath, "name", name),
					resource.TestCheckResourceAttr(fullPath, "street", street),
					resource.TestCheckResourceAttr(fullPath, "city", city),
					resource.TestCheckResourceAttr(fullPath, "region", region),
					resource.TestCheckResourceAttr(fullPath, "postal_code", postalCode),
					resource.TestCheckResourceAttr(fullPath, "country_code", countryCode),
					resource.TestCheckResourceAttr(fullPath, "auto_correct_address", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:            fullPath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_correct_address"},
				Destroy:                 destroyValue,
			},
		},
		CheckDestroy: testVerifySmsAddressDestroyed,
	})
}

func testVerifySmsAddressDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
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
	resourceLabel string,
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
			name = %s
			street = "%s"
			city = "%s"
			region = "%s"
			postal_code = "%s"
			country_code = "%s"
			auto_correct_address = %s
		}
	`, resourceLabel, name, street, city, region, postalCode, countryCode, autoCorrectAddress)
}
