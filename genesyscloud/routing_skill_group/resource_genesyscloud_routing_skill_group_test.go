package routing_skill_group

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func testAccCheckSkillConditions(resourcePath string, targetSkillConditionJson string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourcePath]

		if !ok {
			return fmt.Errorf("resource Not found: %s", resourcePath)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		//Retrieve the skills condition
		resourceSkillConditionsJson := rs.Primary.Attributes["skill_conditions"]

		//Convert the resource and target skill condition to []map. This is an intermediary format.
		var resourceSkillConditionsMap []map[string]interface{}
		var targetSkillConditionsMap []map[string]interface{}

		if err := json.Unmarshal([]byte(resourceSkillConditionsJson), &resourceSkillConditionsMap); err != nil {
			return fmt.Errorf("error converting resource skill conditions from JSON to a Map: %s", err)
		}

		if err := json.Unmarshal([]byte(targetSkillConditionJson), &targetSkillConditionsMap); err != nil {
			return fmt.Errorf("error converting target skill conditions to a Map: %s", err)
		}

		//Convert the resource and target maps back to a string so they have the exact same format.
		r, err := json.Marshal(resourceSkillConditionsMap)
		if err != nil {
			return fmt.Errorf("error converting the resource map back from a Map to JSON: %s", err)
		}
		t, err := json.Marshal(targetSkillConditionsMap)
		if err != nil {
			return fmt.Errorf("error converting the target map back from a Map to JSON: %s", err)
		}

		//Checking to see if our 2 JSON strings are exactly equal.
		resourceStr := string(r)
		target := string(t)
		if resourceStr != target {
			return fmt.Errorf("resource skill_conditions does not match skill_conditions passed in. Expected: %s Actual: %s", resourceStr, target)
		}

		return nil
	}
}

func TestAccResourceRoutingSkillGroupBasic(t *testing.T) {
	t.Parallel()
	var (
		skillGroupResourceLabel    = "testskillgroup1"
		skillGroupResourceFullPath = ResourceType + "." + skillGroupResourceLabel
		skillGroupName1            = "SkillGroup1" + uuid.NewString()
		skillGroupDescription1     = "Description1" + uuid.NewString()
		skillGroupName2            = "SkillGroup2" + uuid.NewString()
		skillGroupDescription2     = "Description2" + uuid.NewString()
		skillCondition1            = `[
			{
			  "routingSkillConditions" : [
				{
				  "routingSkill" : "Series 6",
				  "comparator" : "GreaterThan",
				  "proficiency" : 2,
				  "childConditions" : [{
					"routingSkillConditions" : [],
					"languageSkillConditions" : [],
					"operation" : "And"
				  }]
				}
			  ],
			  "languageSkillConditions" : [],
			  "operation" : "And"
		  }]`

		skillCondition2 = `[
			{
			  "routingSkillConditions" : [
				{
				  "routingSkill" : "Series 6",
				  "comparator" : "EqualTo",
				  "proficiency" : 4,
				  "childConditions" : [{
					"routingSkillConditions" : [],
					"languageSkillConditions" : [],
					"operation" : "And"
				  }]
				}
			  ],
			  "languageSkillConditions" : [],
			  "operation" : "And"
		  }]`
	)

	config1 := `
data "genesyscloud_auth_division_home" "home" {}
` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName1,
		skillGroupDescription1,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition1,
		util.NullValue,
	)

	config2 := `
data "genesyscloud_auth_division_home" "home" {}
` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		util.NullValue,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName1),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription1),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition1),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),
				),
			},
			{
				// Update
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName2),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription2),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition2),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),
				),
			},
			{
				// Import/Read
				ResourceName:      skillGroupResourceFullPath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySkillGroupDestroyed,
	})
}

