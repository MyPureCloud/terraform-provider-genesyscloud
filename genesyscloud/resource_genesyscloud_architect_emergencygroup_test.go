package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceArchitectEmergencyGroups(t *testing.T) {
	t.Parallel()
	var (
		resourceType = "genesyscloud_architect_emergencygroup"
		resourceName = "test_emergency_group"
		name         = "Test Group " + uuid.NewString()
		description  = "The test description"

		updatedDescription = description + " updated"

		flowResource      = "test_flow"
		flowName          = "Terraform Test Flow " + uuid.NewString()
		flowFilePath      = "../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
		inboundCallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	config, err := AuthorizeSdk()
	if err != nil {
		t.Skip("failed to authorize client credentials")
	}

	// TODO: Create the IVR inside the test config once emergency group has been moved to its own package.
	// Currently, the ivr resource cannot be registered for these tests because of a cyclic dependency issue.
	ivrId := "f94e084e-40eb-470b-80d6-0f99cf22d102"
	if !ivrExists(config, ivrId) {
		t.Skip("Skipping because IVR does not exists in the target org.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateFlowResource(
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
					generateEmergencyCallFlow("genesyscloud_flow."+flowResource+".id", strconv.Quote(ivrId)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", trueValue),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "emergency_call_flows.0.ivr_ids.0", ivrId),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResource, "id"),
				),
			},
			{
				// Update
				Config: GenerateFlowResource(
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
					generateEmergencyCallFlow("genesyscloud_flow."+flowResource+".id", strconv.Quote(ivrId)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", falseValue),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResource, "id"),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "emergency_call_flows.0.ivr_ids.0", ivrId),
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

// TODO Remove the below function when emergency_group is moved to its own package
func ivrExists(config *platformclientv2.Configuration, ivrId string) bool {
	api := platformclientv2.NewArchitectApiWithConfig(config)
	if _, _, err := api.GetArchitectIvr(ivrId); err != nil {
		return false
	}
	return true
}
