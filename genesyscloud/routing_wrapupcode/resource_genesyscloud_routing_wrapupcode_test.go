package routing_wrapupcode

import (
	"fmt"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceRoutingWrapupcode(t *testing.T) {
	var (
		codeResourceLabel1 = "routing-wrapupcode1"
		codeName1          = "Terraform Code-" + uuid.NewString()
		codeName2          = "Terraform Code-" + uuid.NewString()
		divResourceLabel   = "test-division"
		divName            = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingWrapupcodeResource(
					codeResourceLabel1,
					codeName1,
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "name", codeName1),
				),
			},
			{
				// Create
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + GenerateRoutingWrapupcodeResource(
					codeResourceLabel1,
					codeName1,
					"genesyscloud_auth_division."+divResourceLabel+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "name", codeName1),
					resource.TestCheckResourceAttrPair(ResourceType+"."+codeResourceLabel1, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
				),
			},
			{
				// Update with a new name
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + GenerateRoutingWrapupcodeResource(
					codeResourceLabel1,
					codeName2,
					"genesyscloud_auth_division."+divResourceLabel+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "name", codeName2),
					resource.TestCheckResourceAttrPair(ResourceType+"."+codeResourceLabel1, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + codeResourceLabel1,
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
		if rs.Type != ResourceType {
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
