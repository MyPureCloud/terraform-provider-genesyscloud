package routing_skill_group

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func testAccCheckSkillConditions(resourceName string, targetSkillConditionJson string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Resource Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID is not set")
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
		skillGroupResource     = "testskillgroup1"
		skillGroupName1        = "SkillGroup1" + uuid.NewString()
		skillGroupDescription1 = "Description1" + uuid.NewString()
		skillGroupName2        = "SkillGroup2" + uuid.NewString()
		skillGroupDescription2 = "Description2" + uuid.NewString()
		skillCondition1        = `[
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

	config1 := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingSkillGroupResource(
		skillGroupResource,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName1,
		skillGroupDescription1,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition1,
		util.NullValue,
	)

	config2 := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingSkillGroupResource(
		skillGroupResource,
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
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription1),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition1),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),
				),
			},
			{
				// Update
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription2),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition2),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_skill_group." + skillGroupResource,
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
		skillGroupResource     = "testskillgroup2"
		skillGroupName1        = "SkillGroup3" + uuid.NewString()
		skillGroupDescription1 = "Description3" + uuid.NewString()
		skillGroupName2        = "SkillGroup4" + uuid.NewString()
		skillGroupDescription2 = "Description4" + uuid.NewString()
		skillCondition1        = `[
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
	authDivision1Resource := "division1"
	authDivision1 := gcloud.GenerateAuthDivisionBasic(authDivision1Resource, authDivision1Name)

	authDivision2Name := "TF Division " + uuid.NewString()
	authDivision2Resource := "division2"
	authDivision2 := gcloud.GenerateAuthDivisionBasic(authDivision2Resource, authDivision2Name)

	memberDivisionIds1 := fmt.Sprintf(`[%s]`, strings.Join([]string{"data.genesyscloud_auth_division_home.home.id"}, ", "))

	memberDivisionIds2 := fmt.Sprintf(`[%s]`, strings.Join([]string{
		"data.genesyscloud_auth_division_home.home.id",
		"genesyscloud_auth_division." + authDivision1Resource + ".id",
		"genesyscloud_auth_division." + authDivision2Resource + ".id",
	}, ", "))

	memberDivisionIds3 := fmt.Sprintf(`[%s]`, strings.Join([]string{
		"data.genesyscloud_auth_division_home.home.id",
		"genesyscloud_auth_division." + authDivision1Resource + ".id",
	}, ", "))

	config1 := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingSkillGroupResource(
		skillGroupResource,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName1,
		skillGroupDescription1,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition1,
		memberDivisionIds1,
	)

	config2 := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingSkillGroupResource(
		skillGroupResource,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		memberDivisionIds2,
	) + authDivision1 + authDivision2

	config3 := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingSkillGroupResource(
		skillGroupResource,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		memberDivisionIds3,
	) + authDivision1

	config4 := fmt.Sprintf(`
	data "genesyscloud_auth_division_home" "home" {}
	`) + generateRoutingSkillGroupResource(
		skillGroupResource,
		"data.genesyscloud_auth_division_home.home",
		skillGroupName2,
		skillGroupDescription2,
		"data.genesyscloud_auth_division_home.home.id",
		skillCondition2,
		"[]",
	)

	config5 := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingSkillGroupResource(
		skillGroupResource,
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
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription1),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition1),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),

					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids.#", "1"),
					util.ValidateResourceAttributeInArray("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids",
						"data.genesyscloud_auth_division_home.home", "id"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription2),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition2),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),

					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids.#", "3"),
					util.ValidateResourceAttributeInArray("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids",
						"data.genesyscloud_auth_division_home.home", "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids",
						"genesyscloud_auth_division."+authDivision1Resource, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids",
						"genesyscloud_auth_division."+authDivision2Resource, "id"),
				),
			},
			{
				Config: config3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription2),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition2),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),

					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids.#", "2"),
					util.ValidateResourceAttributeInArray("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids",
						"data.genesyscloud_auth_division_home.home", "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids",
						"genesyscloud_auth_division."+authDivision1Resource, "id"),
				),
			},
			{
				// Update members array to [] and verify skill group's division is still in there
				Config: config4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription2),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition2),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),

					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids.#", "0"),
					testVerifyMemberDivisionsCleared("genesyscloud_routing_skill_group."+skillGroupResource),
				),
			},
			{
				// Update members array to ["*"] and verify all division ids are in there.
				Config: config5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription2),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition2),
					provider.TestDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),

					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids.0", "*"),
					testVerifyAllDivisionsAssigned("genesyscloud_routing_skill_group."+skillGroupResource, "member_division_ids"),
				),
			},
			{
				ResourceName:            "genesyscloud_routing_skill_group." + skillGroupResource,
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
		skillGroupResourceId  = "testskillgroup3"
		skillGroupName        = "testskillgroup3 " + uuid.NewString()
		skillGroupDescription = uuid.NewString()

		routingSkillResourceId = "routing_skill"
		routingSkillName       = "Skill " + uuid.NewString()

		user1ResourceId = "user_1"
		user2ResourceId = "user_2"
		user3ResourceId = "user_3"
		user1Name       = "tf.test.user " + uuid.NewString()
		user2Name       = "tf.test.user " + uuid.NewString()
		user3Name       = "tf.test.user " + uuid.NewString()
		user1email      = "terraform-" + uuid.NewString() + "@example.com"
		user2email      = "terraform-" + uuid.NewString() + "@example.com"
		user3email      = "terraform-" + uuid.NewString() + "@example.com"

		division1ResourceId = "division_1"
		division2ResourceId = "division_2"
		division3ResourceId = "division_3"
		division1Name       = "tf test divisionB " + uuid.NewString()
		division2Name       = "tf test divisionB " + uuid.NewString()
		division3Name       = "tf test divisionB " + uuid.NewString()

		memberDivisionIds = []string{
			"genesyscloud_auth_division." + division1ResourceId + ".id",
			"genesyscloud_auth_division." + division2ResourceId + ".id",
			"genesyscloud_auth_division." + division3ResourceId + ".id",
		}
	)

	routingSkillResource := routingSkill.GenerateRoutingSkillResource(routingSkillResourceId, routingSkillName)

	division1Resource := gcloud.GenerateAuthDivisionBasic(division1ResourceId, division1Name)
	division2Resource := gcloud.GenerateAuthDivisionBasic(division2ResourceId, division2Name)
	division3Resource := gcloud.GenerateAuthDivisionBasic(division3ResourceId, division3Name)

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
`, user1ResourceId, user1Name, user1email, division1ResourceId, routingSkillResourceId)

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
`, user2ResourceId, user2Name, user2email, division2ResourceId, routingSkillResourceId)

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
`, user3ResourceId, user3Name, user3email, division3ResourceId, routingSkillResourceId)

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
`, skillGroupResourceId, skillGroupName, strings.Join(memberDivisionIds, ", "),
		skillGroupDescription, routingSkillName, user1ResourceId, user2ResourceId, user3ResourceId)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
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
					testVerifySkillGroupMemberCount("genesyscloud_routing_skill_group."+skillGroupResourceId, 3),
				),
			},
			{
				ResourceName:            "genesyscloud_routing_skill_group." + skillGroupResourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"member_division_ids"},
				Destroy:                 true,
			},
		},
		CheckDestroy: testVerifySkillGroupAndUsersDestroyed,
	})
}

