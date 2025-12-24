package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack"
	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

func TestAccResourceResponseManagementResponseAsset(t *testing.T) {
	var (
		resourceLabel         = "responseasset"
		resourceFullPath      = ResourceType + "." + resourceLabel
		testFilesDir          = "test_responseasset_data"
		fileName1             = "yeti-img.png"
		fileName2             = "genesys-img.png"
		fullPath1             = filepath.Join(testFilesDir, fileName1)
		fullPath2             = filepath.Join(testFilesDir, fileName2)
		divisionResourceLabel = "test_div"
		divisionName          = "test tf divison " + uuid.NewString()
	)

	cleanupResponseAssets("genesys")
	cleanupResponseAssets("yeti")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fullPath1, util.NullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFullPath, "filename", fullPath1),
					resource.TestCheckResourceAttr(resourceFullPath, "name", fullPath1),
					provider.TestDefaultHomeDivision(resourceFullPath),
				),
			},
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fullPath2, "genesyscloud_auth_division."+divisionResourceLabel+".id") +
					authDivision.GenerateAuthDivisionBasic(divisionResourceLabel, divisionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFullPath, "filename", fullPath2),
					resource.TestCheckResourceAttr(resourceFullPath, "name", fullPath2),
					resource.TestCheckResourceAttrPair(resourceFullPath, "division_id",
						authDivision.ResourceType+"."+divisionResourceLabel, "id"),
				),
			},
			// Update
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fullPath2, "data.genesyscloud_auth_division_home.home.id") +
					"\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFullPath, "filename", fullPath2),
					resource.TestCheckResourceAttr(resourceFullPath, "name", fullPath2),
					provider.TestDefaultHomeDivision(resourceFullPath),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceFullPath,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"filename",
					"file_content_hash",
				},
			},
		},
		CheckDestroy: testVerifyResponseAssetDestroyed,
	})
}

func TestAccResourceResponseManagementResponseAssetWithNameField(t *testing.T) {
	var (
		resourceLabel         = "responseasset"
		resourceFullPath      = ResourceType + "." + resourceLabel
		testFilesDir          = "test_responseasset_data"
		fileName1             = "yeti-img.png"
		fileName2             = "genesys-img.png"
		fullPath1             = filepath.Join(testFilesDir, fileName1)
		fullPath2             = filepath.Join(testFilesDir, fileName2)
		divisionResourceLabel = "test_div"
		divisionName          = "test tf divison " + uuid.NewString()
	)

	cleanupResponseAssets("genesys")
	cleanupResponseAssets("yeti")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateResponseManagementResponseAssetResourceWithNameField(resourceLabel, fileName1, fullPath1, util.NullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFullPath, "filename", fullPath1),
					resource.TestCheckResourceAttr(resourceFullPath, "name", fileName1),
					provider.TestDefaultHomeDivision(resourceFullPath),
				),
			},
			{
				Config: GenerateResponseManagementResponseAssetResourceWithNameField(resourceLabel, fileName2, fullPath2, "genesyscloud_auth_division."+divisionResourceLabel+".id") +
					authDivision.GenerateAuthDivisionBasic(divisionResourceLabel, divisionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFullPath, "filename", fullPath2),
					resource.TestCheckResourceAttr(resourceFullPath, "name", fileName2),
					resource.TestCheckResourceAttrPair(resourceFullPath, "division_id",
						authDivision.ResourceType+"."+divisionResourceLabel, "id"),
				),
			},
			// Update
			{
				Config: GenerateResponseManagementResponseAssetResourceWithNameField(resourceLabel, fileName2, fullPath2, "data.genesyscloud_auth_division_home.home.id") +
					"\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFullPath, "filename", fullPath2),
					resource.TestCheckResourceAttr(resourceFullPath, "name", fileName2),
					provider.TestDefaultHomeDivision(resourceFullPath),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceFullPath,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"filename",
					"file_content_hash",
				},
			},
		},
		CheckDestroy: testVerifyResponseAssetDestroyed,
	})
}

