package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingSkill(t *testing.T) {
	var (
		skillResource   = "routing-skill"
		skillDataSource = "routing-skill-data"
		skillName       = "Terraform Skill-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateRoutingSkillResource(
					skillResource,
					skillName,
				) + generateRoutingSkillDataSource(skillDataSource, "genesyscloud_routing_skill."+skillResource+".name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_skill."+skillDataSource, "id", "genesyscloud_routing_skill."+skillResource, "id"),
				),
			},
		},
	})
}

func generateRoutingSkillDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_skill" "%s" {
		name = %s
	}
	`, resourceID, name)
}
