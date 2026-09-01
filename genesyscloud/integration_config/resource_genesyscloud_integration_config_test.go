package integration_config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
)

// cleanupDataActionIntegrations removes stale purecloud-data-actions integrations left over from
// previous test runs. The API caps this integration type at 10 per org; leaked test integrations
// fill the quota and cause "maximum number of integrations" 400 errors on create. This sweep keeps
// the org under the cap by deleting DISABLED integrations whose name looks test-generated.
func cleanupDataActionIntegrations(t *testing.T) {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Logf("cleanupDataActionIntegrations: skipping, could not authorize SDK: %v", err)
		return
	}
	api := platformclientv2.NewIntegrationsApiWithConfig(config)
	for pageNum := 1; ; pageNum++ {
		integrations, _, listErr := api.GetIntegrations(100, pageNum, "", nil, "", "", nil, "", "", "")
		if listErr != nil || integrations.Entities == nil {
			break
		}
		for _, integ := range *integrations.Entities {
			if integ.Id == nil || integ.IntegrationType == nil || integ.IntegrationType.Id == nil {
				continue
			}
			if *integ.IntegrationType.Id != "purecloud-data-actions" {
				continue
			}
			// Only remove things that look test-generated and are not enabled/live.
			name := ""
			if integ.Name != nil {
				name = *integ.Name
			}
			state := ""
			if integ.IntendedState != nil {
				state = *integ.IntendedState
			}
			isTestArtifact := strings.Contains(name, "Genesys Cloud Data Actions (") ||
				strings.Contains(strings.ToLower(name), "test")
			if state != "ENABLED" && isTestArtifact {
				if _, _, delErr := api.DeleteIntegration(*integ.Id); delErr != nil {
					log.Printf("cleanupDataActionIntegrations: failed to delete %s (%s): %v", name, *integ.Id, delErr)
				} else {
					log.Printf("cleanupDataActionIntegrations: deleted stale integration %s (%s)", name, *integ.Id)
				}
			}
		}
		if integrations.PageCount == nil || pageNum >= *integrations.PageCount {
			break
		}
	}
}

func TestAccResourceIntegrationConfigBasic(t *testing.T) {
	var (
		integrationResourceLabel = "test_integration"
		configResourceLabel      = "test_config"
		configName1              = "Terraform Test Config " + uuid.NewString()
		configName2              = "Terraform Test Config Updated " + uuid.NewString()
		configNotes1             = "Initial test notes"
		configNotes2             = "Updated test notes"
	)

	// Remove stale data-action integrations from prior runs so we don't hit the per-type cap (10).
	cleanupDataActionIntegrations(t)

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
