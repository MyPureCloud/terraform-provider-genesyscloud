package routing_skill

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceRoutingSkillBasic(t *testing.T) {
	var (
		skillResourceLabel1 = "test-skill1"
		skillName1          = "Terraform Skill" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingSkillResource(
					skillResourceLabel1,
					skillName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill."+skillResourceLabel1, "name", skillName1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_skill." + skillResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySkillsDestroyed,
	})
}

func testVerifySkillsDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_skill" {
			continue
		}

		skill, resp, err := routingAPI.GetRoutingSkill(rs.Primary.ID)
		if skill != nil {
			if skill.State != nil && *skill.State == "deleted" {
				// Skill deleted
				continue
			}
			return fmt.Errorf("Skill (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Skill not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All skills destroyed
	return nil
}
