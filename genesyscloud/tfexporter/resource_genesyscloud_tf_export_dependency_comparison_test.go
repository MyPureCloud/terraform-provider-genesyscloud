package tfexporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	testrunner "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

// TestAccResourceTfExportDependencyResolutionComparison tests that dependency resolution
// produces the same results when using include_filter_resources (regex) vs include_filter_resources_by_id (ID)
// This validates that the change to enable dependency resolution for include_filter_resources_by_id works correctly
func TestAccResourceTfExportDependencyResolutionComparison(t *testing.T) {
	testSetup(t)
	var (
		// Test configuration
		flowName = "Email Decryption Flow"
		flowId   = "b84cbae3-7c54-45dc-ade0-7a30fbccf996"
		flowType = "inboundemail"

		// Export directories
		exportTestDirRegex = testrunner.GetTestTempPath(".terraform-regex-" + uuid.NewString())
		exportTestDirById  = testrunner.GetTestTempPath(".terraform-byid-" + uuid.NewString())

		// Resource labels
		exportResourceLabelRegex = "test-export-regex"
		exportResourceLabelById  = "test-export-byid"
	)

	// Clean up test directories after test completes
	defer os.RemoveAll(exportTestDirRegex)
	defer os.RemoveAll(exportTestDirById)

	// Configuration for regex-based export
	configRegex := fmt.Sprintf(`
resource "genesyscloud_tf_export" "%s" {
	directory = "%s"
	include_state_file = true
	include_filter_resources = [%s]
	export_format = "json"
	split_files_by_resource = true
	enable_dependency_resolution = true
}
`, exportResourceLabelRegex, exportTestDirRegex, strconv.Quote("genesyscloud_flow::"+flowName))

	// Configuration for ID-based export
	configById := fmt.Sprintf(`
resource "genesyscloud_tf_export" "%s" {
	directory = "%s"
	include_state_file = true
	include_filter_resources_by_id = [%s]
	export_format = "json"
	split_files_by_resource = true
	enable_dependency_resolution = true
}
`, exportResourceLabelById, exportTestDirById, strconv.Quote("genesyscloud_flow::"+flowId))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Step 1: Export using regex filter
			{
				Config: configRegex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResourceLabelRegex, "directory", exportTestDirRegex),
					validateExportDirectoryExists(exportTestDirRegex),
				),
			},
			// Step 2: Export using ID filter
			{
				Config: configById,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResourceLabelById, "directory", exportTestDirById),
					validateExportDirectoryExists(exportTestDirById),
					// Compare the exports after both have completed
					compareExportResults(t, exportTestDirRegex, exportTestDirById, flowName, flowId, flowType),
				),
			},
		},
		CheckDestroy: resource.ComposeTestCheckFunc(
			testVerifyExportsDestroyedFunc(exportTestDirRegex),
			testVerifyExportsDestroyedFunc(exportTestDirById),
		),
	})
}

// validateExportDirectoryExists checks that the export directory was created
func validateExportDirectoryExists(directory string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			return fmt.Errorf("export directory %s does not exist", directory)
		}
		return nil
	}
}

