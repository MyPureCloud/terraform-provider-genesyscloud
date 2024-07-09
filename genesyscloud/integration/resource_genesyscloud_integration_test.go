package integration

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	mu sync.Mutex
)

/*
The resource_genesyscloud_integration_test.go contains all of the test cases for running the resource
tests for integrations.
*/
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

		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@example.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create without config
				Config: GenerateIntegrationResource(
					inteResource1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
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
				Config: GenerateIntegrationResource(
					inteResource1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						util.NullValue, //Empty notes
						"",             //Empty credential ID
						util.NullValue, //Empty properties
						util.NullValue, //Empty advanced JSON
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
				Config: GenerateIntegrationResource(
					inteResource1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						util.NullValue, // No name update. Should stay the same
						util.NullValue, //Empty notes
						"",             //Empty credential ID
						util.NullValue, //Empty properties
						util.NullValue, //Empty advanced JSON
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
				Config: GenerateIntegrationResource(
					inteResource1,
					strconv.Quote(enabledState),
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName2),
						strconv.Quote(configNotes),
						"", //Empty credential ID
						util.GenerateJsonEncodedProperties(
							util.GenerateJsonProperty(displayTypeKey, strconv.Quote(propDisplayType)),
							util.GenerateJsonProperty(urlKey, strconv.Quote(propURL)),
							util.GenerateJsonProperty(sandboxKey, strconv.Quote(propSandbox)),
							util.GenerateJsonProperty(groupsKey, fmt.Sprintf(`[%s]`, strconv.Quote(fakeGroupID))),
						),
						util.NullValue,
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
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + generateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + GenerateIntegrationResource(
					inteResource1,
					strconv.Quote(enabledState),
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						strconv.Quote(configNotes),
						"", //Empty credential ID
						util.GenerateJsonEncodedProperties(
							util.GenerateJsonProperty(displayTypeKey, strconv.Quote(propDisplayType)),
							util.GenerateJsonProperty(urlKey, strconv.Quote(propURL)),
							util.GenerateJsonProperty(sandboxKey, strconv.Quote(propSandbox)),
							util.GenerateJsonProperty(groupsKey, fmt.Sprintf(`[%s]`, "genesyscloud_group."+groupResource1+".id")),
						),
						util.NullValue,
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
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
			},
			{ // Remove the group reference and update intendedState and notes
				Config: GenerateIntegrationResource(
					inteResource1,
					util.NullValue, //Change to default value
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						strconv.Quote(configNotes2),
						"", //Empty credentials
						util.GenerateJsonEncodedProperties(
							util.GenerateJsonProperty(displayTypeKey, strconv.Quote(propDisplayType)),
							util.GenerateJsonProperty(urlKey, strconv.Quote(propURL)),
							util.GenerateJsonProperty(sandboxKey, strconv.Quote(propSandbox)),
							util.GenerateJsonProperty(groupsKey, "[]"),
						),
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds
						return nil
					},
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
				Config: GenerateIntegrationResource(
					inteResource1,
					util.NullValue, //Default value
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName2),
						strconv.Quote(configNotes2),
						"", //Empty credentials
						// Use Raw JSON instead of jsonencode function
						fmt.Sprintf(`"{  \"%s\":   \"%s\",  \"%s\": \"%s\",  \"%s\": \"%s\",  \"%s\": %s}"`, displayTypeKey, propDisplayType, urlKey, propURL, sandboxKey, propSandbox, groupsKey, "[]"),
						util.NullValue,
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
				Config: integrationCred.GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(credTypeName1),
					integrationCred.GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1),
						},
					),
				) + GenerateIntegrationResource(
					inteResource2,
					strconv.Quote(enabledState),
					strconv.Quote(typeID2),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						strconv.Quote(configNotes),
						util.GenerateMapProperty(credTypeName1, "genesyscloud_integration_credential."+credResource1+".id"), // Reference credential ID
						util.GenerateJsonEncodedProperties(
							util.GenerateJsonProperty("smtpHost", strconv.Quote("fakeHost")),
						),
						util.NullValue,
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
				Config: integrationCred.GenerateCredentialResource(
					credResource1,
					strconv.Quote(credName1),
					strconv.Quote(credTypeName1),
					integrationCred.GenerateCredentialFields(
						map[string]string{
							key1: strconv.Quote(val1),
						},
					),
				) + GenerateIntegrationResource(
					inteResource2,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID2),
					GenerateIntegrationConfig(
						strconv.Quote(inteName2),
						util.NullValue, // Empty notes
						util.GenerateMapProperty(credTypeName1, "genesyscloud_integration_credential."+credResource1+".id"), // Reference credential ID
						util.GenerateJsonEncodedProperties(
							util.GenerateJsonProperty("smtpHost", strconv.Quote("fakeHost")),
						),
						util.NullValue,
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
		CheckDestroy: testVerifyIntegrationAndUsersDestroyed,
	})
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

func testVerifyIntegrationAndUsersDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_integration" {
			integration, resp, err := integrationAPI.GetIntegration(rs.Primary.ID, 100, 1, "", nil, "", "")
			if integration != nil {
				return fmt.Errorf("Integration (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// Integration not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
		if rs.Type == "genesyscloud_user" {
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if user != nil {
				return fmt.Errorf("User Resource (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// User not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
	}
	// Success. All integrations destroyed
	return nil
}

// TODO: Duplicating this code within the function to not break a cyclic dependency
func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}

// TODO: Duplicating this code within the function to not break a cyclic dependency
func generateBasicGroupResource(resourceID string, name string, nestedBlocks ...string) string {
	return generateGroupResource(resourceID, name, util.NullValue, util.NullValue, util.NullValue, util.TrueValue, nestedBlocks...)
}

func generateGroupResource(
	resourceID string,
	name string,
	desc string,
	groupType string,
	visibility string,
	rulesVisible string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_group" "%s" {
		name = "%s"
		description = %s
		type = %s
		visibility = %s
		rules_visible = %s
        %s
	}
	`, resourceID, name, desc, groupType, visibility, rulesVisible, strings.Join(nestedBlocks, "\n"))
}

func generateGroupOwners(userIDs ...string) string {
	return fmt.Sprintf(`owner_ids = [%s]
	`, strings.Join(userIDs, ","))
}

func checkUserDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching user with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 30
		for i := 0; i < maxAttempts; i++ {

			deleted, err := isUserDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("user %s was not deleted properly", id)
	}
}

func isUserDeleted(id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()

	usersAPI := platformclientv2.NewUsersApi()
	// Attempt to get the user
	_, response, err := usersAPI.GetUser(id, nil, "", "")

	// Check if the user is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // User is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If user is found, it means the user is not deleted
	return false, nil
}
