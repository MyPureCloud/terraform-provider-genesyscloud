package genesyscloud

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
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
		resource := string(r)
		target := string(t)
		if resource != target {
			return fmt.Errorf("resource skill_conditions does not match skill_conditions passed in. Expected: %s Actual: %s", resource, target)
		}

		return nil
	}
}

func TestAccResourceRoutingSkillGroupBasic(t *testing.T) {
	var (
		skillGroupResource     = "testskillgroup1"
		skillGroupName1        = "SkillGroup1" + uuid.NewString()
		skillGroupDescription1 = "Description1" + uuid.NewString()
		skillGroupName2        = "SkillGroup2" + uuid.NewString()
		skillGroupDescription2 = "Description2" + uuid.NewString()
		divHomeRes             = "auth-division-home"
		divHomeName            = "Home"
		homeDesc               = "Home"
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

	config1 := generateAuthDivisionResource(
		divHomeRes,
		divHomeName,
		strconv.Quote(homeDesc),
		trueValue, // Home division
	) +
		generateRoutingSkillGroupResource(
			skillGroupResource,
			"genesyscloud_auth_division."+divHomeRes,
			skillGroupName1,
			skillGroupDescription1,
			"genesyscloud_auth_division."+divHomeRes+".id",
			skillCondition1,
		)

	config2 := generateAuthDivisionResource(
		divHomeRes,
		divHomeName,
		strconv.Quote(homeDesc),
		trueValue, // Home division
	) +
		generateRoutingSkillGroupResource(
			skillGroupResource,
			"genesyscloud_auth_division."+divHomeRes,
			skillGroupName2,
			skillGroupDescription2,
			"genesyscloud_auth_division."+divHomeRes+".id",
			skillCondition2,
		)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription1),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition1),
					testDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),
				),
			},
			{
				// Update
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription2),
					testAccCheckSkillConditions("genesyscloud_routing_skill_group."+skillGroupResource, skillCondition2),
					testDefaultHomeDivision("genesyscloud_routing_skill_group."+skillGroupResource),
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

func generateRoutingSkillGroupResource(
	resourceID string,
	divisionResourceName string,
	name string,
	description string,
	divisionID string,
	skillCondition string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill_group" "%s" {
		depends_on = [%s]
		name = "%s"
		description="%s"
		division_id=%s
		skill_conditions = jsonencode(%s)
	}
	`, resourceID, divisionResourceName, name, description, divisionID, skillCondition)
}

func generateRoutingSkillGroupResourceBasic(
	resourceID string,
	name string,
	description string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill_group" "%s" {
		name = "%s"
		description="%s"
	}
	`, resourceID, name, description)
}

func testVerifySkillGroupDestroyed(state *terraform.State) error {
	// Get default config to set config options
	config := platformclientv2.GetDefaultConfiguration()
	routingAPI := platformclientv2.NewRoutingApi()
	apiClient := &routingAPI.Configuration.APIClient

	// TODO Once this code has been released into the public API we should fix this and use the SDK
	err := config.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return fmt.Errorf("Unexpected error while trying to authorize client in testVerifySkillGroupDestroyed : %s", err)
	}

	headerParams := buildHeaderParams(routingAPI)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_skill_group" {
			continue
		}

		path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups/" + rs.Primary.ID
		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)

		skillGroupPayload := make(map[string]interface{})

		if err != nil {
			if isStatus404(response) {
				break
			}

			return fmt.Errorf("Unexpected error while trying to read skillgroup: %s", err)
		}

		json.Unmarshal(response.RawBody, &skillGroupPayload)

		if skillGroupPayload["id"] != nil && skillGroupPayload["id"] != "" {
			return fmt.Errorf("Skill Group (%s) still exists", rs.Primary.ID)
		}

	}
	// Success. All skills destroyed
	return nil
}
