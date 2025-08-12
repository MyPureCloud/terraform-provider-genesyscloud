package integration_action

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	integration "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The resource_genesyscloud_integration_action_test.go contains all of the test cases for running the resource
tests for integration_actions.
*/
func TestAccResourceIntegrationAction(t *testing.T) {
	var (
		integResourceLabel1 = "test_integration1"
		integTypeID         = "purecloud-data-actions"

		actionResourceLabel1 = "test-action1"
		actionName1          = "Terraform Action1-" + uuid.NewString()
		actionName2          = "Terraform Action2-" + uuid.NewString()
		actionCateg1         = "Genesys Cloud Data Actions"
		actionCateg2         = "Genesys Cloud Data Actions 2"

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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create an integration and an associated action
				Config: integration.GenerateIntegrationResource(
					integResourceLabel1,
					util.NullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResourceLabel1,
					actionName1,
					actionCateg1,
					"genesyscloud_integration."+integResourceLabel1+".id",
					util.NullValue, // Secure default (false)
					util.NullValue, // Timeout default
					util.GenerateJsonSchemaDocStr(inputAttr1),  // contract_input
					util.GenerateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate1,
						reqType1,
						util.NullValue, // Default req templatezz
						"",             // No headers
					),
					// Default config response
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "name", actionName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "category", actionCateg1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "secure", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_timeout_seconds", "0"),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),
					util.ValidateValueInJsonAttr("genesyscloud_integration_action."+actionResourceLabel1, "contract_input", "type", "object"),
					util.ValidateValueInJsonAttr("genesyscloud_integration_action."+actionResourceLabel1, "contract_input", "properties."+inputAttr1+".type", "string"),
					util.ValidateValueInJsonAttr("genesyscloud_integration_action."+actionResourceLabel1, "contract_input", "required", inputAttr1),
					util.ValidateValueInJsonAttr("genesyscloud_integration_action."+actionResourceLabel1, "contract_output", "type", "object"),
					util.ValidateValueInJsonAttr("genesyscloud_integration_action."+actionResourceLabel1, "contract_output", "properties."+outputAttr1+".type", "string"),
					util.ValidateValueInJsonAttr("genesyscloud_integration_action."+actionResourceLabel1, "contract_output", "required", outputAttr1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_url_template", reqUrlTemplate1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_type", reqType1),
				),
			},
			{
				// Update action name, category, timeout, and request/response config
				Config: integration.GenerateIntegrationResource(
					integResourceLabel1,
					util.NullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResourceLabel1,
					actionName2,
					actionCateg2,
					"genesyscloud_integration."+integResourceLabel1+".id",
					util.NullValue, // Secure default (false)
					timeout2,
					util.GenerateJsonSchemaDocStr(inputAttr1),  // contract_input
					util.GenerateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate2,
						reqType2,
						strconv.Quote(reqTemp),
						util.GenerateMapAttrWithMapProperties(
							"headers",
							map[string]string{
								headerKey: strconv.Quote(headerVal1),
							},
						),
					),
					generateIntegrationActionConfigResponse(
						strconv.Quote(successTemplate),
						util.GenerateMapAttrWithMapProperties(
							"translation_map",
							map[string]string{
								transMapAttr: strconv.Quote(transMapVal1),
							},
						),
						util.GenerateMapAttrWithMapProperties(
							"translation_map_defaults",
							map[string]string{
								transMapAttr: strconv.Quote(transMapValDefault1),
							},
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "name", actionName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "category", actionCateg2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "secure", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_timeout_seconds", timeout2),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_url_template", reqUrlTemplate2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_type", reqType2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_template", strings.ReplaceAll(reqTemp, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.headers."+headerKey, headerVal1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_response.0.success_template", strings.ReplaceAll(successTemplate, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_response.0.translation_map."+transMapAttr, transMapVal1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_response.0.translation_map_defaults."+transMapAttr, transMapValDefault1),
				),
			},
			{
				// Update config values as well as secure field which should force a new action to be created
				Config: integration.GenerateIntegrationResource(
					integResourceLabel1,
					util.NullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResourceLabel1,
					actionName2,
					actionCateg2,
					"genesyscloud_integration."+integResourceLabel1+".id",
					util.TrueValue, // Secure
					util.NullValue, // time default
					util.GenerateJsonSchemaDocStr(inputAttr1),  // contract_input
					util.GenerateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate2,
						reqType2,
						strconv.Quote(reqTemp),
						util.GenerateMapAttrWithMapProperties(
							"headers",
							map[string]string{
								headerKey: strconv.Quote(headerVal2),
							},
						),
					),
					generateIntegrationActionConfigResponse(
						strconv.Quote(successTemplate),
						util.GenerateMapAttrWithMapProperties(
							"translation_map",
							map[string]string{
								transMapAttr: strconv.Quote(transMapVal2),
							},
						),
						util.GenerateMapAttrWithMapProperties(
							"translation_map_defaults",
							map[string]string{
								transMapAttr: strconv.Quote(transMapValDefault2),
							},
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "name", actionName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "category", actionCateg2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "secure", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_url_template", reqUrlTemplate2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_type", reqType2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_template", strings.ReplaceAll(reqTemp, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.headers."+headerKey, headerVal2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_response.0.success_template", strings.ReplaceAll(successTemplate, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_response.0.translation_map."+transMapAttr, transMapVal2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_response.0.translation_map_defaults."+transMapAttr, transMapValDefault2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration_action." + actionResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIntegrationActionDestroyed,
	})
}

func TestAccResourceIntegrationActionFunctionData(t *testing.T) {
	// Create temporary zip files for testing
	zipPath1, _ := createTempTestZipFile(t, "test_function1.zip")
	zipPath2, _ := createTempTestZipFile(t, "test_function2.zip")

	var (
		integResourceLabel1 = "test_integration_function1"
		integTypeID         = "function-data-actions"

		actionResourceLabel1 = "test-function-action1"
		actionName1          = "Terraform Function Action1-" + uuid.NewString()
		actionName2          = "Terraform Function Action2-" + uuid.NewString()
		actionCateg1         = "Function Data Actions"
		actionCateg2         = "Function Data Actions"

		timeout1 = "15"
		timeout2 = "10"

		inputAttr1  = "service"
		outputAttr1 = "status"

		// Function configuration values
		description1 = "Custom function for data processing"
		description2 = "Updated function for advanced data processing"
		handler1     = "index.handler"
		handler2     = "src/main.handler"
		runtime1     = "nodejs22.x"
		runtime2     = "nodejs22.x"
		filePath1    = zipPath1
		filePath2    = zipPath2
		// publish field is not in the schema, so removing these
		// publish1         = "true"
		// publish2         = "false"
		headerVal2 = "no-store"

		// Request/Response configuration values
		reqUrlTemplate1 = "/api/v2/users"
		reqUrlTemplate2 = "/api/v2/integrations"
		reqType1        = "GET"
		reqType2        = "PUT"
		reqTemp         = "{ \\\"service\\\": \\\"$${input.service}\\\" }"
		headerKey       = "Cache-Control"
		//headerVal2          = "no-store"
		successTemplate = "{ \\\"name\\\": $${nameValue}, \\\"build\\\": $${buildNumber} }"
		transMapAttr    = "nameValue"
		//transMapVal2        = "$.NewName"
		//transMapValDefault2 = "NotKnown"

		transMapVal2        = "$.NewName"
		transMapValDefault2 = "NotKnown"
	)

	// Note: This test validates the schema structure for function_config
	// The actual file validation will be tested in integration tests

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),

		Steps: []resource.TestStep{
			{
				// Create an integration and an associated function data action
				Config: integration.GenerateIntegrationResource(
					integResourceLabel1,
					util.NullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResourceLabel1,
					actionName1,
					actionCateg1,
					"genesyscloud_integration."+integResourceLabel1+".id",
					util.NullValue, // Secure default (false)
					util.NullValue, // Timeout default
					util.GenerateJsonSchemaDocStr(inputAttr1),  // contract_input
					util.GenerateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate1, // Use the variable to fix linter error
						reqType1,
						util.NullValue, // Default req template
						"headers = {}", // Empty headers object as API expects
					),
					generateIntegrationActionConfigResponse(
						util.NullValue,                  // Default success template
						"translation_map = {}",          // Empty translation_map as API expects
						"translation_map_defaults = {}", // Empty translation_map_defaults as API expects
					),
					generateIntegrationActionFunctionConfig(
						strconv.Quote(description1),
						strconv.Quote(handler1),
						strconv.Quote(runtime1),
						timeout1,
						strconv.Quote(filePath1),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "name", actionName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "category", actionCateg1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "secure", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_timeout_seconds", "0"),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),
					// For function data actions, config_request and config_response are not set by the API
					// so we don't check for them
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.description", description1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.handler", handler1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.runtime", runtime1),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.timeout_seconds", timeout1),
					// file_path and file_content_hash are input-only fields not returned by the API
					// so we can't verify them in the state
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Update config values as well as secure field which should force a new action to be created
				Config: integration.GenerateIntegrationResource(
					integResourceLabel1,
					util.NullValue,
					strconv.Quote(integTypeID),
				) + generateIntegrationActionResource(
					actionResourceLabel1,
					actionName2,
					actionCateg2,
					"genesyscloud_integration."+integResourceLabel1+".id",
					util.TrueValue, // Secure
					util.NullValue, // time default
					util.GenerateJsonSchemaDocStr(inputAttr1),  // contract_input
					util.GenerateJsonSchemaDocStr(outputAttr1), // contract_output
					generateIntegrationActionConfigRequest(
						reqUrlTemplate2,
						reqType2,
						strconv.Quote(reqTemp),
						util.GenerateMapAttrWithMapProperties(
							"headers",
							map[string]string{
								headerKey: strconv.Quote(headerVal2),
							},
						),
					),
					generateIntegrationActionConfigResponse(
						strconv.Quote(successTemplate),
						util.GenerateMapAttrWithMapProperties(
							"translation_map",
							map[string]string{
								transMapAttr: strconv.Quote(transMapVal2),
							},
						),
						util.GenerateMapAttrWithMapProperties(
							"translation_map_defaults",
							map[string]string{
								transMapAttr: strconv.Quote(transMapValDefault2),
							},
						),
					),
					generateIntegrationActionFunctionConfig(
						strconv.Quote(description2),
						strconv.Quote(handler2),
						strconv.Quote(runtime2),
						timeout2,
						strconv.Quote(filePath2),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "name", actionName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "category", actionCateg2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "secure", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_type", reqType2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "config_request.0.request_template", strings.ReplaceAll(reqTemp, "$${", "${")),

					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.description", description2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.handler", handler2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.runtime", runtime2),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+actionResourceLabel1, "function_config.0.timeout_seconds", timeout2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration_action." + actionResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIntegrationActionDestroyed,
	})
}

func generateIntegrationActionResource(resourceLabel, name, category, integId, secure, timeout, contractIn, contractOut string, blocks ...string) string {
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
	`, resourceLabel, name, category, integId, secure, timeout, contractIn, contractOut, strings.Join(blocks, "\n"))
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

func generateIntegrationActionFunctionConfig(description, handler, runtime, timeoutSeconds, filePath string) string {
	return fmt.Sprintf(`function_config {
        description = %s
        handler = %s
        runtime = %s
        timeout_seconds = %s
        file_path = %s
	}
	`, description, handler, runtime, timeoutSeconds, filePath)
}

// createTempTestZipFile creates a temporary zip file with some test content for testing purposes
func createTempTestZipFile(t *testing.T, filename string) (string, string) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test_integration_action_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create the zip file path
	zipPath := filepath.Join(tempDir, filename)

	// Create the zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	// Create a zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add some test files to the zip
	testFiles := map[string]string{
		"index.js": `exports.handler = async function(event, context) {
    console.log('Test function executed');
    return { statusCode: 200, body: 'Hello from test function!' };
};`,
		"package.json": `{
    "name": "test-function",
    "version": "1.0.0",
    "main": "index.js",
    "dependencies": {}
}`,
		"README.md": `# Test Function
This is a test function for integration action testing.`,
	}

	// Add each file to the zip
	for fileName, content := range testFiles {
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}

		_, err = io.WriteString(fileWriter, content)
		if err != nil {
			t.Fatalf("Failed to write content to zip file: %v", err)
		}
	}

	// Calculate a simple hash for the file content
	// In a real scenario, this would be a proper hash, but for testing we'll use a simple one
	fileHash := "sha256:test_hash_" + filename + "_" + uuid.NewString()[:8]

	// Clean up function to remove temp files after test
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return zipPath, fileHash
}

func testVerifyIntegrationActionDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_action" {
			continue
		}

		action, resp, err := integrationAPI.GetIntegrationsAction(rs.Primary.ID, "", false, false)
		if action != nil {
			return fmt.Errorf("Integration action (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
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