func cleanupResponseAssets(folderName string) error {
	var (
		name    = "name"
		fields  = []string{name}
		varType = "STARTS_WITH"
	)
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return err
	}
	respManagementApi := platformclientv2.NewResponseManagementApiWithConfig(config)

	var filter = platformclientv2.Responseassetfilter{
		Fields:  &fields,
		Value:   &folderName,
		VarType: &varType,
	}

	var body = platformclientv2.Responseassetsearchrequest{
		Query:  &[]platformclientv2.Responseassetfilter{filter},
		SortBy: &name,
	}

	responseData, _, err := respManagementApi.PostResponsemanagementResponseassetsSearch(body, nil)
	if err != nil {
		return err
	}

	if responseData.Results != nil && len(*responseData.Results) > 0 {
		for _, result := range *responseData.Results {
			_, err = respManagementApi.DeleteResponsemanagementResponseasset(*result.Id)
			if err != nil {
				log.Printf("Failed to delete response assets %s: %v", *result.Id, err)
			}
		}
	}
	return nil
}

func testVerifyResponseAssetDestroyed(state *terraform.State) error {
	responseManagementAPI := platformclientv2.NewResponseManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_responsemanagement_responseasset" {
			continue
		}
		responseAsset, resp, err := responseManagementAPI.GetResponsemanagementResponseasset(rs.Primary.ID)
		if responseAsset != nil {
			return fmt.Errorf("response asset (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// response asset not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All response assets destroyed
	return nil
}

// TestAccResourceResponseManagementResponseAssetWithS3 tests the responsemanagement responseasset resource using LocalStack for S3 operations.
// This test validates that the terraform-provider-genesyscloud can successfully create response assets
// with S3 integration through file content hashing using LocalStack as a local AWS service emulator.
//
// Prerequisites:
//   - LocalStack must be running (either locally or in CI)
//   - Environment variables must be set:
//   - USE_LOCAL_STACK=true
//   - LOCAL_STACK_IMAGE_URI=<localstack-image-uri>
//
// Test Flow:
//
//  1. Creates a temporary test image file with initial content
//
//  2. Sets up LocalStack S3 bucket and uploads the test file
//
//  3. Creates a response asset using terraform with S3 source
//
//  4. Updates the file content and re-uploads to S3
//
//  5. Verifies the response asset is recreated with a new GUID due to file_content_hash change
//
//  6. Tests S3 integration through file content hashing mechanism
//
//  7. Cleans up S3 bucket and temporary files
//
// This test is designed to run in CI environments where LocalStack is available
// and properly configured with the required environment variables.
func TestAccResourceResponseManagementResponseAssetWithS3(t *testing.T) {
	/*
		// To run this test locally, uncomment the environment variables below and run `localstack start` from another terminal
		// See more about localstack cli here: https://docs.localstack.cloud/aws/getting-started/installation/
		os.Setenv(localStackEnv.UseLocalStackEnvVar, "true")
		os.Setenv(localStackEnv.LocalStackImageUriEnvVar, "localstack/localstack:latest")
	*/

	imageURI := os.Getenv(localStackEnv.LocalStackImageUriEnvVar)
	if imageURI == "" || !localStackEnv.LocalStackIsActive() {
		t.Skipf("Missing env variables (%s or %s), indicating that localstack is not running", localStackEnv.LocalStackImageUriEnvVar, localStackEnv.UseLocalStackEnvVar)
	}

	ctx := context.Background()
	localStackManager, err := localstack.NewLocalStackManager(ctx)
	if err != nil {
		t.Fatalf("Failed to initialise LocalStackManager: %s", err.Error())
	}

	var (
		resourceLabel = "responseasset-s3"
		bucketName    = "testbucket-" + strings.ToLower(strings.ReplaceAll(uuid.NewString(), "-", ""))
		objectKey     = "test-image.png"
		s3URI         = fmt.Sprintf("s3://%s/%s", bucketName, objectKey)

		fullResourcePath = ResourceType + "." + resourceLabel
	)

	defer func() {
		// cleanup the file objectKey if it exists in the tmp directory
		fullPath := filepath.Join(os.TempDir(), objectKey)
		if _, err := os.Stat(fullPath); err == nil {
			if err := os.Remove(fullPath); err != nil {
				t.Logf("[WARN] Failed to remove file %s: %v", fullPath, err)
			}
		}
	}()

	// Create a temporary test image file with initial content
	// We'll use a simple PNG file content for testing
	initialImageContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC, 0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82}

	// Create a temporary file for the test image
	tmpFile, err := os.CreateTemp("", "test-image-*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Write the initial image content to the temp file
	if err := os.WriteFile(tmpFile.Name(), initialImageContent, 0644); err != nil {
		t.Fatalf("Failed to write initial image content: %v", err)
	}

	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Logf("[WARN] Failed to close temp file: %v", err)
		}

		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("[WARN] Failed to remove temp file: %v", err)
		}
	}()

	// Setup S3 bucket and upload test file
	t.Log("Setting up S3 bucket and uploading test file...")
	err = localStackManager.SetupS3Bucket(bucketName, tmpFile.Name(), objectKey)
	if err != nil {
		t.Fatalf("Failed to setup S3 bucket: %v", err)
	}

	// Cleanup S3 bucket after test
	defer func() {
		if err := localStackManager.CleanupS3Bucket(bucketName); err != nil {
			t.Logf("[WARN] Failed to cleanup S3 bucket: %v", err)
		}
	}()

	// Store the initial resource ID to verify it changes
	var initialResourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, s3URI, util.NullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourcePath, "filename", s3URI),
					provider.TestDefaultHomeDivision(fullResourcePath),
					// Store the initial resource ID
					func(s *terraform.State) error {
						rs := s.RootModule().Resources[fullResourcePath]
						initialResourceID = rs.Primary.ID
						t.Logf("Initial response asset created with ID: %s", initialResourceID)
						return nil
					},
				),
			},
			{
				PreConfig: func() {
					// Update image content to trigger file_content_hash change
					t.Log("Updating image content to trigger file_content_hash change")

					// Create new image content (different from initial)
					updatedImageContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC, 0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82, 0x00, 0x00, 0x00, 0x01}

					// Seek to beginning and truncate to overwrite the file
					if _, err := tmpFile.Seek(0, 0); err != nil {
						t.Fatalf("Failed to seek to beginning of file: %v", err)
					}
					if err := tmpFile.Truncate(0); err != nil {
						t.Fatalf("Failed to truncate file: %v", err)
					}
					_, err := tmpFile.Write(updatedImageContent)
					if err != nil {
						t.Fatalf("Failed to write updated image content: %v", err)
					}

					// Re-upload updated file to S3
					t.Log("Re-uploading updated image file to S3")
					err = localStackManager.SetupS3Bucket(bucketName, tmpFile.Name(), objectKey)
					if err != nil {
						t.Fatalf("Failed to re-upload file to S3 in PreConfig: %v", err)
					}
				},
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, s3URI, util.NullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourcePath, "filename", s3URI),
					provider.TestDefaultHomeDivision(fullResourcePath),
					// Verify that the resource ID has changed due to file_content_hash change
					func(s *terraform.State) error {
						rs := s.RootModule().Resources[fullResourcePath]
						newResourceID := rs.Primary.ID
						t.Logf("Updated response asset has new ID: %s", newResourceID)

						if newResourceID == initialResourceID {
							return fmt.Errorf("expected resource ID to change due to file_content_hash change, but it remained the same: %s", newResourceID)
						}

						t.Logf("Successfully verified resource recreation: old ID %s -> new ID %s", initialResourceID, newResourceID)
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      fullResourcePath,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"file_content_hash",
					"filename",
				},
			},
		},
		CheckDestroy: testVerifyResponseAssetDestroyed,
	})
}
