package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v48/platformclientv2"
)

func TestAccResourceDidPoolBasic(t *testing.T) {
	var (
		didPoolResource1         = "test-didpool1"
		didPoolStartPhoneNumber1 = "+13175550000"
		didPoolEndPhoneNumber1   = "+13175550005"
		didPoolDescription1      = "Test description"
		didPoolComments1         = "Test comments"
		didPoolProvider1         = "PURE_CLOUD"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateDidPoolResource(
					didPoolResource1,
					didPoolStartPhoneNumber1,
					didPoolEndPhoneNumber1,
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "start_phone_number", didPoolStartPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "end_phone_number", didPoolEndPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "description", ""),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "comments", ""),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "pool_provider", ""),
				),
			},
			{
				// Update
				Config: generateDidPoolResource(
					didPoolResource1,
					didPoolStartPhoneNumber1,
					didPoolEndPhoneNumber1,
					strconv.Quote(didPoolDescription1),
					strconv.Quote(didPoolComments1),
					strconv.Quote(didPoolProvider1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "start_phone_number", didPoolStartPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "end_phone_number", didPoolEndPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "description", didPoolDescription1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "comments", didPoolComments1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_did_pool."+didPoolResource1, "pool_provider", didPoolProvider1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_did_pool." + didPoolResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDidPoolsDestroyed,
	})
}

func generateDidPoolResource(
	resourceID string,
	startPhoneNumber string,
	endPhoneNumber string,
	description string,
	comments string,
	poolProvider string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_did_pool" "%s" {
		start_phone_number = "%s"
		end_phone_number = "%s"
		description = %s
		comments = %s
		pool_provider = %s
	}
	`, resourceID, startPhoneNumber, endPhoneNumber, description, comments, poolProvider)
}

func testVerifyDidPoolsDestroyed(state *terraform.State) error {
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_did_pool" {
			continue
		}

		didPool, resp, err := telephonyAPI.GetTelephonyProvidersEdgesDidpool(rs.Primary.ID)
		if didPool != nil {
			if didPool.State != nil && *didPool.State == "deleted" {
				// DID pool deleted
				continue
			}
			return fmt.Errorf("DID Pool (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
			// DID pool not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All DID pool destroyed
	return nil
}
