package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceRoutingSkillBasic(t *testing.T) {
	var (
		skillResource1 = "test-skill1"
		skillName1     = "Terraform Skill" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingSkillResource(
					skillResource1,
					skillName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill."+skillResource1, "name", skillName1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_skill." + skillResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySkillsDestroyed,
	})
}

func generateRoutingSkillResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}
	`, resourceID, name)
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
		} else if IsStatus404(resp) {
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
