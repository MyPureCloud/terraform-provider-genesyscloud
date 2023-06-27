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

func TestAccResourceIntegrationAction(t *testing.T) {
	var (
		integResource1 = "test_integration1"
		integTypeID    = "purecloud-data-actions"

		actionResource1 = "test-action1"
		actionName1     = "Terraform Action1-" + uuid.NewString()
		actionName2     = "Terraform Action2-" + uuid.NewString()
		actionCateg1    = "Genesys Cloud Data Actions"
		actionCateg2    = "Genesys Cloud Data Actions 2"

		timeout2 = "20"

		inputAttr1  = "service"
		outputAttr1 = "status"

		reqUrlTemplate1     = "/api/v2/users"
		reqUrlTemplate2     = "/api/v2/integrations"
		reqType1            = "GET"
		reqType2            = "PUT"
		reqTemp             = "{ \\\"service\\\": \\\"$${input.service}\\\" }"
		headerKey           = "Cache-Control"
		headerVal1          = "no-cache"
		headerVal2          = "no-store"
		successTemplate     = "{ \\\"name\\\": $${nameValue}, \\\"build\\\": $${buildNumber} }"
		transMapAttr        = "nameValue"
		transMapVal1        = "$.Name"
		transMapVal2        = "$.NewName"
		transMapValDefault1 = "UNKNOWN"
		transMapValDefault2 = "NotKnown"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create an integration and an associated action
				Config: generateIntegrationResource(
					integResource1,
					nullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResource1,
					actionName1,
					actionCateg1,
					"genesyscloud_integration."+integResource1+".id",
					nullValue,                             // Secure default (false)
					nullValue,                             // Timeout default
					generateJsonSchemaDocStr(inputAttr1),  // contract_input
					generateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate1,
						reqType1,
						nullValue, // Default req templatezz
						"",        // No headers
					),
					// Default config response
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "name", actionName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "category", actionCateg1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "secure", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_timeout_seconds", "0"),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResource1, "integration_id", "genesyscloud_integration."+integResource1, "id"),
					validateValueInJsonAttr("genesyscloud_integration_action."+actionResource1, "contract_input", "type", "object"),
					validateValueInJsonAttr("genesyscloud_integration_action."+actionResource1, "contract_input", "properties."+inputAttr1+".type", "string"),
					validateValueInJsonAttr("genesyscloud_integration_action."+actionResource1, "contract_input", "required", inputAttr1),
					validateValueInJsonAttr("genesyscloud_integration_action."+actionResource1, "contract_output", "type", "object"),
					validateValueInJsonAttr("genesyscloud_integration_action."+actionResource1, "contract_output", "properties."+outputAttr1+".type", "string"),
					validateValueInJsonAttr("genesyscloud_integration_action."+actionResource1, "contract_output", "required", outputAttr1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_url_template", reqUrlTemplate1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_type", reqType1),
				),
			},
			{
				// Update action name, category, timeout, and request/response config
				Config: generateIntegrationResource(
					integResource1,
					nullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResource1,
					actionName2,
					actionCateg2,
					"genesyscloud_integration."+integResource1+".id",
					nullValue, // Secure default (false)
					timeout2,
					generateJsonSchemaDocStr(inputAttr1),  // contract_input
					generateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate2,
						reqType2,
						strconv.Quote(reqTemp),
						generateMapAttr(
							"headers",
							generateMapProperty(headerKey, strconv.Quote(headerVal1)),
						),
					),
					generateIntegrationActionConfigResponse(
						strconv.Quote(successTemplate),
						generateMapAttr(
							"translation_map",
							generateMapProperty(transMapAttr, strconv.Quote(transMapVal1)),
						),
						generateMapAttr(
							"translation_map_defaults",
							generateMapProperty(transMapAttr, strconv.Quote(transMapValDefault1)),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "name", actionName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "category", actionCateg2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "secure", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_timeout_seconds", timeout2),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResource1, "integration_id", "genesyscloud_integration."+integResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_url_template", reqUrlTemplate2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_type", reqType2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_template", strings.ReplaceAll(reqTemp, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.headers."+headerKey, headerVal1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_response.0.success_template", strings.ReplaceAll(successTemplate, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_response.0.translation_map."+transMapAttr, transMapVal1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_response.0.translation_map_defaults."+transMapAttr, transMapValDefault1),
				),
			},
			{
				// Update config values as well as secure field which should force a new action to be created
				Config: generateIntegrationResource(
					integResource1,
					nullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResource1,
					actionName2,
					actionCateg2,
					"genesyscloud_integration."+integResource1+".id",
					trueValue,                             // Secure
					nullValue,                             // time default
					generateJsonSchemaDocStr(inputAttr1),  // contract_input
					generateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate2,
						reqType2,
						strconv.Quote(reqTemp),
						generateMapAttr(
							"headers",
							generateMapProperty(headerKey, strconv.Quote(headerVal2)),
						),
					),
					generateIntegrationActionConfigResponse(
						strconv.Quote(successTemplate),
						generateMapAttr(
							"translation_map",
							generateMapProperty(transMapAttr, strconv.Quote(transMapVal2)),
						),
						generateMapAttr(
							"translation_map_defaults",
							generateMapProperty(transMapAttr, strconv.Quote(transMapValDefault2)),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "name", actionName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "category", actionCateg2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "secure", trueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResource1, "integration_id", "genesyscloud_integration."+integResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_url_template", reqUrlTemplate2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_type", reqType2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.request_template", strings.ReplaceAll(reqTemp, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_request.0.headers."+headerKey, headerVal2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_response.0.success_template", strings.ReplaceAll(successTemplate, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_response.0.translation_map."+transMapAttr, transMapVal2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResource1, "config_response.0.translation_map_defaults."+transMapAttr, transMapValDefault2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration_action." + actionResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIntegrationActionDestroyed,
	})
}

func generateIntegrationActionResource(resourceID, name, category, integId, secure, timeout, contractIn, contractOut string, blocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration_action" "%s" {
        name = "%s"
        category = "%s"
        integration_id = %s
        secure = %s
		config_timeout_seconds = %s
        contract_input = %s
        contract_output = %s
        %s
	}
	`, resourceID, name, category, integId, secure, timeout, contractIn, contractOut, strings.Join(blocks, "\n"))
}

func generateIntegrationActionConfigRequest(reqUrlTemplate, reqType, reqTemp, headers string) string {
	return fmt.Sprintf(`config_request {
        request_url_template = "%s"
        request_type = "%s"
        request_template = %s
        %s
	}
	`, reqUrlTemplate, reqType, reqTemp, headers)
}

func generateIntegrationActionConfigResponse(successTemp string, blocks ...string) string {
	return fmt.Sprintf(`config_response {
        success_template = %s
        %s
	}
	`, successTemp, strings.Join(blocks, "\n"))
}

func generateJsonSchemaDocStr(properties ...string) string {
	attrType := "type"
	attrProperties := "properties"
	typeObject := "object"
	typeStr := "string" // All string props

	propStrs := []string{}
	for _, prop := range properties {
		propStrs = append(propStrs, generateJsonProperty(prop, generateJsonObject(
			generateJsonProperty(attrType, strconv.Quote(typeStr)),
		)))
	}
	allProps := strings.Join(propStrs, "\n")

	return generateJsonEncodedProperties(
		// First field is required
		generateJsonArrayProperty("required", strconv.Quote(properties[0])),
		generateJsonProperty(attrType, strconv.Quote(typeObject)),
		generateJsonProperty(attrProperties, generateJsonObject(
			allProps,
		)),
	)
}

func testVerifyIntegrationActionDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_action" {
			continue
		}

		action, resp, err := integrationAPI.GetIntegrationsAction(rs.Primary.ID, "", false)
		if action != nil {
			return fmt.Errorf("Integration action (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
			// Action not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All actions destroyed
	return nil
}
