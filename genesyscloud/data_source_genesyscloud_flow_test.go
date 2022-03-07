package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFlow(t *testing.T) {
	var (
		flowDataSource    = "flow-data"
		flowName          = "test flow"
		inboundcallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)

		flowResource = "test_flow"
		filePath     = "../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateFlowResource(
					flowResource,
					filePath,
					inboundcallConfig,
				) + generateFlowDataSource(
					flowDataSource,
					"genesyscloud_flow." + flowResource,
					flowName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_flow."+flowDataSource, "id", "genesyscloud_flow."+flowResource, "id"),
				),
			},
		},
	})
}

func generateFlowDataSource(
	resourceID,
	dependsOn,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_flow" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceID, name, dependsOn)
}