// compareExportResults compares two export directories to ensure they produced the same results
func compareExportResults(t *testing.T, regexDir, byIdDir, flowName, flowId, flowType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Logf("Comparing export results between regex (%s) and by-id (%s) exports", regexDir, byIdDir)

		// Read the JSON config files from both exports
		regexFiles, err := getExportedFiles(regexDir)
		if err != nil {
			return fmt.Errorf("failed to read regex export directory: %w", err)
		}

		byIdFiles, err := getExportedFiles(byIdDir)
		if err != nil {
			return fmt.Errorf("failed to read by-id export directory: %w", err)
		}

		t.Logf("Regex export contains %d files", len(regexFiles))
		t.Logf("By-ID export contains %d files", len(byIdFiles))

		// Compare the number of files
		if len(regexFiles) != len(byIdFiles) {
			return fmt.Errorf("export file count mismatch: regex=%d, by-id=%d\nRegex files: %v\nBy-ID files: %v",
				len(regexFiles), len(byIdFiles), regexFiles, byIdFiles)
		}

		// Compare each file exists in both exports
		for _, file := range regexFiles {
			if !contains(byIdFiles, file) {
				return fmt.Errorf("file %s exists in regex export but not in by-id export", file)
			}
		}

		// Read and compare the state files
		regexStateFile := filepath.Join(regexDir, defaultTfStateFile)
		byIdStateFile := filepath.Join(byIdDir, defaultTfStateFile)

		regexState, err := readStateFile(regexStateFile)
		if err != nil {
			return fmt.Errorf("failed to read regex state file: %w", err)
		}

		byIdState, err := readStateFile(byIdStateFile)
		if err != nil {
			return fmt.Errorf("failed to read by-id state file: %w", err)
		}

		// Compare resource counts
		regexResourceCount := len(regexState.Resources)
		byIdResourceCount := len(byIdState.Resources)

		t.Logf("Regex export has %d resources", regexResourceCount)
		t.Logf("By-ID export has %d resources", byIdResourceCount)

		if regexResourceCount != byIdResourceCount {
			return fmt.Errorf("resource count mismatch: regex=%d, by-id=%d", regexResourceCount, byIdResourceCount)
		}

		// Extract resource types and IDs from both exports
		regexResources := extractResourceInfo(regexState)
		byIdResources := extractResourceInfo(byIdState)

		// Compare resource types
		if len(regexResources) != len(byIdResources) {
			return fmt.Errorf("unique resource count mismatch: regex=%d, by-id=%d", len(regexResources), len(byIdResources))
		}

		// Log the exported resources for debugging
		t.Logf("Exported resource types and counts:")
		for resType, ids := range regexResources {
			t.Logf("  %s: %d resources", resType, len(ids))
		}

		// Verify each resource type and ID exists in both exports
		for resType, regexIds := range regexResources {
			byIdIds, exists := byIdResources[resType]
			if !exists {
				return fmt.Errorf("resource type %s exists in regex export but not in by-id export", resType)
			}

			if len(regexIds) != len(byIdIds) {
				return fmt.Errorf("resource count mismatch for type %s: regex=%d, by-id=%d",
					resType, len(regexIds), len(byIdIds))
			}

			// Check that all IDs match
			for id := range regexIds {
				if _, exists := byIdIds[id]; !exists {
					return fmt.Errorf("resource %s with ID %s exists in regex export but not in by-id export",
						resType, id)
				}
			}
		}

		// Verify the main flow is included
		flowResources := regexResources["genesyscloud_flow"]
		if len(flowResources) == 0 {
			return fmt.Errorf("no genesyscloud_flow resources found in exports")
		}

		if _, exists := flowResources[flowId]; !exists {
			return fmt.Errorf("expected flow with ID %s not found in exports", flowId)
		}

		t.Logf("✓ Export comparison successful: both methods produced identical results with %d total resources", regexResourceCount)
		return nil
	}
}

// getExportedFiles returns a list of files in the export directory
func getExportedFiles(directory string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TerraformState represents a simplified terraform state file structure
type TerraformState struct {
	Version   int                 `json:"version"`
	Resources []TerraformResource `json:"resources"`
}

// TerraformResource represents a resource in the terraform state
type TerraformResource struct {
	Type      string                      `json:"type"`
	Name      string                      `json:"name"`
	Instances []TerraformResourceInstance `json:"instances"`
}

// TerraformResourceInstance represents an instance of a resource
type TerraformResourceInstance struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// readStateFile reads and parses a terraform state file
func readStateFile(filePath string) (*TerraformState, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var state TerraformState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// extractResourceInfo extracts resource types and IDs from a state file
func extractResourceInfo(state *TerraformState) map[string]map[string]bool {
	resources := make(map[string]map[string]bool)

	for _, res := range state.Resources {
		if resources[res.Type] == nil {
			resources[res.Type] = make(map[string]bool)
		}

		for _, instance := range res.Instances {
			if id, ok := instance.Attributes["id"].(string); ok {
				resources[res.Type][id] = true
			}
		}
	}

	return resources
}