func generateRoutingSkillGroupResource(
	resourceID string,
	divisionResourceName string,
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
	`, resourceID, divisionResourceName, name, description, divisionID, skillCondition, memberDivisionIds)
}

func testVerifySkillGroupMemberCount(resourceName string, count int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// Authorize client credentials
		config, err := provider.AuthorizeSdk()
		if err != nil {
			return fmt.Errorf("unexpected error while trying to authorize client in testVerifyAllDivisionsAssigned : %s", err)
		}
		routingAPI := platformclientv2.NewRoutingApiWithConfig(config)

		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		log.Print("Sleeping to allow for skillgroups member count to update.")
		time.Sleep(10 * time.Second)

		// get skill group via GET /api/v2/routing/skillgroups/{skillGroupId}
		skillGroup, resp, err := routingAPI.GetRoutingSkillgroup(resourceID)
		if err != nil {
			return fmt.Errorf("Failed to get skill group %s: %v %s", resourceID, err, resp)
		}

		if *skillGroup.MemberCount != count {
			return fmt.Errorf("Expected member count to be %v, got %v for skill group %s", count, *skillGroup.MemberCount, resourceID)
		}
		return nil
	}
}

func testVerifyMemberDivisionsCleared(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		// Authorize client credentials
		config := platformclientv2.GetDefaultConfiguration()
		routingAPI := platformclientv2.NewRoutingApi()
		err := config.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
		if err != nil {
			return fmt.Errorf("Unexpected error while trying to authorize client in testVerifyAllDivisionsAssigned : %s", err)
		}

		// get member divisions for this skill group via GET /api/v2/routing/skillgroups/{skillGroupId}/members/divisions
		skillGroupMemberDivisionIds, diagErr := getAllSkillGroupMemberDivisionIds(routingAPI, resourceID)
		if diagErr != nil {
			return fmt.Errorf("%v", diagErr)
		}

		divisionId, ok := resourceState.Primary.Attributes["division_id"]
		if !ok {
			return fmt.Errorf("No divisionId found for %s in state", resourceID)
		}

		if len(skillGroupMemberDivisionIds) != 1 {
			return fmt.Errorf("Expected skill group %s to only have one member division assigned", resourceID)
		}

		if divisionId != skillGroupMemberDivisionIds[0] {
			return fmt.Errorf("Expected division_id %s to equal the assigned member division ID %s for skill group %s", divisionId, skillGroupMemberDivisionIds[0], resourceID)
		}

		return nil
	}
}

func testVerifyAllDivisionsAssigned(resourceName string, attrName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}

		resourceID := resourceState.Primary.ID
		numValuesStr, ok := resourceState.Primary.Attributes[attrName+".#"]
		if !ok {
			return fmt.Errorf("No %s found for %s in state", attrName, resourceID)
		}

		if numValuesStr != "1" || resourceState.Primary.Attributes[attrName+".0"] != "*" {
			return fmt.Errorf(`Expected %s to contain one item: "*"`, attrName)
		}

		// Authorize client credentials
		config := platformclientv2.GetDefaultConfiguration()
		routingAPI := platformclientv2.NewRoutingApi()
		err := config.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
		if err != nil {
			return fmt.Errorf("Unexpected error while trying to authorize client in testVerifyAllDivisionsAssigned : %s", err)
		}

		// get member divisions for this skill group via GET /api/v2/routing/skillgroups/{skillGroupId}/members/divisions
		skillGroupMemberDivisionIds, diagErr := getAllSkillGroupMemberDivisionIds(routingAPI, resourceID)
		if diagErr != nil {
			return fmt.Errorf("%v", diagErr)
		}

		// get all auth divisions via GET /api/v2/authorization/divisions
		allAuthDivisionIds := make([]string, 0)
		divisionResourcesMap, diagErr := getAllAuthDivisions(nil, config)
		if err != nil {
			return fmt.Errorf("%v", diagErr)
		}

		for id, _ := range divisionResourcesMap {
			allAuthDivisionIds = append(allAuthDivisionIds, id)
		}

		// Preventing a large n² comparison equation from executing
		maxLengthForListItemComparision := 20
		if len(allAuthDivisionIds) < maxLengthForListItemComparision {
			if lists.AreEquivalent(allAuthDivisionIds, skillGroupMemberDivisionIds) {
				return nil
			} else {
				return fmt.Errorf("Expected %s to equal the list of all auth divisions", attrName)
			}
		}

		if len(allAuthDivisionIds) == len(skillGroupMemberDivisionIds) {
			return nil
		}

		return fmt.Errorf("Expected %s length to equal the number of all auth divisions", attrName)
	}
}

func testVerifySkillGroupDestroyed(state *terraform.State) error {
	// Get default config to set config options
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return fmt.Errorf("unexpected error while trying to authorize client in testVerifySkillGroupDestroyed : %s", err)
	}
	routingAPI := platformclientv2.NewRoutingApiWithConfig(config)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_skill_group" {
			continue
		}

		skillGroup, resp, err := routingAPI.GetRoutingSkillgroup(rs.Primary.ID)

		if skillGroup != nil {
			return fmt.Errorf("Skill Group (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Division not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
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
		if rs.Type == "genesyscloud_routing_skill_group" {
			group, response, err := routingAPI.GetRoutingSkillgroup(rs.Primary.ID)

			if group != nil {
				return fmt.Errorf("team (%s) still exists", rs.Primary.ID)
			}
			if util.IsStatus404(response) {
				continue
			}
			return fmt.Errorf("Unexpected error: %s", err)
		}

		if rs.Type == "genesyscloud_user" {
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if user != nil {
				return fmt.Errorf("User Resource (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				continue
			} else {
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
	}
	// Success. All skills destroyed
	return nil
}
func getAllSkillGroupMemberDivisionIds(routingAPI *platformclientv2.RoutingApi, resourceId string) ([]string, diag.Diagnostics) {
	sdkconfig, _ := provider.AuthorizeSdk()
	api := platformclientv2.NewRoutingApiWithConfig(sdkconfig)

	divisions, resp, err := api.GetRoutingSkillgroupMembersDivisions(resourceId, "")

	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_routing_skill_group", fmt.Sprintf("Failed to update Routing Utilization %s error: %s", resourceId, err), resp)
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
	sdkConfig, _ := provider.AuthorizeSdk()
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
