package genesyscloud

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

func TestAccResourceScriptBasic(t *testing.T) {
	t.Parallel()
	var (
		resourceId    = "script"
		name          = "testscriptname" + uuid.NewString()
		nameUpdated   = "testscriptname" + uuid.NewString()
		filePath      = testrunner.GetTestDataPath("resource", "genesyscloud_script", "test_script.json")
		substitutions = make(map[string]string, 0)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceId,
					name,
					filePath,
					generateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "script_name", name),
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "filepath", filePath),
				),
			},
			// Update
			{
				Config: generateScriptResource(
					resourceId,
					nameUpdated,
					filePath,
					generateSubstitutionsMap(substitutions),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "script_name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "filepath", filePath),
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
