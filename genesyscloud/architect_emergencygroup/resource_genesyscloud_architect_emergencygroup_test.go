package architect_emergencygroup

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceArchitectEmergencyGroups(t *testing.T) {

	var (
		resourceLabel    = "test_emergency_group"
		fullResourcePath = ResourceType + "." + resourceLabel
		name             = "Test Group " + uuid.NewString()
		description      = "The test description"

		updatedDescription = description + " updated"

		flowResourceLabel = "test_flow"
		flowName          = "Terraform Emergency Test Flow " + uuid.NewString()
		flowFilePath      = filepath.Join(testrunner.RootDir, "/examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml")
		inboundCallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	flowResourceConfig := architect_flow.GenerateFlowResource(
		flowResourceLabel,
		flowFilePath,
		inboundCallConfig,
		false,
	)

	ivrResourceLabel := "ivr"
	ivrConfig := architect_ivr.IvrConfigStruct{
		ResourceLabel: ivrResourceLabel,
		Name:          "tf test ivr " + uuid.NewString(),
		Dnis:          []string{},
	}
	ivrConfigResource := architect_ivr.GenerateIvrConfigResource(&ivrConfig)
	ivrFullResourcePath := architect_ivr.ResourceType + "." + ivrResourceLabel

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: flowResourceConfig + ivrConfigResource + GenerateArchitectEmergencyGroupResource(
					resourceLabel,
					name,
					util.NullValue,
					description,
					util.TrueValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceLabel+".id", ivrFullResourcePath+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourcePath, "name", name),
					resource.TestCheckResourceAttr(fullResourcePath, "description", description),
					resource.TestCheckResourceAttr(fullResourcePath, "enabled", util.TrueValue),
					resource.TestCheckResourceAttrPair(fullResourcePath, "emergency_call_flows.0.ivr_ids.0",
						ivrFullResourcePath, "id"),
					resource.TestCheckResourceAttrPair(fullResourcePath, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResourceLabel, "id"),
				),
			},
			{
				// Update
				Config: ivrConfigResource + flowResourceConfig + GenerateArchitectEmergencyGroupResource(
					resourceLabel,
					name,
					util.NullValue,
					updatedDescription,
					util.FalseValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceLabel+".id", ivrFullResourcePath+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourcePath, "name", name),
					resource.TestCheckResourceAttr(fullResourcePath, "description", updatedDescription),
					resource.TestCheckResourceAttr(fullResourcePath, "enabled", util.FalseValue),
					resource.TestCheckResourceAttrPair(fullResourcePath, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(fullResourcePath, "emergency_call_flows.0.ivr_ids.0",
						ivrFullResourcePath, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      fullResourcePath,
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
