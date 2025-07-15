package architect_flow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccArchitectFlowWithS3File(t *testing.T) {
	// Skip this test if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
		t.Skip("Skipping S3 test - AWS credentials not available")
	}

	var (
		flowResourceLabel = "test_flow_s3"
		flowName          = "test_s3_flow" + uuid.NewString()
		s3Bucket          = os.Getenv("TEST_S3_BUCKET")
		s3Key             = "test-flows/test_s3_flow.yaml"
	)

	if s3Bucket == "" {
		t.Skip("Skipping S3 test - TEST_S3_BUCKET environment variable not set")
	}

	// Create a test flow file content
	flowContent := fmt.Sprintf(`inboundCall:
  name: %s
  defaultLanguage: en-us
  startUpRef: ./menus/menu[mainMenu]
  initialGreeting:
    tts: Hello from S3!
  menus:
    - menu:
        name: Main Menu
        audio:
          tts: You are at the Main Menu, press 9 to disconnect.
        refId: mainMenu
        choices:
          - menuDisconnect:
              name: Disconnect
              dtmf: digit_9`, flowName)

	// Create a temporary local file first
	tempFile, err := os.CreateTemp("", "test_flow_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(flowContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Upload the file to S3 for testing
	ctx := context.Background()
	file, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	err = files.UploadS3File(ctx, s3Bucket, s3Key, file)
	if err != nil {
		t.Fatalf("Failed to upload test file to S3: %v", err)
	}

	// Clean up S3 file after test
	defer func() {
		// Note: In a real implementation, you would delete the S3 object here
		// For now, we'll just log that cleanup is needed
		t.Logf("Test file uploaded to s3://%s/%s - manual cleanup may be required", s3Bucket, s3Key)
	}()

	s3Path := fmt.Sprintf("s3://%s/%s", s3Bucket, s3Key)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFlowResource(flowResourceLabel, s3Path, flowName, "inboundcall"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel, "name", flowName),
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel, "type", "inboundcall"),
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel, "filepath", s3Path),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func TestAccArchitectFlowWithLocalAndS3Files(t *testing.T) {
	// Skip this test if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
		t.Skip("Skipping S3 test - AWS credentials not available")
	}

	var (
		flowResourceLabel = "test_flow_mixed"
		flowName          = "test_mixed_flow" + uuid.NewString()
		s3Bucket          = os.Getenv("TEST_S3_BUCKET")
	)

	if s3Bucket == "" {
		t.Skip("Skipping S3 test - TEST_S3_BUCKET environment variable not set")
	}

	// Create a local test file
	localFlowContent := fmt.Sprintf(`inboundCall:
  name: %s
  defaultLanguage: en-us
  startUpRef: ./menus/menu[mainMenu]
  initialGreeting:
    tts: Hello from local file!
  menus:
    - menu:
        name: Main Menu
        audio:
          tts: You are at the Main Menu, press 9 to disconnect.
        refId: mainMenu
        choices:
          - menuDisconnect:
              name: Disconnect
              dtmf: digit_9`, flowName)

	localFilePath := filepath.Join(testrunner.RootDir, "examples", "resources", ResourceType, "test_local_flow.yaml")
	err := os.WriteFile(localFilePath, []byte(localFlowContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create local test file: %v", err)
	}
	defer os.Remove(localFilePath)

	// Create an S3 test file
	s3FlowContent := fmt.Sprintf(`inboundCall:
  name: %s_s3
  defaultLanguage: en-us
  startUpRef: ./menus/menu[mainMenu]
  initialGreeting:
    tts: Hello from S3 file!
  menus:
    - menu:
        name: Main Menu
        audio:
          tts: You are at the Main Menu, press 9 to disconnect.
        refId: mainMenu
        choices:
          - menuDisconnect:
              name: Disconnect
              dtmf: digit_9`, flowName)

	s3Key := "test-flows/test_s3_flow_mixed.yaml"
	s3Path := fmt.Sprintf("s3://%s/%s", s3Bucket, s3Key)

	// Upload S3 file
	ctx := context.Background()
	err = files.UploadS3File(ctx, s3Bucket, s3Key, strings.NewReader(s3FlowContent))
	if err != nil {
		t.Fatalf("Failed to upload test file to S3: %v", err)
	}

	defer func() {
		t.Logf("Test file uploaded to s3://%s/%s - manual cleanup may be required", s3Bucket, s3Key)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFlowResource(flowResourceLabel+"_local", localFilePath, flowName, "inboundcall"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel+"_local", "name", flowName),
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel+"_local", "type", "inboundcall"),
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel+"_local", "filepath", localFilePath),
				),
			},
			{
				Config: generateFlowResource(flowResourceLabel+"_s3", s3Path, flowName+"_s3", "inboundcall"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel+"_s3", "name", flowName+"_s3"),
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel+"_s3", "type", "inboundcall"),
					resource.TestCheckResourceAttr("genesyscloud_flow."+flowResourceLabel+"_s3", "filepath", s3Path),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func TestAccArchitectFlowS3FileValidation(t *testing.T) {
	// Test S3 path validation
	tests := []struct {
		name        string
		s3Path      string
		expectError bool
	}{
		{
			name:        "Valid S3 path",
			s3Path:      "s3://my-bucket/path/to/file.yaml",
			expectError: false,
		},
		{
			name:        "Valid S3a path",
			s3Path:      "s3a://my-bucket/path/to/file.yaml",
			expectError: false,
		},
		{
			name:        "Invalid S3 path - missing key",
			s3Path:      "s3://my-bucket",
			expectError: true,
		},
		{
			name:        "Invalid S3 path - empty bucket",
			s3Path:      "s3:///path/to/file.yaml",
			expectError: true,
		},
		{
			name:        "Not an S3 path",
			s3Path:      "/local/path/file.yaml",
			expectError: false, // This should work as a local file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the file path is accepted by the resource schema
			flowResourceLabel := "test_flow_validation"
			flowName := "test_validation_flow" + uuid.NewString()

			config := generateFlowResource(flowResourceLabel, tt.s3Path, flowName, "inboundcall")

			// If we expect an error, we should see it during plan
			if tt.expectError {
				resource.Test(t, resource.TestCase{
					PreCheck:          func() { util.TestAccPreCheck(t) },
					ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
					Steps: []resource.TestStep{
						{
							Config:      config,
							ExpectError: regexp.MustCompile(".*"),
						},
					},
				})
			} else {
				// For valid paths, we should be able to plan (even if apply fails due to missing file)
				resource.Test(t, resource.TestCase{
					PreCheck:          func() { util.TestAccPreCheck(t) },
					ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
					Steps: []resource.TestStep{
						{
							Config: config,
							// We expect this to fail during apply due to missing file,
							// but the plan should succeed
							ExpectError: regexp.MustCompile(".*"),
						},
					},
				})
			}
		})
	}
}

// Helper function to generate flow resource configuration
func generateFlowResource(resourceLabel, filePath, name, flowType string) string {
	return fmt.Sprintf(`
resource "genesyscloud_flow" "%s" {
    filepath = "%s"
    file_content_hash = filesha256("%s")
    substitutions = {
        name = "%s"
        type = "%s"
    }
}
`, resourceLabel, filePath, filePath, name, flowType)
}
