package scripts

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
   Testcases for the resources schema
*/

func getTestDataPath(elem ...string) string {
	basePath := filepath.Join("..", "..", "test", "data")
	subPath := filepath.Join(elem...)
	return filepath.Join(basePath, subPath)
}

func TestAccResourceScriptBasic(t *testing.T) {
	var (
		resourceId    = "script"
		name          = "testscriptname" + uuid.NewString()
		nameUpdated   = "testscriptname" + uuid.NewString()
		filePath      = getTestDataPath("resource", resourceName, "test_script.json")
		substitutions = make(map[string]string)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceId,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "script_name", name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "filepath", filePath),
					validateScriptPublished(resourceName+"."+resourceId),
				),
			},
			// Update
			{
				Config: generateScriptResource(
					resourceId,
					nameUpdated,
					filePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "script_name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "filepath", filePath),
					validateScriptPublished(resourceName+"."+resourceId),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceName + "." + resourceId,
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
		resourceId          = "script-subs"
		name                = "testscriptname" + uuid.NewString()
		filePath            = getTestDataPath("resource", resourceName, "test_script.json")
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
					resourceId,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "script_name", name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "filepath", filePath),
					validateScriptPublished(resourceName+"."+resourceId),
					getScriptId(resourceName+"."+resourceId, &scriptIdAfterCreate),
				),
			},
			// Update
			{
				Config: generateScriptResource(
					resourceId,
					name,
					filePath,
					util.GenerateSubstitutionsMap(substitutionsUpdate),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "script_name", name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "filepath", filePath),
					validateScriptPublished(resourceName+"."+resourceId),
					getScriptId(resourceName+"."+resourceId, &scriptIdAfterUpdate),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceName + "." + resourceId,
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

// getScriptId retrieves the script GUID from the state
func getScriptId(scriptResourceName string, id *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		scriptResource, ok := state.RootModule().Resources[scriptResourceName]
		if !ok {
			return fmt.Errorf("failed to find script %s in state", scriptResourceName)
		}
		*id = scriptResource.Primary.ID
		return nil
	}
}

func generateScriptResource(resourceId, scriptName, filePath, substitutions string) string {
	fullyQualifiedPath, _ := filepath.Abs(filePath)
	return fmt.Sprintf(`
resource "%s" "%s" {
	script_name       = "%s"
	filepath          = "%s"
	file_content_hash = filesha256("%s")
	%s
}	
	`, resourceName, resourceId, scriptName, filePath, fullyQualifiedPath, substitutions)
}

func testVerifyScriptDestroyed(state *terraform.State) error {
	scriptsAPI := platformclientv2.NewScriptsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != resourceName {
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
func validateScriptPublished(scriptResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		scriptResource, ok := state.RootModule().Resources[scriptResourceName]
		if !ok {
			return fmt.Errorf("Failed to find script %s in state", scriptResourceName)
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
