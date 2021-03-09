package genesyscloud

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const exportTestDir = "../.terraform"

func TestAccResourceTfExport(t *testing.T) {
	var (
		exportResource1 = "test-export1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Run export without state file
				Config: generateTfExportResource(
					exportResource1,
					exportTestDir,
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(filepath.Join(exportTestDir, defaultTfJSONFile)),
				),
			},
			{
				// Run export with state file
				Config: generateTfExportResource(
					exportResource1,
					exportTestDir,
					trueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(filepath.Join(exportTestDir, defaultTfJSONFile)),
					validateFileCreated(filepath.Join(exportTestDir, defaultTfStateFile)),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyed,
	})
}

func generateTfExportResource(
	resourceID string,
	directory string,
	includeState string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		resource_types = ["genesyscloud_routing_queue", "genesyscloud_routing_skill"]
        include_state_file = %s
    }
	`, resourceID, directory, includeState)
}

func validateFileCreated(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("Failed to find file %s", filename)
		}
		return nil
	}
}

func testVerifyExportsDestroyed(state *terraform.State) error {
	// Check config file deleted
	configPath := filepath.Join(exportTestDir, defaultTfJSONFile)
	_, err := os.Stat(configPath)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to delete config file %s", configPath)
	}

	// Check state file deleted
	statePath := filepath.Join(exportTestDir, defaultTfStateFile)
	_, err = os.Stat(statePath)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to delete state file %s", configPath)
	}
	return nil
}
