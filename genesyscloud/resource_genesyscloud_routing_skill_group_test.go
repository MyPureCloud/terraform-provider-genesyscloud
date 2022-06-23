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

func TestAccResourceRoutingSkillGroupBasic(t *testing.T) {
	var (
		skillGroupResource    = "testskillgroup1"
		skillGroupName        = "SkillGroup" + uuid.NewString()
		skillGroupDescription = "Description" + uuid.NewString()
		divHomeRes            = "auth-division-home"
		divHomeName           = "Home"
		homeDesc              = "Home"
		skillCondition        = `[
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
	)

	config := generateAuthDivisionResource(
		divHomeRes,
		divHomeName,
		strconv.Quote(homeDesc),
		trueValue, // Home division
	) +
		generateRoutingSkillGroupResource(
			skillGroupResource,
			"genesyscloud_auth_division."+divHomeRes,
			skillGroupName,
			skillGroupDescription,
			"genesyscloud_auth_division."+divHomeRes+".id",
			skillCondition,
		)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName),
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "description", skillGroupDescription),
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
