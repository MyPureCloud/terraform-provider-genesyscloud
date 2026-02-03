package scripts

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack"
	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
)

func TestAccResourceScriptBasic(t *testing.T) {
	var (
		resourceLabel = "script"
		name          = "testscriptname" + uuid.NewString()
		nameUpdated   = "testscriptname" + uuid.NewString()
		filePath      = testrunner.GetTestDataPath("resource", ResourceType, "test_script.json")
		substitutions = make(map[string]string)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceLabel,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "filepath", filePath),
					validateScriptPublished(ResourceType+"."+resourceLabel),
				),
			},
			// Update
			{
				Config: generateScriptResource(
					resourceLabel,
					nameUpdated,
					filePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", nameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "filepath", filePath),
					validateScriptPublished(ResourceType+"."+resourceLabel),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"filepath",
					"file_content_hash",
					"substitutions",
				},
			},
		},
		CheckDestroy: testVerifyScriptDestroyed,
	})
}

func TestAccResourceScriptUpdate(t *testing.T) {
	var (
		resourceLabel       = "script-subs"
		name                = "testscriptname" + uuid.NewString()
		filePath            = testrunner.GetTestDataPath("resource", ResourceType, "test_script.json")
		substitutions       = make(map[string]string)
		substitutionsUpdate = make(map[string]string)

		scriptIdAfterCreate string
		scriptIdAfterUpdate string
	)

	substitutions["foo"] = "bar"
	substitutionsUpdate["hello"] = "world"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceLabel,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "filepath", filePath),
					validateScriptPublished(ResourceType+"."+resourceLabel),
					getScriptId(ResourceType+"."+resourceLabel, &scriptIdAfterCreate),
				),
			},
			// Update
			{
				Config: generateScriptResource(
					resourceLabel,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutionsUpdate),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "filepath", filePath),
					validateScriptPublished(ResourceType+"."+resourceLabel),
					getScriptId(ResourceType+"."+resourceLabel, &scriptIdAfterUpdate),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"filepath",
					"file_content_hash",
					"substitutions",
				},
			},
		},
		CheckDestroy: testVerifyScriptDestroyed,
	})

	if scriptIdAfterCreate != scriptIdAfterUpdate {
		t.Errorf("Expected script ID to remain the same after update. Before: %s After: %s", scriptIdAfterCreate, scriptIdAfterUpdate)
	}
}

// TestAccDefaultScriptIDS validates that the default script IDs we have defined as constants are still correct.
// In every Genesys Cloud org, there are default scripts that share the same ID. This test case will alert us if that ID
// changes, or if they take a different approach as opposed to reusing **Globally Unique Identifiers**
func TestAccDefaultScriptIDS(t *testing.T) {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		t.Skipf("Skipping because we failed to authorize client credentials: %s", err.Error())
	}

	apiInstance := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	allDefaultScriptIDs := []string{
		constants.DefaultCallbackScriptID,
		constants.DefaultInboundScriptID,
		constants.DefaultOutboundScriptID,
	}

	for _, id := range allDefaultScriptIDs {
		t.Logf("Reading '%s' by '%s'", constants.DefaultScriptMap[id], id)
		_, resp, err := apiInstance.GetScriptsPublishedScriptId(id, "")
		if err == nil {
			t.Logf("Successfully read '%s' by '%s'", constants.DefaultScriptMap[id], id)
			continue
		}
		if util.IsStatus404(resp) {
			t.Fatalf("Expected '%s' to be found using ID '%s'. Error: %s", constants.DefaultScriptMap[id], id, err.Error())
		}
		t.Fatalf("Unexpected error occurred while validating script '%s' by ID '%s': %s", constants.DefaultScriptMap[id], id, err.Error())
	}
}

