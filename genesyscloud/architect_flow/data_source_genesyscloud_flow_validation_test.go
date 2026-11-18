package architect_flow

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Tests the genesyscloud_flow_validation data source against a published flow, selected both
// by flow_id and by name and type, and checks the schema rules that select the flow.
func TestAccDataSourceFlowValidation(t *testing.T) {
	var (
		flowResourceLabel = "test_flow"
		byIdLabel         = "by_id"
		byNameLabel       = "by_name"
		flowName          = "Terraform Flow Validation Test-" + uuid.NewString()
		filePath          = uniqueFlowYamlPath(t, "inboundcall_flow_validation_example")

		inboundcallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)

		flowResourcePath = ResourceType + "." + flowResourceLabel
		byIdPath         = "data." + ValidationDataSourceType + "." + byIdLabel
		byNamePath       = "data." + ValidationDataSourceType + "." + byNameLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Neither flow_id nor name is set
				Config:      fmt.Sprintf(`data "%s" "invalid" {}`, ValidationDataSourceType),
				ExpectError: regexp.MustCompile(`one of\s+` + "`flow_id,name`" + `\s+must be specified`),
			},
			{
				// name without type
				Config:      fmt.Sprintf(`data "%s" "invalid" { name = "some flow" }`, ValidationDataSourceType),
				ExpectError: regexp.MustCompile(`all of\s+` + "`name,type`" + `\s+must be specified`),
			},
			{
				// flow_id together with name and type
				Config: fmt.Sprintf(`data "%s" "invalid" {
					flow_id = "%s"
					name    = "some flow"
					type    = "inboundcall"
				}`, ValidationDataSourceType, uuid.NewString()),
				ExpectError: regexp.MustCompile(`conflicts with flow_id`),
			},
			{
				// Validate the published flow by ID and by name and type, with fail_on_error unset
				Config: GenerateFlowResource(flowResourceLabel, filePath, inboundcallConfig, false) +
					generateFlowValidationDataSourceById(byIdLabel, flowResourcePath+".id") +
					generateFlowValidationDataSourceByName(byNameLabel, flowName, "inboundcall", flowResourcePath),
				Check: resource.ComposeTestCheckFunc(
					checkFlowValidationPassed(byIdPath),
					checkFlowValidationPassed(byNamePath),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func checkFlowValidationPassed(path string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(path, "job_id"),
		resource.TestCheckResourceAttr(path, "status", validateJobStatusSuccess),
		resource.TestCheckResourceAttr(path, "valid", "true"),
		resource.TestCheckResourceAttr(path, "errors.#", "0"),
	)
}

func generateFlowValidationDataSourceById(label, flowIdRef string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		flow_id = %s
	}
	`, ValidationDataSourceType, label, flowIdRef)
}

func generateFlowValidationDataSourceByName(label, name, varType, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name       = %s
		type       = %s
		depends_on = [%s]
	}
	`, ValidationDataSourceType, label, strconv.Quote(name), strconv.Quote(varType), dependsOn)
}