func TestAccResourceRoutingSkillGroupMemberDivisionsBasic(t *testing.T) {
	t.Parallel()
	var (
		skillGroupResourceLabel    = "testskillgroup2"
		skillGroupResourceFullPath = ResourceType + "." + skillGroupResourceLabel
		skillGroupName1            = "SkillGroup3" + uuid.NewString()
		skillGroupDescription1     = "Description3" + uuid.NewString()
		skillGroupName2            = "SkillGroup4" + uuid.NewString()
		skillGroupDescription2     = "Description4" + uuid.NewString()
		skillCondition1            = `[
			{
			  "routingSkillConditions" : [
				{
				  "routingSkill" : "Series 6",
				  "comparator" : "GreaterThan",
				  "proficiency" : 2,
				  "childConditions" : [{
					"routingSkillConditions" : [],
					"languageSkillConditions" : [],
					"operation" : "And"
				  }]
				}
			  ],
			  "languageSkillConditions" : [],
			  "operation" : "And"
		  }]`

		skillCondition2 = `[
			{
			  "routingSkillConditions" : [
				{
				  "routingSkill" : "Series 6",
				  "comparator" : "EqualTo",
				  "proficiency" : 4,
				  "childConditions" : [{
					"routingSkillConditions" : [],
					"languageSkillConditions" : [],
					"operation" : "And"
				  }]
				}
			  ],
			  "languageSkillConditions" : [],
			  "operation" : "And"
		  }]`
	)

	authDivision1Name := "TF Division " + uuid.NewString()
	authDivision1ResourceLabel := "division1"
	authDivision1 := authDivision.GenerateAuthDivisionBasic(authDivision1ResourceLabel, authDivision1Name)

	authDivision2Name := "TF Division " + uuid.NewString()
	authDivision2ResourceLabel := "division2"
	authDivision2 := authDivision.GenerateAuthDivisionBasic(authDivision2ResourceLabel, authDivision2Name)

	memberDivisionIds1 := fmt.Sprintf(`[%s]`, strings.Join([]string{"data.genesyscloud_auth_division_home.home.id"}, ", "))

	memberDivisionIds2 := fmt.Sprintf(`[%s]`, strings.Join([]string{
		"data.genesyscloud_auth_division_home.home.id",
		"genesyscloud_auth_division." + authDivision1ResourceLabel + ".id",
		"genesyscloud_auth_division." + authDivision2ResourceLabel + ".id",
	}, ", "))

	memberDivisionIds3 := fmt.Sprintf(`[%s]`, strings.Join([]string{
		"data.genesyscloud_auth_division_home.home.id",
		"genesyscloud_auth_division." + authDivision1ResourceLabel + ".id",
	}, ", "))

	config1 := `
data "genesyscloud_auth_division_home" "home" {}
` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName1,
		skillGroupDescription1,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition1,
		memberDivisionIds1,
	)

	config2 := `
data "genesyscloud_auth_division_home" "home" {}
` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		memberDivisionIds2,
	) + authDivision1 + authDivision2

	config3 := `
data "genesyscloud_auth_division_home" "home" {}
` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		memberDivisionIds3,
	) + authDivision1

	config4 := `
	data "genesyscloud_auth_division_home" "home" {}
	` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		"[]",
	)

	config5 := `
data "genesyscloud_auth_division_home" "home" {}
` + generateRoutingSkillGroupResource(
		skillGroupResourceLabel,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		`["*"]`,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName1),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription1),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition1),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),

					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "member_division_ids.#", "1"),
					util.ValidateResourceAttributeInArray(skillGroupResourceFullPath, "member_division_ids",
						"data.genesyscloud_auth_division_home.home", "id"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName2),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription2),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition2),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),

					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "member_division_ids.#", "3"),
					util.ValidateResourceAttributeInArray(skillGroupResourceFullPath, "member_division_ids",
						"data.genesyscloud_auth_division_home.home", "id"),
					util.ValidateResourceAttributeInArray(skillGroupResourceFullPath, "member_division_ids",
						"genesyscloud_auth_division."+authDivision1ResourceLabel, "id"),
					util.ValidateResourceAttributeInArray(skillGroupResourceFullPath, "member_division_ids",
						"genesyscloud_auth_division."+authDivision2ResourceLabel, "id"),
				),
			},
			{
				Config: config3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName2),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription2),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition2),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),

					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "member_division_ids.#", "2"),
					util.ValidateResourceAttributeInArray(skillGroupResourceFullPath, "member_division_ids",
						"data.genesyscloud_auth_division_home.home", "id"),
					util.ValidateResourceAttributeInArray(skillGroupResourceFullPath, "member_division_ids",
						"genesyscloud_auth_division."+authDivision1ResourceLabel, "id"),
				),
			},
			{
				// Update members array to [] and verify skill group's division is still in there
				Config: config4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName2),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription2),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition2),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),

					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "member_division_ids.#", "0"),
					testVerifyMemberDivisionsCleared(skillGroupResourceFullPath),
				),
			},
			{
				// Update members array to ["*"] and verify all division ids are in there.
				Config: config5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "name", skillGroupName2),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "description", skillGroupDescription2),
					testAccCheckSkillConditions(skillGroupResourceFullPath, skillCondition2),
					provider.TestDefaultHomeDivision(skillGroupResourceFullPath),

					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "member_division_ids.#", "1"),
					resource.TestCheckResourceAttr(skillGroupResourceFullPath, "member_division_ids.0", "*"),
					testVerifyAllDivisionsAssigned(skillGroupResourceFullPath, "member_division_ids"),
				),
			},
			{
				ResourceName:            skillGroupResourceFullPath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"member_division_ids"},
			},
		},
		CheckDestroy: testVerifySkillGroupDestroyed,
	})
}

