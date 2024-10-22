package architect_emergencygroup

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	architectIvr "terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v143/platformclientv2"
)

func TestAccResourceArchitectEmergencyGroups(t *testing.T) {
	var (
		resourceType = "genesyscloud_architect_emergencygroup"
		resourceName = "test_emergency_group"
		name         = "Test Group " + uuid.NewString()
		description  = "The test description"

		updatedDescription = description + " updated"

		ivrConfigResource1 = "test-ivrconfig1"
		ivrConfigName      = "terraform-ivrconfig-" + uuid.NewString()

		flowResource      = "test_flow"
		flowName          = "Terraform Emergency Test Flow " + uuid.NewString()
		flowFilePath      = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
		inboundCallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	ivrResource := architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
		ResourceID:  ivrConfigResource1,
		Name:        ivrConfigName,
		Description: "",
		Dnis:        nil, // No dnis
		DependsOn:   "",  // No depends_on
	})

	flowResourceConfig := architect_flow.GenerateFlowResource(
		flowResource,
		flowFilePath,
		inboundCallConfig,
		false,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: flowResourceConfig + ivrResource + GenerateArchitectEmergencyGroupResource(
					resourceName,
					name,
					util.NullValue,
					description,
					util.TrueValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResource+".id", "genesyscloud_architect_ivr."+ivrConfigResource1+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", util.TrueValue),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.ivr_ids.0",
						"genesyscloud_architect_ivr."+ivrConfigResource1, "id"),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResource, "id"),
				),
			},
			{
				// Update
				Config: flowResourceConfig + ivrResource + GenerateArchitectEmergencyGroupResource(
					resourceName,
					name,
					util.NullValue,
					updatedDescription,
					util.FalseValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResource+".id", "genesyscloud_architect_ivr."+ivrConfigResource1+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", util.FalseValue),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.emergency_flow_id",
						"genesyscloud_flow."+flowResource, "id"),
					resource.TestCheckResourceAttrPair(resourceType+"."+resourceName, "emergency_call_flows.0.ivr_ids.0",
						"genesyscloud_architect_ivr."+ivrConfigResource1, "id"),
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

func TestAccResourceArchitectEmergencyGroupMultipleIvrs(t *testing.T) {
	var (
		resourceType = "genesyscloud_architect_emergencygroup"
		resourceName = "test_emergency_group"
		name         = "Test Group " + uuid.NewString()
		description  = "The test description"

		ivrResourceName1 = "test-ivrconfig1"
		ivrName1         = "terraform-ivrconfig-1-" + uuid.NewString()

		ivrResourceName2 = "test-ivrconfig2"
		ivrName2         = "terraform-ivrconfig-2-" + uuid.NewString()

		flowResourceName1  = "test_flow1"
		flowName1          = "Terraform Emergency Test Flow 1 " + uuid.NewString()
		flowFilePath1      = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
		inboundCallConfig1 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName1)

		flowResourceName2  = "test_flow2"
		flowName2          = "Terraform Emergency Test Flow 2 " + uuid.NewString()
		flowFilePath2      = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example2.yaml"
		inboundCallConfig2 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName2)
	)

	ivrResource1 := architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
		ResourceID:  ivrResourceName1,
		Name:        ivrName1,
		Description: "",
		Dnis:        nil, // No dnis
		DependsOn:   "",  // No depends_on
	})

	ivrResource2 := architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
		ResourceID:  ivrResourceName2,
		Name:        ivrName2,
		Description: "",
		Dnis:        nil, // No dnis
		DependsOn:   "",  // No depends_on
	})

	flowResourceConfig1 := architect_flow.GenerateFlowResource(
		flowResourceName1,
		flowFilePath1,
		inboundCallConfig1,
		false,
	)

	flowResourceConfig2 := architect_flow.GenerateFlowResource(
		flowResourceName2,
		flowFilePath2,
		inboundCallConfig2,
		false,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: flowResourceConfig1 + ivrResource1 + GenerateArchitectEmergencyGroupResource(
					resourceName,
					name,
					util.NullValue,
					description,
					util.TrueValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceName1+".id", "genesyscloud_architect_ivr."+ivrResourceName1+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "emergency_call_flows.#", "1"),
					validateEmergencyCallFlows(
						resourceType+"."+resourceName,
						"genesyscloud_flow."+flowResourceName1,
						"genesyscloud_architect_ivr."+ivrResourceName1,
					),
				),
			},
			{
				Config: flowResourceConfig1 + flowResourceConfig2 + ivrResource1 + ivrResource2 + GenerateArchitectEmergencyGroupResource(
					resourceName,
					name,
					util.NullValue,
					description,
					util.TrueValue,
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceName1+".id", "genesyscloud_architect_ivr."+ivrResourceName1+".id"),
					generateEmergencyCallFlow("genesyscloud_flow."+flowResourceName2+".id", "genesyscloud_architect_ivr."+ivrResourceName2+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr(resourceType+"."+resourceName, "emergency_call_flows.#", "2"),
					validateEmergencyCallFlows(
						resourceType+"."+resourceName,
						"genesyscloud_flow."+flowResourceName1,
						"genesyscloud_architect_ivr."+ivrResourceName1,
					),
					validateEmergencyCallFlows(
						resourceType+"."+resourceName,
						"genesyscloud_flow."+flowResourceName2,
						"genesyscloud_architect_ivr."+ivrResourceName2,
					),
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

func validateEmergencyCallFlows(emergencyGroupResourceName, flowResourceName, architectIvrResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// Get emergency group form state
		emergencyGroupResource, ok := state.RootModule().Resources[emergencyGroupResourceName]
		if !ok {
			return fmt.Errorf("failed to find emergency group %s in state", emergencyGroupResourceName)
		}

		// Get IVR resource form state
		ivrResource, ok := state.RootModule().Resources[architectIvrResourceName]
		if !ok {
			return fmt.Errorf("failed to find architect IVR %s in state", architectIvrResourceName)
		}
		ivrId := ivrResource.Primary.ID

		// Get flow resource form state
		flowResource, ok := state.RootModule().Resources[flowResourceName]
		if !ok {
			return fmt.Errorf("failed to find flow %s in state", flowResourceName)
		}
		flowId := flowResource.Primary.ID

		callFlowsAttr, ok := emergencyGroupResource.Primary.Attributes["emergency_call_flows.#"]
		if !ok {
			return fmt.Errorf("no emegrncy call flows found for emergency group %s", emergencyGroupResource.Primary.ID)
		}

		callFlowNumber, _ := strconv.Atoi(callFlowsAttr)
		for i := 0; i < callFlowNumber; i++ {
			iterationString := strconv.Itoa(i)
			if emergencyGroupResource.Primary.Attributes["emergency_call_flows."+iterationString+".emergency_flow_id"] == flowId {
				if emergencyGroupResource.Primary.Attributes["emergency_call_flows."+iterationString+".ivr_ids.0"] == ivrId {
					// emergency group was created with correct flowId and IvrId
					return nil
				}

				actualIvr := emergencyGroupResource.Primary.Attributes["emergency_call_flows."+iterationString+".ivr_ids.0"]
				return fmt.Errorf("flow %s found in emergency call flows with incorrect Ivr. \nexpected: %s, \nactual: %s", flowId, ivrId, actualIvr)
			}
		}

		return fmt.Errorf("flow %s not found in emergency call flows", flowId)
	}
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
