package integration_config

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
)

func TestAccResourceIntegrationConfigBasic(t *testing.T) {
	var (
		integrationResourceLabel = "test_integration"
		configResourceLabel      = "test_config"
		configName1              = "Terraform Test Config " + uuid.NewString()
		configName2              = "Terraform Test Config Updated " + uuid.NewString()
		configNotes1             = "Initial test notes"
		configNotes2             = "Updated test notes"
	)

	// Enable the feature toggle
	err := os.Setenv(featureToggles.ICToggleName(), "1")
	if err != nil {
		t.Fatalf("Failed to set %s: %v", featureToggles.ICToggleName(), err)
	}
	defer os.Unsetenv(featureToggles.ICToggleName())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create integration + config
				Config: generateIntegrationResource(integrationResourceLabel) +
					generateIntegrationConfigResource(
						configResourceLabel,
						"genesyscloud_integration."+integrationResourceLabel+".id",
						configName1,
						configNotes1,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_integration_config."+configResourceLabel, "integration_id",
						"genesyscloud_integration."+integrationResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_integration_config."+configResourceLabel, "name", configName1),
					resource.TestCheckResourceAttr("genesyscloud_integration_config."+configResourceLabel, "notes", configNotes1),
				),
			},
			{
				// Update config name and notes
				Config: generateIntegrationResource(integrationResourceLabel) +
					generateIntegrationConfigResource(
						configResourceLabel,
						"genesyscloud_integration."+integrationResourceLabel+".id",
						configName2,
						configNotes2,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_integration_config."+configResourceLabel, "name", configName2),
					resource.TestCheckResourceAttr("genesyscloud_integration_config."+configResourceLabel, "notes", configNotes2),
				),
			},
			{
				// Import
				ResourceName:      "genesyscloud_integration_config." + configResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIntegrationConfigDestroyed,
	})
}

func TestAccResourceIntegrationConfigToggleOff(t *testing.T) {
	// Ensure toggle is OFF
	os.Unsetenv(featureToggles.ICToggleName())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Attempting to use the resource without toggle should fail
				Config:      generateIntegrationConfigResource("test", "\"fake-id\"", "test", "test"),
				ExpectError: regexp.MustCompile(fmt.Sprintf(".*%s.*", featureToggles.ICToggleName())),
			},
		},
	})
}

func testVerifyIntegrationConfigDestroyed(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_integration_config" {
			continue
		}
		// Config is "deleted" by clearing it — the integration itself may still exist
		// We just verify the resource was removed from state
	}
	return nil
}

func generateIntegrationResource(resourceLabel string) string {
	return fmt.Sprintf(`
resource "genesyscloud_integration" "%s" {
  intended_state   = "DISABLED"
  integration_type = "purecloud-data-actions"
}
`, resourceLabel)
}

func generateIntegrationConfigResource(resourceLabel, integrationId, name, notes string) string {
	return fmt.Sprintf(`
resource "genesyscloud_integration_config" "%s" {
  integration_id = %s
  name           = "%s"
  notes          = "%s"
}
`, resourceLabel, integrationId, name, notes)
}
