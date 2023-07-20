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

func TestAccResourceIntegration(t *testing.T) {
	var (
		inteResource1 = "test_integration1"
		inteResource2 = "test_integration2"
		inteName1     = "Terraform Integration Test-" + uuid.NewString()
		inteName2     = "Terraform Integration Test-" + uuid.NewString()

		defaultState = "DISABLED"
		enabledState = "ENABLED"
		configNotes  = "some notes"
		configNotes2 = "This is a note"

		typeID  = "embedded-client-app"
		typeID2 = "custom-smtp-server"

		displayTypeKey  = "displayType"
		sandboxKey      = "sandbox"
		urlKey          = "url"
		groupsKey       = "groups"
		propDisplayType = "standalone"
		propSandbox     = "allow-scripts,allow-same-origin,allow-forms,allow-modals"
		propURL         = "https://mypurecloud.github.io/purecloud-premium-app/wizard/index.html"
		groupResource1  = "test_group"
		groupName       = "terraform integration test group-" + uuid.NewString()
		fakeGroupID     = "123456789"
		emptyJSON       = "{}"

		credResource1 = "test_credential"
		credName1     = "Terraform Credential Test-" + uuid.NewString()
		credTypeName1 = "basicAuth"
		key1          = "userName"
		val1          = "someUserName"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create without config
				Config: generateIntegrationResource(
					inteResource1,
					nullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					// No config block
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState), // Default value would be "DISABLED"
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.properties", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
				),
			},
			{
				// Update only name
				Config: generateIntegrationResource(
					inteResource1,
					nullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					generateIntegrationConfig(
						strconv.Quote(inteName1),
						nullValue, //Empty notes
						"",        //Empty credential ID
						nullValue, //Empty properties
						nullValue, //Empty advanced JSON
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState), // Default value would be "DISABLED"
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.properties", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
				),
			},
			{
				// All nullvalue for config. Nothing should change here.
				Config: generateIntegrationResource(
					inteResource1,
					nullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					generateIntegrationConfig(
						nullValue, // No name update. Should stay the same
						nullValue, //Empty notes
						"",        //Empty credential ID
						nullValue, //Empty properties
						nullValue, //Empty advanced JSON
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState), // Default value would be "DISABLED"
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.properties", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
				),
			},
			{
				// Update intendedState, name, notes, properties
				Config: generateIntegrationResource(
					inteResource1,
					strconv.Quote(enabledState),
					strconv.Quote(typeID),
					generateIntegrationConfig(
						strconv.Quote(inteName2),
						strconv.Quote(configNotes),
						"", //Empty credential ID
						generateJsonEncodedProperties(
							generateJsonProperty(displayTypeKey, strconv.Quote(propDisplayType)),
							generateJsonProperty(urlKey, strconv.Quote(propURL)),
							generateJsonProperty(sandboxKey, strconv.Quote(propSandbox)),
							generateJsonProperty(groupsKey, fmt.Sprintf(`[%s]`, strconv.Quote(fakeGroupID))),
						),
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", enabledState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.name", inteName2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
					validateIntegrationProperties("genesyscloud_integration."+inteResource1, "", propDisplayType, propSandbox, propURL, fakeGroupID),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration." + inteResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Create a group first and use it as reference for a new integration
				Config: generateBasicGroupResource(
					groupResource1,
					groupName,
				) + generateIntegrationResource(
					inteResource1,
					strconv.Quote(enabledState),
					strconv.Quote(typeID),
					generateIntegrationConfig(
						strconv.Quote(inteName1),
						strconv.Quote(configNotes),
						"", //Empty credential ID
						generateJsonEncodedProperties(
							generateJsonProperty(displayTypeKey, strconv.Quote(propDisplayType)),
							generateJsonProperty(urlKey, strconv.Quote(propURL)),
							generateJsonProperty(sandboxKey, strconv.Quote(propSandbox)),
							generateJsonProperty(groupsKey, fmt.Sprintf(`[%s]`, "genesyscloud_group."+groupResource1+".id")),
						),
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", enabledState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
					validateIntegrationProperties("genesyscloud_integration."+inteResource1, "genesyscloud_group."+groupResource1, propDisplayType, propSandbox, propURL, ""),
				),
			},
			{ // Remove the group reference and update intendedState and notes
				Config: generateIntegrationResource(
					inteResource1,
					nullValue, //Change to default value
					strconv.Quote(typeID),
					generateIntegrationConfig(
						strconv.Quote(inteName1),
						strconv.Quote(configNotes2),
						"", //Empty credentials
						generateJsonEncodedProperties(
							generateJsonProperty(displayTypeKey, strconv.Quote(propDisplayType)),
							generateJsonProperty(urlKey, strconv.Quote(propURL)),
							generateJsonProperty(sandboxKey, strconv.Quote(propSandbox)),
							generateJsonProperty(groupsKey, "[]"),
						),
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
					validateIntegrationProperties("genesyscloud_integration."+inteResource1, "", propDisplayType, propSandbox, propURL, ""),
				),
			},
			{ // Update integration name and test Raw JSON string
				Config: generateIntegrationResource(
					inteResource1,
					nullValue, //Default value
					strconv.Quote(typeID),
					generateIntegrationConfig(
						strconv.Quote(inteName2),
						strconv.Quote(configNotes2),
						"", //Empty credentials
						// Use Raw JSON instead of jsonencode function
						fmt.Sprintf(`"{  \"%s\":   \"%s\",  \"%s\": \"%s\",  \"%s\": \"%s\",  \"%s\": %s}"`, displayTypeKey, propDisplayType, urlKey, propURL, sandboxKey, propSandbox, groupsKey, "[]"),
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.name", inteName2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckNoResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials.%"),
					validateIntegrationProperties("genesyscloud_integration."+inteResource1, "", propDisplayType, propSandbox, propURL, ""),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration." + inteResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Create a credential and use it as reference for the new integration
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(credTypeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1)),
					),
				) + generateIntegrationResource(
					inteResource2,
					strconv.Quote(enabledState),
					strconv.Quote(typeID2),
					generateIntegrationConfig(
						strconv.Quote(inteName1),
						strconv.Quote(configNotes),
						generateMapProperty(credTypeName1, "genesyscloud_integration_credential."+credResource1+".id"), // Reference credential ID
						generateJsonEncodedProperties(
							generateJsonProperty("smtpHost", strconv.Quote("fakeHost")),
						),
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "intended_state", enabledState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "integration_type", typeID2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "config.0.name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "config.0.notes", configNotes),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "config.0.advanced", emptyJSON),
					resource.TestCheckResourceAttrPair("genesyscloud_integration."+inteResource2, "config.0.credentials."+credTypeName1, "genesyscloud_integration_credential."+credResource1, "id"),
				),
			},
			{ // Update integration with credential specified
				Config: generateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(credTypeName1),
					generateCredentialFields(
						generateMapProperty(key1, strconv.Quote(val1)),
					),
				) + generateIntegrationResource(
					inteResource2,
					nullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID2),
					generateIntegrationConfig(
						strconv.Quote(inteName2),
						nullValue, // Empty notes
						generateMapProperty(credTypeName1, "genesyscloud_integration_credential."+credResource1+".id"), // Reference credential ID
						generateJsonEncodedProperties(
							generateJsonProperty("smtpHost", strconv.Quote("fakeHost")),
						),
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "intended_state", defaultState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "integration_type", typeID2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "config.0.name", inteName2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "config.0.notes", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource2, "config.0.advanced", emptyJSON),
					resource.TestCheckResourceAttrPair("genesyscloud_integration."+inteResource2, "config.0.credentials."+credTypeName1, "genesyscloud_integration_credential."+credResource1, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration." + inteResource2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIntegrationDestroyed,
	})
}

func generateIntegrationResource(resourceID string, intendedState string, integrationType string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration" "%s" {
        intended_state = %s
        integration_type = %s
        %s
	}
	`, resourceID, intendedState, integrationType, strings.Join(attrs, "\n"))
}

func generateIntegrationConfig(name string, notes string, cred string, props string, adv string) string {
	return fmt.Sprintf(`config {
        name = %s
        notes = %s
        credentials = {
            %s
        }
        properties = %s
        advanced = %s
	}
	`, name, notes, cred, props, adv)
}

func validateIntegrationProperties(integrationResourceName string, groupResourceName string, propDisplayType string, propSandbox string, propURL string, groupID string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		integrationResource, ok := state.RootModule().Resources[integrationResourceName]
		if !ok {
			return fmt.Errorf("Failed to find integration %s in state", integrationResourceName)
		}
		integrationID := integrationResource.Primary.ID

		var expectGroupID string
		if groupResourceName == "" {
			if groupID == "" {
				expectGroupID = ""
			} else {
				expectGroupID = strconv.Quote(groupID)
			}
		} else {
			groupResource, ok := state.RootModule().Resources[groupResourceName]
			if !ok {
				return fmt.Errorf("Failed to find group %s in state", groupResourceName)
			}
			expectGroupID = strconv.Quote(groupResource.Primary.ID)
		}

		properties, ok := integrationResource.Primary.Attributes["config.0.properties"]
		if !ok {
			return fmt.Errorf("No properties found for integration %s in state", integrationID)
		}

		expectProperties := fmt.Sprintf(`{"displayType":"%s","groups":[%s],"sandbox":"%s","url":"%s"}`, propDisplayType, expectGroupID, propSandbox, propURL)

		if properties == expectProperties {
			return nil
		}

		return fmt.Errorf("Found group %s does not match with integration %s", groupID, integrationID)
	}
}

func testVerifyIntegrationDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration" {
			continue
		}

		integration, resp, err := integrationAPI.GetIntegration(rs.Primary.ID, 100, 1, "", nil, "", "")
		if integration != nil {
			return fmt.Errorf("Integration (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
			// Integration not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All integrations destroyed
	return nil
}