/*
1. Create users with a particular set of skills and assign to individual divisions.
2. Add those divisions to genesyscloud_routing_skill_group.members_divisions_ids array.
3. Verify the skill group added those users when they match the skill expression.
*/
func TestAccResourceRoutingSkillGroupMemberDivisionsUsersAssigned(t *testing.T) {
	var (
		skillGroupResourceLabel    = "testskillgroup3"
		skillGroupName             = "testskillgroup3 " + uuid.NewString()
		skillGroupDescription      = uuid.NewString()
		skillGroupResourceFullPath = ResourceType + "." + skillGroupResourceLabel

		routingSkillResourceLabel = "routing_skill"
		routingSkillName          = "Skill " + uuid.NewString()

		user1ResourceLabel = "user_1"
		user2ResourceLabel = "user_2"
		user3ResourceLabel = "user_3"
		user1Name          = "tf.test.user " + uuid.NewString()
		user2Name          = "tf.test.user " + uuid.NewString()
		user3Name          = "tf.test.user " + uuid.NewString()
		user1email         = "terraform-" + uuid.NewString() + "@example.com"
		user2email         = "terraform-" + uuid.NewString() + "@example.com"
		user3email         = "terraform-" + uuid.NewString() + "@example.com"

		division1ResourceLabel = "division_1"
		division2ResourceLabel = "division_2"
		division3ResourceLabel = "division_3"
		division1Name          = "tf test divisionB " + uuid.NewString()
		division2Name          = "tf test divisionB " + uuid.NewString()
		division3Name          = "tf test divisionB " + uuid.NewString()

		memberDivisionIds = []string{
			"genesyscloud_auth_division." + division1ResourceLabel + ".id",
			"genesyscloud_auth_division." + division2ResourceLabel + ".id",
			"genesyscloud_auth_division." + division3ResourceLabel + ".id",
		}
	)

	routingSkillResource := routingSkill.GenerateRoutingSkillResource(routingSkillResourceLabel, routingSkillName)

	division1Resource := authDivision.GenerateAuthDivisionBasic(division1ResourceLabel, division1Name)
	division2Resource := authDivision.GenerateAuthDivisionBasic(division2ResourceLabel, division2Name)
	division3Resource := authDivision.GenerateAuthDivisionBasic(division3ResourceLabel, division3Name)

	user1Resource := fmt.Sprintf(`
resource "genesyscloud_user" "%s" {
	name        = "%s"
	email       = "%s"
	division_id = genesyscloud_auth_division.%s.id
	routing_skills {
		skill_id    = genesyscloud_routing_skill.%s.id
    	proficiency = 2.5
	}
}
`, user1ResourceLabel, user1Name, user1email, division1ResourceLabel, routingSkillResourceLabel)

	user2Resource := fmt.Sprintf(`
resource "genesyscloud_user" "%s" {
	name        = "%s"
	email       = "%s"
	division_id = genesyscloud_auth_division.%s.id
	routing_skills {
		skill_id    = genesyscloud_routing_skill.%s.id
    	proficiency = 2.5
	}
}
`, user2ResourceLabel, user2Name, user2email, division2ResourceLabel, routingSkillResourceLabel)

	user3Resource := fmt.Sprintf(`
resource "genesyscloud_user" "%s" {
	name        = "%s"
	email       = "%s"
	division_id = genesyscloud_auth_division.%s.id
	routing_skills {
		skill_id    = genesyscloud_routing_skill.%s.id
    	proficiency = 2.5
	}
}
`, user3ResourceLabel, user3Name, user3email, division3ResourceLabel, routingSkillResourceLabel)

	skillGroupResource := fmt.Sprintf(`
resource "genesyscloud_routing_skill_group" "%s" {
	name = "%s"
	member_division_ids = [%s]
	description = "%s"
	skill_conditions = jsonencode(
		[
		  {
			"routingSkillConditions" : [
			  {
				"routingSkill" : "%s",
				"comparator" : "GreaterThan",
				"proficiency" : 2,
				"childConditions" : [{
				  "routingSkillConditions" : [],
				  "languageSkillConditions" : [],
				  "operation" : "And"
				}]
			  }
			],
			"languageSkillConditions" : [],
			"operation" : "And"
		}]
	)

	depends_on = [genesyscloud_user.%s, genesyscloud_user.%s, genesyscloud_user.%s ]
}
`, skillGroupResourceLabel, skillGroupName, strings.Join(memberDivisionIds, ", "),
		skillGroupDescription, routingSkillName, user1ResourceLabel, user2ResourceLabel, user3ResourceLabel)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: routingSkillResource +
					division1Resource +
					division2Resource +
					division3Resource +
					user1Resource +
					user2Resource +
					user3Resource,
			},
			{
				PreConfig: func() {
					time.Sleep(45 * time.Second)
				},
				Config: skillGroupResource +
					routingSkillResource +
					division1Resource +
					division2Resource +
					division3Resource +
					user1Resource +
					user2Resource +
					user3Resource,
				Check: resource.ComposeTestCheckFunc(
					testVerifySkillGroupMemberCount(skillGroupResourceFullPath, 3),
				),
			},
			{
				ResourceName:            skillGroupResourceFullPath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"member_division_ids"},
				Destroy:                 true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(60 * time.Second)
			return testVerifySkillGroupAndUsersDestroyed(state)
		},
	})
}

