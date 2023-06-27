package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func TestAccResourceArchitectEmergencyGroups(t *testing.T) {
	t.Parallel()
	var (
		resourceType = "genesyscloud_architect_emergencygroup"
		resourceName = "test_emergency_group"
		name         = "Test Group " + uuid.NewString()
		description  = "The test description"

		updatedDescription = description + " updated"

		ivrResourceID = "test-ivr"
		ivrName       = "Test IVR " + uuid.NewString()

		flowResource      = "test_flow"
		flowName          = "Terraform Test Flow " + uuid.NewString()
		flowFilePath      = "../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
		inboundCallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  ivrResourceID,
					name:        ivrName,
					description: "",
					dnis:        nil,
					depends_on:  "",
				}) + GenerateFlowResource(
					flowResource,
					flowFilePath,
					inboundCallConfig,
					false,
				) + generateArchitectEmergencyGroupResource(
					resourceName,
					name,
					nullValue,
					description,
					trueValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResource+".id", "genesyscloud_architect_ivr."+ivrResourceID+".id"),
				) + generateFlowDataSource(
					"flow",
					"genesyscloud_flow."+flowResource,
					flowName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", trueValue),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.emergency_flow_id",
						"data.genesyscloud_flow.flow", "id"),
				),
			},
			{
				// Update
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  ivrResourceID,
					name:        ivrName,
					description: "",
					dnis:        nil,
					depends_on:  "",
				}) + GenerateFlowResource(
					flowResource,
					flowFilePath,
					inboundCallConfig,
					false,
				) + generateArchitectEmergencyGroupResource(
					resourceName,
					name,
					nullValue,
					updatedDescription,
					falseValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResource+".id", "genesyscloud_architect_ivr."+ivrResourceID+".id"),
				) + generateFlowDataSource(
					"flow",
					"genesyscloud_flow."+flowResource,
					flowName,
				) + generateIvrDataSource(
					"ivr",
					strconv.Quote(ivrName),
					"genesyscloud_architect_ivr."+ivrResourceID,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", falseValue),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.emergency_flow_id",
						"data.genesyscloud_flow.flow", "id"),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.ivr_ids.0",
						"data.genesyscloud_architect_ivr.ivr", "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceType + "." + resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEmergencyGroupDestroyed,
	})
}

func generateArchitectEmergencyGroupResource(
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
		} else if IsStatus404(resp) {
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
