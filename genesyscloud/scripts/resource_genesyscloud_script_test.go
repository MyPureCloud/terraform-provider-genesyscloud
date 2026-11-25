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
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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

// TestAccResourceScriptUpdateWithDuplicateName is a regression test for issue #2670.
// When another script in the org shares the same name as a Terraform-managed script,
// updating the managed script must still succeed: the update path must act on the
// script ID already held in state, not re-resolve the script by name (which fails
// with "more than one script found with name ...").
func TestAccResourceScriptUpdateWithDuplicateName(t *testing.T) {
	var (
		resourceLabel       = "script-dup"
		name                = "testscriptname" + uuid.NewString()
		filePath            = testrunner.GetTestDataPath("resource", ResourceType, "test_script.json")
		substitutions       = make(map[string]string)
		substitutionsUpdate = make(map[string]string)

		scriptIdAfterCreate string
		scriptIdAfterUpdate string
		duplicateScriptId   string
	)

	substitutions["foo"] = "bar"
	substitutionsUpdate["hello"] = "world"

	// Cleanup for the out-of-band duplicate script, which Terraform does not manage.
	defer func() {
		if duplicateScriptId == "" {
			return
		}
		sdkConfig, err := provider.AuthorizeSdk()
		if err != nil {
			t.Logf("Warning: failed to authorize to clean up duplicate script %s: %s", duplicateScriptId, err.Error())
			return
		}
		proxy := getScriptsProxy(sdkConfig)
		if err := proxy.deleteScript(context.Background(), duplicateScriptId); err != nil {
			t.Logf("Warning: failed to delete duplicate script %s: %s", duplicateScriptId, err.Error())
		}
	}()

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
					validateScriptPublished(ResourceType+"."+resourceLabel),
					getScriptId(ResourceType+"."+resourceLabel, &scriptIdAfterCreate),
					// Out-of-band: upload a SECOND script with the same name so the org
					// now contains two scripts named `name`. This mirrors the manual repro.
					createDuplicateNamedScript(name, filePath, &duplicateScriptId),
				),
			},
			// Update the managed script. Before the fix this failed with
			// "more than one script found with name ...". It must now succeed.
			{
				Config: generateScriptResource(
					resourceLabel,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutionsUpdate),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", name),
					validateScriptPublished(ResourceType+"."+resourceLabel),
					getScriptId(ResourceType+"."+resourceLabel, &scriptIdAfterUpdate),
				),
			},
		},
		CheckDestroy: testVerifyScriptDestroyed,
	})

	if scriptIdAfterCreate != scriptIdAfterUpdate {
		t.Errorf("Expected script ID to remain the same after update. Before: %s After: %s", scriptIdAfterCreate, scriptIdAfterUpdate)
	}
}

// createDuplicateNamedScript uploads a second, unmanaged script sharing scriptName and
// captures its ID in duplicateScriptId so the test can clean it up afterwards.
func createDuplicateNamedScript(scriptName, filePath string, duplicateScriptId *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		sdkConfig, err := provider.AuthorizeSdk()
		if err != nil {
			return fmt.Errorf("failed to authorize to create duplicate script: %w", err)
		}
		proxy := getScriptsProxy(sdkConfig)
		ctx := context.Background()

		// Record the IDs that already exist for this name so we can identify the new one.
		before, _, err := proxy.getScriptsByName(ctx, scriptName)
		if err != nil {
			return fmt.Errorf("failed to list scripts named %q before duplicate upload: %w", scriptName, err)
		}
		existing := make(map[string]bool, len(before))
		for _, s := range before {
			existing[*s.Id] = true
		}

		body, err := proxy.uploadScriptFile(filePath, scriptName, "", nil)
		if err != nil {
			return fmt.Errorf("failed to upload duplicate script %q: %w", scriptName, err)
		}
		success, err := proxy.verifyScriptUploadSuccess(ctx, body)
		if err != nil {
			return fmt.Errorf("failed to verify duplicate script upload %q: %w", scriptName, err)
		}
		if !success {
			return fmt.Errorf("duplicate script %q failed to upload successfully", scriptName)
		}

		after, _, err := proxy.getScriptsByName(ctx, scriptName)
		if err != nil {
			return fmt.Errorf("failed to list scripts named %q after duplicate upload: %w", scriptName, err)
		}
		for _, s := range after {
			if !existing[*s.Id] {
				*duplicateScriptId = *s.Id
				break
			}
		}
		if *duplicateScriptId == "" {
			return fmt.Errorf("could not identify the newly created duplicate script named %q", scriptName)
		}
		return nil
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

		if err != nil {
			return fmt.Errorf("error retrieving published script %s: %s", scriptID, err)
		}

		//if response == 200
		if resp != nil && resp.StatusCode == http.StatusOK && *script.Id == scriptID {
			return nil
		}

		//If the item is not found this indicates it is not published
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Script %s was created, but not published.", scriptID)
		}

		return fmt.Errorf("Unexpected response for script %s", scriptID)
	}
}
