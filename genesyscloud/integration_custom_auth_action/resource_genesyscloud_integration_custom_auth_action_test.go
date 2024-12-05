package integration_custom_auth_action

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/integration"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

type customAuthActionResource struct {
	name           string
	integrationId  string
	configRequest  *customAuthActionResourceConfigRequest
	configResponse *customAuthActionResourceConfigResponse
}

type customAuthActionResourceConfigRequest struct {
	requestUrlTemplate string
	requestType        string
	requestTemplate    string
	headers            map[string]string
}

type customAuthActionResourceConfigResponse struct {
	successTemplate        string
	translationMap         map[string]string
	translationMapDefaults map[string]string
}

/*
The resource_genesyscloud_integration_action_test.go contains all of the test cases for running the resource
tests for integration_custom_auth_actions.
*/
func TestAccResourceIntegrationCustomAuthAction(t *testing.T) {
	var (
		// Integration Credentials
		credentialResourceLabel1   = "test_integration_credential_1"
		credentialResourceTypeAttr = "Terraform Cred-" + uuid.NewString()
		credKey1                   = "loginUrl"
		credVal1                   = "https://www.test-login.com"
		credentialResourceConfig   = integrationCred.GenerateCredentialResource(
			credentialResourceLabel1,
			strconv.Quote(credentialResourceTypeAttr),
			strconv.Quote(customAuthCredentialType),
			integrationCred.GenerateCredentialFields(
				map[string]string{
					credKey1: strconv.Quote(credVal1),
				},
			),
		)

		// Web Services Data Action Integration
		integResourceLabel1       = "test_integration1"
		integResourceTypeAttr1    = "Terraform Integration-" + uuid.NewString()
		integTypeID               = "custom-rest-actions"
		integrationResourceConfig = integration.GenerateIntegrationResource(
			integResourceLabel1,
			util.NullValue,
			strconv.Quote(integTypeID),
			integration.GenerateIntegrationConfig(
				strconv.Quote(integResourceTypeAttr1),
				util.NullValue, // no notes
				fmt.Sprintf("basicAuth = genesyscloud_integration_credential.%s.id", credentialResourceLabel1),
				util.NullValue, // no properties
				util.NullValue, // no advanced properties
			),
		)

		// Custom auth action resource def
		actionResourceLabel1 = "test-auth-action-1"
		requestTemplate1     = "$${input.rawRequest}"
		responseTemplate1    = "{ \\\"name\\\": $${nameValue}, \\\"build\\\": $${buildNumber} }"

		headerKey1 = "Cache-Control"
		headerVal1 = "no-cache"
		headers    = map[string]string{
			headerKey1: strconv.Quote(headerVal1),
		}

		translationMapKey1 = "nameValue"
		translationMapKey2 = "buildNumber"
		translationMapVal1 = "$.Name"
		translationMapVal2 = "$.Build-Version"
		translationMap     = map[string]string{
			translationMapKey1: strconv.Quote(translationMapVal1),
			translationMapKey2: strconv.Quote(translationMapVal2),
		}

		translationMapDefKey1  = "buildNumber"
		translationMapDefVal1  = "UNKNOWN"
		translationMapDefaults = map[string]string{
			translationMapDefKey1: strconv.Quote(translationMapDefVal1),
		}

		oauthActionResource = customAuthActionResource{
			integrationId: "genesyscloud_integration." + integResourceLabel1 + ".id",
			configRequest: &customAuthActionResourceConfigRequest{
				requestUrlTemplate: "https://www.whatever.com/",
				requestType:        "POST",
				requestTemplate:    strconv.Quote(requestTemplate1),
				headers:            headers,
			},
			configResponse: &customAuthActionResourceConfigResponse{
				successTemplate:        strconv.Quote(responseTemplate1),
				translationMap:         translationMap,
				translationMapDefaults: translationMapDefaults,
			},
		}

		oauthActionResource2 = customAuthActionResource{
			name:          "Terraform Action1-" + uuid.NewString(),
			integrationId: "genesyscloud_integration." + integResourceLabel1 + ".id",
			configRequest: &customAuthActionResourceConfigRequest{
				requestUrlTemplate: "https://www.whatever2.com/",
				requestType:        "POST",
				requestTemplate:    strconv.Quote(requestTemplate1),
			},
			configResponse: &customAuthActionResourceConfigResponse{
				successTemplate: strconv.Quote(responseTemplate1),
			},
		}
	)

	config := credentialResourceConfig +
		integrationResourceConfig +
		generateIntegrationCustomAuthActionResource(actionResourceLabel1, &oauthActionResource)

	config2 := credentialResourceConfig +
		integrationResourceConfig +
		generateIntegrationCustomAuthActionResource(actionResourceLabel1, &oauthActionResource2)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Modify the custom auth action of the integration
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),

					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.request_url_template", oauthActionResource.configRequest.requestUrlTemplate),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.request_type", oauthActionResource.configRequest.requestType),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.request_template", strings.ReplaceAll(requestTemplate1, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.headers."+headerKey1, headerVal1),

					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.success_template", strings.ReplaceAll(responseTemplate1, "$${", "${")),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.translation_map."+translationMapKey1, translationMapVal1),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.translation_map."+translationMapKey2, translationMapVal2),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.translation_map_defaults."+translationMapDefKey1, translationMapDefVal1),
				),
			},
			{
				// Change and delete some properties
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "name", oauthActionResource2.name),
					resource.TestCheckResourceAttrPair("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "integration_id", "genesyscloud_integration."+integResourceLabel1, "id"),

					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.request_url_template", oauthActionResource2.configRequest.requestUrlTemplate),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.request_type", oauthActionResource2.configRequest.requestType),
					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.request_template", strings.ReplaceAll(requestTemplate1, "$${", "${")),
					resource.TestCheckNoResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_request.0.headers.%"),

					resource.TestCheckResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.success_template", strings.ReplaceAll(responseTemplate1, "$${", "${")),
					resource.TestCheckNoResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.translation_map.%"),
					resource.TestCheckNoResourceAttr("genesyscloud_integration_custom_auth_action."+actionResourceLabel1, "config_response.0.translation_map_defaults.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration_custom_auth_action." + actionResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
			},
		},
		CheckDestroy: testVerifyIntegrationActionDestroyed,
	})
}

