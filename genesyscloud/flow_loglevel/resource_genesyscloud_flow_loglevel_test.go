package flow_loglevel

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceFlowLogLevel(t *testing.T) {
	var (
		flowResource         = "test_logLevel_flow1"
		resourceId           = "flow_log_level" + uuid.NewString()
		flowName             = "Terraform Test Flow log level " + uuid.NewString()
		flowLoglevelBase     = "Base"
		flowLoglevelAll      = "All"
		flowLogLevelDisabled = "Disabled"
		flowId               = "${genesyscloud_flow." + flowResource + ".id}"
		filePath             = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
		inboundCallConfig    = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	flowResourceConfig := architect_flow.GenerateFlowResource(
		flowResource,
		filePath,
		inboundCallConfig,
		true,
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create using flow log level Base
				Config: flowResourceConfig + generateFlowLogLevelResource(
					flowId,
					flowLoglevelBase,
					resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_loglevel."+resourceId, "flow_log_level", flowLoglevelBase),
				),
			},
			{
				// Update using flow log level All
				Config: flowResourceConfig + generateFlowLogLevelResource(
					flowId,
					flowLoglevelAll,
					resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_loglevel."+resourceId, "flow_log_level", flowLoglevelAll),
				),
			},
			{
				// Update using flow log level Disabled
				Config: flowResourceConfig + generateFlowLogLevelResource(
					flowId,
					flowLogLevelDisabled,
					resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_loglevel."+resourceId, "flow_log_level", flowLogLevelDisabled),
				),
			},
		},
		CheckDestroy: testVerifyFlowLogLevelDestroyed,
	})
}

func testVerifyFlowLogLevelDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_flow_loglevel" {
			continue
		}
		flowLogLevel, resp, err := architectAPI.GetFlowInstancesSettingsLoglevels(rs.Primary.ID, nil)
		if flowLogLevel != nil {
			return fmt.Errorf("flowLogLevel for flowId (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// flow log level not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All low log levels deleted
	return nil
}

func generateFlowLogLevelResource(
	flowId string,
	flowLoglevel string,
	resourceId string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_loglevel" "%s" {
	  flow_id					= "%s"
	  flow_log_level 			= "%s"
	}`,
		resourceId,
		flowId,
		flowLoglevel)
}
