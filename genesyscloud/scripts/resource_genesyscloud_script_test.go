package scripts

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
   Testcases for the resources schema
*/

func getTestDataPath(elem ...string) string {
	basePath := filepath.Join("../..", "test", "data")
	subPath := filepath.Join(elem...)
	return filepath.Join(basePath, subPath)
}

func TestAccResourceScriptBasic(t *testing.T) {
	t.Parallel()
	var (
		resourceId    = "script"
		name          = "testscriptname" + uuid.NewString()
		nameUpdated   = "testscriptname" + uuid.NewString()
		filePath      = getTestDataPath("resource", "genesyscloud_script", "test_script.json")
		substitutions = make(map[string]string, 0)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceId,
					name,
					filePath,
					gcloud.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "script_name", name),
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "filepath", filePath),
					validateScriptPublished("genesyscloud_script."+resourceId),
				),
			},
			// Update
			{
				Config: generateScriptResource(
					resourceId,
					nameUpdated,
					filePath,
					gcloud.GenerateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "script_name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "filepath", filePath),
					validateScriptPublished("genesyscloud_script."+resourceId),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_script." + resourceId,
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

func generateScriptResource(resourceId, scriptName, filePath, substitutions string) string {
	fullyQualifiedPath, _ := filepath.Abs(filePath)
	return fmt.Sprintf(`
resource "genesyscloud_script" "%s" {
	script_name       = "%s"
	filepath          = "%s"
	file_content_hash = filesha256("%s")
	%s
}	
	`, resourceId, scriptName, filePath, fullyQualifiedPath, substitutions)
}

func testVerifyScriptDestroyed(state *terraform.State) error {
	scriptsAPI := platformclientv2.NewScriptsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_script" {
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

// validateScriptPublished checks to see if the script has been published after it was creart
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
