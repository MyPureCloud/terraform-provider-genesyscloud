package scripts

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

/*
   Testcases for the resources schema
*/

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
	file_content_hash = filesha256("%s")
	%s
}
	`, ResourceType, resourceLabel, scriptName, filePath, filePath, substitutions)
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

func TestAccResourceScriptS3File(t *testing.T) {
	var (
		resourceLabel = "script-s3"
		name          = "testscriptname" + uuid.NewString()
		filePath      = "s3://test-bucket/scripts/test_script.json"
		substitutions = make(map[string]string)
	)

	substitutions["foo"] = "bar"

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

func TestAccResourceScriptS3FileUpdate(t *testing.T) {
	var (
		resourceLabel       = "script-s3-update"
		name                = "testscriptname" + uuid.NewString()
		filePath            = "s3://test-bucket/scripts/test_script.json"
		filePathUpdated     = "s3://test-bucket/scripts/test_script_updated.json"
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
			// Update with different S3 file
			{
				Config: generateScriptResource(
					resourceLabel,
					name,
					filePathUpdated,
					util.GenerateSubstitutionsMap(substitutionsUpdate),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "script_name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "filepath", filePathUpdated),
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

func TestAccResourceScriptMixedFiles(t *testing.T) {
	var (
		resourceLabel = "script-mixed"
		name1         = "testscriptname" + uuid.NewString()
		name2         = "testscriptname" + uuid.NewString()
		localFilePath = testrunner.GetTestDataPath("resource", ResourceType, "test_script.json")
		s3FilePath    = "s3://test-bucket/scripts/test_script.json"
		substitutions = make(map[string]string)
	)

	substitutions["foo"] = "bar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceLabel+"1",
					name1,
					localFilePath,
					util.GenerateSubstitutionsMap(substitutions),
				) + generateScriptResource(
					resourceLabel+"2",
					name2,
					s3FilePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel+"1", "script_name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel+"1", "filepath", localFilePath),
					validateScriptPublished(ResourceType+"."+resourceLabel+"1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel+"2", "script_name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel+"2", "filepath", s3FilePath),
					validateScriptPublished(ResourceType+"."+resourceLabel+"2"),
				),
			},
		},
		CheckDestroy: testVerifyScriptDestroyed,
	})
}
