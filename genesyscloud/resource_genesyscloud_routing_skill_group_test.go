package genesyscloud

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

func TestAccResourceRoutingSkillGroupBasic(t *testing.T) {
	var (
		skillGroupResource = "testskillgroup1"
		skillGroupName     = "SkillGroup" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingSkillGroupResource(
					skillGroupResource,
					skillGroupName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_skill_group."+skillGroupResource, "name", skillGroupName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_skill_group." + skillGroupResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySkillsDestroyed,
	})
}

func generateRoutingSkillGroupResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill_group" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}

func testVerifySkillGroupDestroyed(state *terraform.State) error {

	// Get default config to set config options
	config := platformclientv2.GetDefaultConfiguration()
	routingAPI := platformclientv2.NewRoutingApi()
	apiClient := &routingAPI.Configuration.APIClient

	// TODO Once this code has been released into the public API we should fix this and use the SDK
	err := config.AuthorizeClientCredentials(os.Getenv("GENESYS_CLOUD_CLIENT_ID"), os.Getenv("GENESYS_CLOUD_CLIENT_SECRET"))
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
			return fmt.Errorf("Unexpected error while trying to read skillgroup: %s", err)
		}

		if isStatus404(response) {
			continue
		}

		json.Unmarshal(response.RawBody, &skillGroupPayload)

		if skillGroupPayload["id"] != nil && skillGroupPayload["id"] != "" {
			return fmt.Errorf("Skill Group (%s) still exists", rs.Primary.ID)
		}

	}
	// Success. All skills destroyed
	return nil
}
