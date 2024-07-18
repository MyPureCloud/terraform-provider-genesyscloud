package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceRoutingWrapupcode(t *testing.T) {
	var (
		codeResource1 = "routing-wrapupcode1"
		codeName1     = "Terraform Code-" + uuid.NewString()
		codeName2     = "Terraform Code-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingWrapupcodeResource(
					codeResource1,
					codeName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+codeResource1, "name", codeName1),
				),
			},
			{
				// Update with a new name
				Config: GenerateRoutingWrapupcodeResource(
					codeResource1,
					codeName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+codeResource1, "name", codeName2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_wrapupcode." + codeResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyWrapupcodesDestroyed,
	})
}

func testVerifyWrapupcodesDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_wrapupcode" {
			continue
		}

		wrapupcode, resp, err := routingAPI.GetRoutingWrapupcode(rs.Primary.ID)
		if wrapupcode != nil {
			return fmt.Errorf("Wrapupcode (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Wrapupcode not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All wrapupcodes destroyed
	return nil
}