// TestAccResourceScriptS3 tests the script resource using LocalStack for S3 operations.
// This test validates that the terraform-provider-genesyscloud can successfully deploy scripts from S3 buckets
// using LocalStack as a local AWS service emulator.
//
// Prerequisites:
//   - LocalStack must be running (either locally or in CI)
//   - Environment variables must be set:
//   - USE_LOCAL_STACK=true
//   - LOCAL_STACK_IMAGE_URI=<localstack-image-uri>
//
// Key Features Tested:
//   - S3 file path support (s3://bucket/key format)
//   - Script creation from S3 sources
//   - Script publication verification
//   - Import state functionality
//   - Proper S3 resource cleanup
func TestAccResourceScriptS3(t *testing.T) {
	var (
		resourceLabel = "script_s3"
		name          = "testscriptname_s3" + uuid.NewString()
		filePath      = testrunner.GetTestDataPath("resource", ResourceType, "test_script.json")
		substitutions = make(map[string]string)
	)

	/*
		// Set up LocalStack environment variables
		// To run this test locally, uncomment this block and run `localstack start` from another terminal
		// See more about localstack cli here: https://docs.localstack.cloud/aws/getting-started/installation/
		os.Setenv(localStackEnv.UseLocalStackEnvVar, "true")
		os.Setenv(localStackEnv.LocalStackImageUriEnvVar, "localstack/localstack:latest")
	*/

	// Check if LocalStack is available
	imageURI := os.Getenv(localStackEnv.LocalStackImageUriEnvVar)
	if imageURI == "" || !localStackEnv.LocalStackIsActive() {
		t.Skipf("Missing env variables (%s or %s), indicating that localstack is not running", localStackEnv.LocalStackImageUriEnvVar, localStackEnv.UseLocalStackEnvVar)
	}

	// Initialize LocalStack manager for S3 operations
	ctx := context.Background()
	localStackManager, err := localstack.NewLocalStackManager(ctx)
	if err != nil {
		t.Fatalf("Failed to initialise LocalStackManager: %s", err.Error())
	}

	// Test data setup - create unique bucket name and object key
	bucketName := "test-script-bucket-" + uuid.NewString()
	objectKey := "test_script.json"

	// Set up S3 bucket and upload original file
	// This creates the S3 bucket and uploads the test script file for S3-based testing
	err = localStackManager.SetupS3Bucket(bucketName, filePath, objectKey)
	if err != nil {
		t.Fatalf("Failed to set up S3 bucket: %v", err)
	}

	defer func() {
		if err := localStackManager.CleanupS3Bucket(bucketName); err != nil {
			t.Logf("Warning: failed to cleanup bucket: %v", err)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResourceS3(
					resourceLabel,
					name,
					bucketName,
					objectKey,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "filepath", "s3://"+bucketName+"/"+objectKey),
					validateScriptPublished(ResourceType+"."+resourceLabel),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"filepath",
					"file_content_hash",
					"substitutions",
				},
			},
		},
		CheckDestroy: testVerifyScriptDestroyed,
	})
}

// getScriptId retrieves the script GUID from the state
func getScriptId(scriptResourcePath string, id *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		scriptResource, ok := state.RootModule().Resources[scriptResourcePath]
		if !ok {
			return fmt.Errorf("failed to find script %s in state", scriptResourcePath)
		}
		*id = scriptResource.Primary.ID
		return nil
	}
}

func generateScriptResource(resourceLabel, scriptName, filePath, substitutions string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	script_name       = "%s"
	filepath          = "%s"
	%s
}
	`, ResourceType, resourceLabel, scriptName, filePath, substitutions)
}

// generateScriptResourceS3 generates a terraform configuration for testing scripts with S3 file sources.
// This function creates a resource configuration that uses S3 paths (s3://bucket/key) instead of local file paths.
func generateScriptResourceS3(resourceLabel, scriptName, bucketName, objectKey, substitutions string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	script_name       = "%s"
	filepath          = "s3://%s/%s"
	%s
}
	`, ResourceType, resourceLabel, scriptName, bucketName, objectKey, substitutions)
}

func testVerifyScriptDestroyed(state *terraform.State) error {
	scriptsAPI := platformclientv2.NewScriptsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		script, resp, err := scriptsAPI.GetScript(rs.Primary.ID)
		if script != nil {
			return fmt.Errorf("Script (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == http.StatusNotFound {
			// Script not found, as expected.
			log.Printf("Script (%s) successfully deleted", rs.Primary.ID)
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All Scripts destroyed
	return nil
}

// validateScriptPublished checks to see if the script has been published after it was created
func validateScriptPublished(scriptResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		scriptResource, ok := state.RootModule().Resources[scriptResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find script %s in state", scriptResourcePath)
		}

		scriptID := scriptResource.Primary.ID
		scriptsAPI := platformclientv2.NewScriptsApi()

		script, resp, err := scriptsAPI.GetScriptsPublishedScriptId(scriptID, "")

		//if response == 200
		if resp.StatusCode == http.StatusOK && *script.Id == scriptID {
			return nil
		}

		//If the item is not found this indicates it is not published
		if resp.StatusCode == http.StatusNotFound && err == nil {
			return fmt.Errorf("Script %s was created, but not published.", scriptID)
		}

		//Some APIs will return an error code even if the response code is a 404.
		if resp.StatusCode == http.StatusNotFound && err == nil {
			return fmt.Errorf("Script %s was created, but not published.", scriptID)
		}

		//Err
		if err != nil {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
		return nil
	}
}