func generateIntegrationCustomAuthActionResource(resourceLabel string, res *customAuthActionResource) string {
	name := ""
	if res.name != "" {
		name = fmt.Sprintf("name = %s", strconv.Quote(res.name))
	}

	return fmt.Sprintf(`resource "genesyscloud_integration_custom_auth_action" "%s" {
        integration_id = %s
        %s
        %s
		%s
	}
	`, resourceLabel, res.integrationId, name,
		generateIntegrationActionConfigRequest(res.configRequest),
		generateIntegrationActionConfigResponse(res.configResponse))
}

func generateIntegrationActionConfigRequest(req *customAuthActionResourceConfigRequest) string {
	headers := ""
	if req.headers != nil {
		headers = util.GenerateMapAttrWithMapProperties("headers", req.headers)
	}

	return fmt.Sprintf(`config_request {
		request_url_template = "%s"
		request_type = "%s"
		request_template = %s
		%s
	}
	`, req.requestUrlTemplate, req.requestType, req.requestTemplate, headers)
}

func generateIntegrationActionConfigResponse(res *customAuthActionResourceConfigResponse) string {
	translationMap := ""
	if res.translationMap != nil {
		translationMap = util.GenerateMapAttrWithMapProperties("translation_map", res.translationMap)
	}

	translationMapDefaults := ""
	if res.translationMapDefaults != nil {
		translationMapDefaults = util.GenerateMapAttrWithMapProperties("translation_map_defaults", res.translationMapDefaults)
	}

	return fmt.Sprintf(`config_response {
		success_template = %s
		%s
		%s
	}
	`, res.successTemplate, translationMap, translationMapDefaults)
}

func testVerifyIntegrationActionDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_custom_auth_action" {
			continue
		}

		action, resp, err := integrationAPI.GetIntegrationsAction(rs.Primary.ID, "", false)
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
