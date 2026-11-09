package routing_skill

import (
	"fmt"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccDataSourceRoutingSkillUsers requires the PURE-7660_skills-assignments feature toggle to be enabled in the test org
func TestAccDataSourceRoutingSkillUsers(t *testing.T) {
	var (
		skillResourceLabel      = "skill"
		skillName               = "Terraform Skill Users " + uuid.NewString()
		emptySkillResourceLabel = "empty_skill"
		emptySkillName          = "Terraform Skill Users Empty " + uuid.NewString()

		user1ResourceLabel = "user1"
		user1Email         = "terraform-skill-users-" + uuid.NewString() + "@example.com"
		user2ResourceLabel = "user2"
		user2Email         = "terraform-skill-users-" + uuid.NewString() + "@example.com"

		byIdDataResourceLabel   = "by_id"
		byNameDataResourceLabel = "by_name"
		emptyDataResourceLabel  = "empty"
		skillResourcePath       = ResourceType + "." + skillResourceLabel
		emptySkillResourcePath  = ResourceType + "." + emptySkillResourceLabel
		user1ResourcePath       = user.ResourceType + "." + user1ResourceLabel
		user2ResourcePath       = user.ResourceType + "." + user2ResourceLabel
		byIdDataResourcePath    = "data." + UsersResourceType + "." + byIdDataResourceLabel
		byNameDataResourcePath  = "data." + UsersResourceType + "." + byNameDataResourceLabel
		emptyDataResourcePath   = "data." + UsersResourceType + "." + emptyDataResourceLabel
		skillIdAttrRef          = skillResourcePath + ".id"
		skillNameAttrRef        = skillResourcePath + ".name"
		emptySkillIdAttrRef     = emptySkillResourcePath + ".id"
		dependsOnUsersAndSkills = fmt.Sprintf("%s, %s, %s, %s", skillResourcePath, emptySkillResourcePath, user1ResourcePath, user2ResourcePath)
	)

	resourcesConfig := GenerateRoutingSkillResource(skillResourceLabel, skillName) +
		GenerateRoutingSkillResource(emptySkillResourceLabel, emptySkillName) +
		generateUserWithRoutingSkill(user1ResourceLabel, user1Email, "Terraform Skill User 1", skillIdAttrRef, "2.5") +
		generateUserWithRoutingSkill(user2ResourceLabel, user2Email, "Terraform Skill User 2", skillIdAttrRef, "4")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create the skills and the users assigned to them, then give the skill assignments time to propagate
				Config: resourcesConfig,
				Check: resource.ComposeTestCheckFunc(
					waitSeconds(5 * time.Second),
				),
			},
			{
				// Read the skill's users by skill ID and by skill name, and read a skill with no users
				Config: resourcesConfig +
					generateRoutingSkillUsersDataSource(byIdDataResourceLabel, "skill_id", skillIdAttrRef, dependsOnUsersAndSkills) +
					generateRoutingSkillUsersDataSource(byNameDataResourceLabel, "skill_name", skillNameAttrRef, dependsOnUsersAndSkills) +
					generateRoutingSkillUsersDataSource(emptyDataResourceLabel, "skill_id", emptySkillIdAttrRef, dependsOnUsersAndSkills),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(byIdDataResourcePath, "id", skillResourcePath, "id"),
					resource.TestCheckResourceAttrPair(byIdDataResourcePath, "skill_id", skillResourcePath, "id"),
					resource.TestCheckResourceAttr(byIdDataResourcePath, "user_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(byIdDataResourcePath, "user_ids.*", user1ResourcePath, "id"),
					resource.TestCheckTypeSetElemAttrPair(byIdDataResourcePath, "user_ids.*", user2ResourcePath, "id"),

					resource.TestCheckResourceAttrPair(byNameDataResourcePath, "id", skillResourcePath, "id"),
					resource.TestCheckResourceAttrPair(byNameDataResourcePath, "skill_name", skillResourcePath, "name"),
					resource.TestCheckResourceAttr(byNameDataResourcePath, "user_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(byNameDataResourcePath, "user_ids.*", user1ResourcePath, "id"),
					resource.TestCheckTypeSetElemAttrPair(byNameDataResourcePath, "user_ids.*", user2ResourcePath, "id"),

					resource.TestCheckResourceAttrPair(emptyDataResourcePath, "id", emptySkillResourcePath, "id"),
					resource.TestCheckResourceAttr(emptyDataResourcePath, "user_ids.#", "0"),
				),
			},
		},
	})
}

func generateUserWithRoutingSkill(resourceLabel, email, name, skillIdAttrRef, proficiency string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
	email = "%s"
	name  = "%s"
	routing_skills {
		skill_id    = %s
		proficiency = %s
	}
}
`, user.ResourceType, resourceLabel, email, name, skillIdAttrRef, proficiency)
}

func generateRoutingSkillUsersDataSource(dataResourceLabel, attrKey, attrRef, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
	%s = %s
	depends_on = [%s]
}
`, UsersResourceType, dataResourceLabel, attrKey, attrRef, dependsOn)
}