func generateRoutingSkillGroupResource(
	resourceLabel string,
	divisionResourcePath string,
	name string,
	description string,
	divisionID string,
	skillCondition string,
	memberDivisionIds string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill_group" "%s" {
		depends_on = [%s]
		name = "%s"
		description="%s"
		division_id=%s
		skill_conditions = jsonencode(%s)
		member_division_ids = %s
	}
	`, resourceLabel, divisionResourcePath, name, description, divisionID, skillCondition, memberDivisionIds)
}

func testVerifySkillGroupMemberCount(resourcePath string, count int) resource.TestCheckFunc {
	const logPrefix = "testVerifySkillGroupMemberCount:"
	return func(state *terraform.State) error {
		retryErr := util.WithRetries(context.Background(), 2*time.Minute, func() *retry.RetryError {
			routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

			resourceState, ok := state.RootModule().Resources[resourcePath]
			if !ok {
				return retry.NonRetryableError(fmt.Errorf("failed to find resourceState %s in state", resourcePath))
			}
			resourceId := resourceState.Primary.ID

			log.Printf("%s Reading skill group '%s'", logPrefix, resourceId)
			skillGroup, resp, err := routingAPI.GetRoutingSkillgroup(resourceId)
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("%s skill group '%s' not found", logPrefix, resourceId))
			}
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%s failed to get skill group %s: %s. Response: %s", logPrefix, resourceId, err.Error(), resp.String()))
			}

			if *skillGroup.MemberCount != count {
				return retry.RetryableError(fmt.Errorf("expected member count to be %d, got %d for skill group '%s'", count, *skillGroup.MemberCount, resourceId))
			}

			return nil
		})
		if retryErr != nil {
			return fmt.Errorf("%v", retryErr)
		}
		return nil
	}
}

func testVerifyMemberDivisionsCleared(resourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resourceState %s in state", resourcePath)
		}
		resourceId := resourceState.Primary.ID

		// get member divisions for this skill group via GET /api/v2/routing/skillgroups/{skillGroupId}/members/divisions
		skillGroupMemberDivisionIds, diagErr := getAllSkillGroupMemberDivisionIds(resourceId)
		if diagErr != nil {
			return fmt.Errorf("%v", diagErr)
		}

		divisionId, ok := resourceState.Primary.Attributes["division_id"]
		if !ok {
			return fmt.Errorf("no divisionId found for %s in state", resourceId)
		}

		if len(skillGroupMemberDivisionIds) != 1 {
			return fmt.Errorf("expected skill group %s to only have one member division assigned", resourceId)
		}

		if divisionId != skillGroupMemberDivisionIds[0] {
			return fmt.Errorf("expected division_id %s to equal the assigned member division ID %s for skill group %s", divisionId, skillGroupMemberDivisionIds[0], resourceId)
		}

		return nil
	}
}

func testVerifyAllDivisionsAssigned(resourcePath string, attrName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resourceState %s in state", resourcePath)
		}

		resourceId := resourceState.Primary.ID
		numValuesStr, ok := resourceState.Primary.Attributes[attrName+".#"]
		if !ok {
			return fmt.Errorf("no %s found for %s in state", attrName, resourceId)
		}

		if numValuesStr != "1" || resourceState.Primary.Attributes[attrName+".0"] != "*" {
			return fmt.Errorf(`expected %s to contain one item: "*"`, attrName)
		}

		// get member divisions for this skill group via GET /api/v2/routing/skillgroups/{skillGroupId}/members/divisions
		skillGroupMemberDivisionIds, diagErr := getAllSkillGroupMemberDivisionIds(resourceId)
		if diagErr != nil {
			return fmt.Errorf("%v", diagErr)
		}

		// get all auth divisions via GET /api/v2/authorization/divisions
		allAuthDivisionIds := make([]string, 0)
		divisionResourcesMap, diagErr := getAllAuthDivisions(context.Background(), sdkConfig)
		if diagErr != nil {
			return fmt.Errorf("%v", diagErr)
		}

		for id := range divisionResourcesMap {
			allAuthDivisionIds = append(allAuthDivisionIds, id)
		}

		// member_division_ids should not contain more than one item when the value of an item is "*"
		if lists.ItemInSlice("*", skillGroupMemberDivisionIds) {
			return nil
		} else if lists.AreEquivalent(allAuthDivisionIds, skillGroupMemberDivisionIds) {
			return nil
		} else {
			return fmt.Errorf("expected %s to equal the list of all auth divisions", attrName)
		}

	}
}

func testVerifySkillGroupDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		skillGroup, resp, err := routingAPI.GetRoutingSkillgroup(rs.Primary.ID)

		if skillGroup != nil {
			return fmt.Errorf("skill Group (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Division not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All skills destroyed
	return nil
}

func testVerifySkillGroupAndUsersDestroyed(state *terraform.State) error {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return fmt.Errorf("unexpected error while trying to authorize client in testVerifySkillGroupDestroyed : %s", err)
	}
	routingAPI := platformclientv2.NewRoutingApiWithConfig(config)
	usersAPI := platformclientv2.NewUsersApiWithConfig(config)

	for _, rs := range state.RootModule().Resources {
		if rs.Type == ResourceType {
			group, response, err := routingAPI.GetRoutingSkillgroup(rs.Primary.ID)

			if group != nil {
				return fmt.Errorf("team (%s) still exists", rs.Primary.ID)
			}
			if util.IsStatus404(response) {
				continue
			}
			return fmt.Errorf("unexpected error: %s", err)
		}

		if rs.Type == "genesyscloud_user" {
			err = checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if user != nil {
				return fmt.Errorf("user Resource (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				continue
			} else {
				return fmt.Errorf("unexpected error: %s", err)
			}
		}
	}
	// Success. All skills destroyed
	return nil
}

func getAllSkillGroupMemberDivisionIds(resourceLabel string) ([]string, diag.Diagnostics) {
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	divisions, resp, err := api.GetRoutingSkillgroupMembersDivisions(resourceLabel, "")

	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Routing Utilization %s error: %s", resourceLabel, err), resp)
	}

	apiSkillGroupMemberDivisionIds := make([]string, 0)
	for _, entity := range *divisions.Entities {
		apiSkillGroupMemberDivisionIds = append(apiSkillGroupMemberDivisionIds, *entity.Id)

	}

	return apiSkillGroupMemberDivisionIds, nil
}

func checkUserDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching user with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 30
		for i := 0; i < maxAttempts; i++ {

			deleted, err := isUserDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("user %s was not deleted properly", id)
	}
}

func isUserDeleted(id string) (bool, error) {
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)
	// Attempt to get the user
	_, response, err := usersAPI.GetUser(id, nil, "", "")

	// Check if the user is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // User is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If user is found, it means the user is not deleted
	return false, nil
}
