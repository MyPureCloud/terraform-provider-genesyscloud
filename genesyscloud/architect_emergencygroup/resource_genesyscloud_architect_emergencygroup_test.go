package architect_emergencygroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceArchitectEmergencyGroups(t *testing.T) {

	var (
		resourceLabel = "test_emergency_group"
		name          = "Test Group " + uuid.NewString()
		description   = "The test description"

		updatedDescription = description + " updated"

		flowResourceLabel = "test_flow"
		flowName          = "Terraform Emergency Test Flow " + uuid.NewString()
		flowFilePath      = filepath.Join(testrunner.RootDir, "/examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml")
		inboundCallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Skip("failed to authorize client credentials")
	}

	// TODO: Create the IVR inside the test config once emergency group has been moved to its own package.
	// Currently, the ivr resource cannot be registered for these tests because of a cyclic dependency issue.
	ivrId := "f94e084e-40eb-470b-80d6-0f99cf22d102"
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		ivrId = "770e3998-11b7-4c96-beb8-215b83201c29"
	}

	if !ivrExists(config, ivrId) {
		t.Skip("Skipping because IVR does not exists in the target org.")
	}

	flowResourceConfig := architect_flow.GenerateFlowResource(
		flowResourceLabel,
		flowFilePath,
		inboundCallConfig,
		false,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: flowResourceConfig + GenerateArchitectEmergencyGroupResource(
					resourceLabel,
					name,
					util.NullValue,
					description,
					util.TrueValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceLabel+".id", strconv.Quote(ivrId)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "enabled", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "emergency_call_flows.0.ivr_ids.0", ivrId),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResourceLabel, "id"),
				),
			},
			{
				// Update
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel,
					flowFilePath,
					inboundCallConfig,
					false,
				) + GenerateArchitectEmergencyGroupResource(
					resourceLabel,
					name,
					util.NullValue,
					updatedDescription,
					util.FalseValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceLabel+".id", strconv.Quote(ivrId)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", updatedDescription),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "enabled", util.FalseValue),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResourceLabel, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "emergency_call_flows.0.ivr_ids.0", ivrId),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEmergencyGroupDestroyed,
	})
}

func GenerateArchitectEmergencyGroupResource(
	eGroupResource string,
	name string,
	divisionId string,
	description string,
	enabled string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_emergencygroup" "%s" {
		name        = "%s"
		division_id = %s
		description = "%s"
		enabled     = %s
		%s
	}
	`, eGroupResource, name, divisionId, description, enabled, strings.Join(nestedBlocks, "\n"))
}

func generateEmergencyCallFlow(flowID string, ivrIDs ...string) string {
	return fmt.Sprintf(`emergency_call_flows {
	emergency_flow_id = %s
	ivr_ids           = [%s]
}
`, flowID, strings.Join(ivrIDs, ", "))
}

func testVerifyEmergencyGroupDestroyed(state *terraform.State) error {
	archAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_emergencygroup" {
			continue
		}

		eGroup, resp, err := archAPI.GetArchitectEmergencygroup(rs.Primary.ID)
		if eGroup != nil {
			return fmt.Errorf("emergency group (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Emergency group not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All emergency groups destroyed
	return nil
}

// TODO Remove the below function when emergency_group is moved to its own package
func ivrExists(config *platformclientv2.Configuration, ivrId string) bool {
	api := platformclientv2.NewArchitectApiWithConfig(config)
	if _, _, err := api.GetArchitectIvr(ivrId); err != nil {
		return false
	}
	return true
}
