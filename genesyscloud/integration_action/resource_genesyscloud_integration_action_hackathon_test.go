package integration_action

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	integration "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"
	integrationCred "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_credential"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

/*
TestAccResourceHackathon2026IntegrationAndActions creates an ENABLED purecloud-data-actions
integration and three associated data actions, matching the hackathon_2026 example resources.
*/
func TestAccResourceHackathon2026IntegrationAndActions(t *testing.T) {
	var (
		credResourceLabel = "hackathon_2026_credential"
		credName          = "Hackathon 2026 Credential-" + uuid.NewString()
		credTypeName      = "pureCloudOAuthClient"

		integResourceLabel = "hackathon_2026_integration"
		integName          = "Hackathon 2026 Integration-" + uuid.NewString()
		integTypeID        = "purecloud-data-actions"
		enabledState       = "ENABLED"

		firstActionLabel  = "hackathon_2026_first_action"
		secondActionLabel = "hackathon_2026_second_action"
		thirdActionLabel  = "hackathon_2026_third_action"

		firstActionName  = "Hackathon 2026 First Action-" + uuid.NewString()
		secondActionName = "Hackathon 2026 Second Action-" + uuid.NewString()
		thirdActionName  = "Hackathon 2026 Third Action-" + uuid.NewString()

		actionCategory = "Genesys Cloud Data Action"
		timeout        = "20"

		inputAttr1  = "examplestr"
		outputAttr1 = "status"

		reqUrlTemplate = "https://www.example.com/health/check/services/$${input.service}"
		reqType        = "GET"
		reqTemp        = "$${input.rawRequest}"
		headerKey      = "Cache-Control"
		headerVal      = "no-cache"
		successTemp    = "{ \\\"name\\\": $${nameValue}, \\\"build\\\": $${buildNumber} }"
	)

	config := integrationCred.GenerateCredentialResource(
		credResourceLabel,
		strconv.Quote(credName),
		strconv.Quote(credTypeName),
		integrationCred.GenerateCredentialFields(
			map[string]string{
				"clientId":     strconv.Quote("ASDDHO292DSO2232DA"),
				"clientSecret": strconv.Quote("XXXXXXXXXXXXXX"),
			},
		),
	) + integration.GenerateIntegrationResource(
		integResourceLabel,
		strconv.Quote(enabledState),
		strconv.Quote(integTypeID),
		integration.GenerateIntegrationConfig(
			strconv.Quote(integName),
			util.NullValue,
			util.GenerateMapProperty(credTypeName, "genesyscloud_integration_credential."+credResourceLabel+".id"),
			util.NullValue,
			util.NullValue,
		),
	) + generateHackathonIntegrationActionResource(
		firstActionLabel,
		firstActionName,
		actionCategory,
		"genesyscloud_integration."+integResourceLabel+".id",
		timeout,
		inputAttr1,
		outputAttr1,
		reqUrlTemplate,
		reqType,
		reqTemp,
		headerKey,
		headerVal,
		successTemp,
	) + generateHackathonIntegrationActionResource(
		secondActionLabel,
		secondActionName,
		actionCategory,
		"genesyscloud_integration."+integResourceLabel+".id",
		timeout,
		inputAttr1,
		outputAttr1,
		reqUrlTemplate,
		reqType,
		reqTemp,
		headerKey,
		headerVal,
		successTemp,
	) + generateHackathonIntegrationActionResource(
		thirdActionLabel,
		thirdActionName,
		actionCategory,
		"genesyscloud_integration."+integResourceLabel+".id",
		timeout,
		inputAttr1,
		outputAttr1,
		reqUrlTemplate,
		reqType,
		reqTemp,
		headerKey,
		headerVal,
		successTemp,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+integResourceLabel, "intended_state", enabledState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+integResourceLabel, "integration_type", integTypeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+integResourceLabel, "config.0.name", integName),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_integration."+integResourceLabel, "config.0.credentials."+credTypeName,
						"genesyscloud_integration_credential."+credResourceLabel, "id",
					),

					resource.TestCheckResourceAttr("genesyscloud_integration_action."+firstActionLabel, "name", firstActionName),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+firstActionLabel, "category", actionCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+firstActionLabel, "secure", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+firstActionLabel, "config_timeout_seconds", timeout),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_integration_action."+firstActionLabel, "integration_id",
						"genesyscloud_integration."+integResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+firstActionLabel, "config_request.0.request_type", reqType),

					resource.TestCheckResourceAttr("genesyscloud_integration_action."+secondActionLabel, "name", secondActionName),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+secondActionLabel, "category", actionCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+secondActionLabel, "secure", util.TrueValue),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_integration_action."+secondActionLabel, "integration_id",
						"genesyscloud_integration."+integResourceLabel, "id",
					),

					resource.TestCheckResourceAttr("genesyscloud_integration_action."+thirdActionLabel, "name", thirdActionName),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+thirdActionLabel, "category", actionCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration_action."+thirdActionLabel, "secure", util.TrueValue),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_integration_action."+thirdActionLabel, "integration_id",
						"genesyscloud_integration."+integResourceLabel, "id",
					),
				),
			},
			{
				ResourceName:      "genesyscloud_integration." + integResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "genesyscloud_integration_action." + firstActionLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "genesyscloud_integration_action." + secondActionLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "genesyscloud_integration_action." + thirdActionLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyHackathon2026ResourcesDestroyed,
	})
}

func generateHackathonIntegrationActionResource(
	resourceLabel,
	name,
	category,
	integId,
	timeout,
	inputAttr,
	outputAttr,
	reqUrlTemplate,
	reqType,
	reqTemp,
	headerKey,
	headerVal,
	successTemp string,
) string {
	return generateIntegrationActionResource(
		resourceLabel,
		name,
		category,
		integId,
		util.TrueValue,
		timeout,
		util.GenerateJsonSchemaDocStr(inputAttr),
		util.GenerateJsonSchemaDocStr(outputAttr),
		generateIntegrationActionConfigRequest(
			reqUrlTemplate,
			reqType,
			strconv.Quote(reqTemp),
			util.GenerateMapAttrWithMapProperties(
				"headers",
				map[string]string{
					headerKey: strconv.Quote(headerVal),
				},
			),
		),
		generateIntegrationActionConfigResponse(
			strconv.Quote(successTemp),
			util.GenerateMapAttrWithMapProperties(
				"translation_map",
				map[string]string{
					"nameValue":   strconv.Quote("$.Name"),
					"buildNumber": strconv.Quote("$.Build-Version"),
				},
			),
			util.GenerateMapAttrWithMapProperties(
				"translation_map_defaults",
				map[string]string{
					"buildNumber": strconv.Quote("UNKNOWN"),
				},
			),
		),
	)
}

func testVerifyHackathon2026ResourcesDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()

	// Verify integrations are destroyed first (before data actions)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration" {
			continue
		}
		integ, resp, err := integrationAPI.GetIntegration(rs.Primary.ID, 100, 1, "", nil, "", "")
		if integ != nil {
			return fmt.Errorf("Integration (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}

	// Then verify data actions are destroyed
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_action" {
			continue
		}
		action, resp, err := integrationAPI.GetIntegrationsAction(rs.Primary.ID, "", false, false)
		if action != nil {
			return fmt.Errorf("Integration action (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}
