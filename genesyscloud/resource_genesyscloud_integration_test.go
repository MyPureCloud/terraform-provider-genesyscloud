package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func TestAccResourceIntegration(t *testing.T) {
	var (
		inteResource1 = "test_integration1"
		inteName1     = "Terraform Integration Test-" + uuid.NewString()
		inteName2     = "Terraform Integration Test-" + uuid.NewString()

		defaultState = "DISABLED"
		enabledState = "ENABLED"
		configNotes  = "some notes"
		configNotes2 = "This is a note"

		typeID          = "embedded-client-app"
		typeName        = "Client Application"
		typeDescription = "Embeds third-party webapps via iframe in the Genesys Cloud UI."
		typeProvider    = "clientapps"
		typeCategory    = "Client Apps"

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
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIntegrationResource(
					inteResource1,
					inteName1,
					nullValue, //Empty intended_state, default value is "DISABLED"
					generateIntegrationType(
						strconv.Quote(typeID),
						strconv.Quote(typeName),
						strconv.Quote(typeDescription),
						strconv.Quote(typeProvider),
						strconv.Quote(typeCategory),
					),
					generateIntegrationConfig(
						inteName1,
						nullValue, //Empty notes
						nullValue, //Empty credential ID
						nullValue, //Empty properties
						nullValue, //Empty advanced JSON
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState), // Default value would be "DISABLED"
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.id", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.name", typeName),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.description", typeDescription),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.provider", typeProvider),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.category", typeCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", ""),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.properties", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials", emptyJSON),
				),
			},
			{
				// Update intendedState, name, notes, properties
				Config: generateIntegrationResource(
					inteResource1,
					inteName2,
					strconv.Quote(enabledState),
					generateIntegrationType(
						strconv.Quote(typeID),
						strconv.Quote(typeName),
						strconv.Quote(typeDescription),
						strconv.Quote(typeProvider),
						strconv.Quote(typeCategory),
					),
					generateIntegrationConfig(
						inteName2,
						strconv.Quote(configNotes),
						nullValue, //Empty credential ID
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
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "name", inteName2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", enabledState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.id", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.name", typeName),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.description", typeDescription),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.provider", typeProvider),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.category", typeCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials", emptyJSON),
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
				Config: generateGroupResource(
					groupResource1,
					groupName,
					nullValue, // No description
					nullValue, // Default type
					nullValue, // Default visibility
					nullValue, // Default rules_visible
				) + generateIntegrationResource(
					inteResource1,
					inteName1,
					strconv.Quote(enabledState),
					generateIntegrationType(
						strconv.Quote(typeID),
						strconv.Quote(typeName),
						strconv.Quote(typeDescription),
						strconv.Quote(typeProvider),
						strconv.Quote(typeCategory),
					),
					generateIntegrationConfig(
						inteName1,
						strconv.Quote(configNotes),
						nullValue, //Empty credential ID
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
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", enabledState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.id", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.name", typeName),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.description", typeDescription),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.provider", typeProvider),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.category", typeCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials", emptyJSON),
					validateIntegrationProperties("genesyscloud_integration."+inteResource1, "genesyscloud_group."+groupResource1, propDisplayType, propSandbox, propURL, ""),
				),
			},
			{ // Remove the group reference and update intendedState and notes
				Config: generateIntegrationResource(
					inteResource1,
					inteName1,
					nullValue, //Change to default value
					generateIntegrationType(
						strconv.Quote(typeID),
						strconv.Quote(typeName),
						strconv.Quote(typeDescription),
						strconv.Quote(typeProvider),
						strconv.Quote(typeCategory),
					),
					generateIntegrationConfig(
						inteName1,
						strconv.Quote(configNotes2),
						nullValue, //Empty credentials
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
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "name", inteName1),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "intended_state", defaultState),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.id", typeID),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.name", typeName),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.description", typeDescription),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.provider", typeProvider),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "integration_type.0.category", typeCategory),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.notes", configNotes2),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.advanced", emptyJSON),
					resource.TestCheckResourceAttr("genesyscloud_integration."+inteResource1, "config.0.credentials", emptyJSON),
					validateIntegrationProperties("genesyscloud_integration."+inteResource1, "", propDisplayType, propSandbox, propURL, ""),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_integration." + inteResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			//TODO: After credential resource is created, create a test credential and add it into properties
		},
		CheckDestroy: testVerifyDivisionsDestroyed,
	})
}

func generateIntegrationResource(resourceID string, name string, intendedState string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration" "%s" {
        name = "%s"
        intended_state = %s
        %s
	}
	`, resourceID, name, intendedState, strings.Join(attrs, "\n"))
}

func generateIntegrationType(id string, name string, description string, provider string, category string) string {
	return fmt.Sprintf(`integration_type {
		id = %s
        name = %s
        description = %s
        provider = %s
        category = %s
	}
	`, id, name, description, provider, category)
}

func generateIntegrationConfig(name string, notes string, cred string, props string, adv string) string {
	return fmt.Sprintf(`config {
        name = "%s"
        notes = %s
        credentials = %s
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
		} else if resp != nil && resp.StatusCode == 404 {
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
