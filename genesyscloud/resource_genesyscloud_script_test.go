package genesyscloud

import (
	"fmt"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceScriptBasic(t *testing.T) {
	var (
		resourceId = "script"
		name       = "testscriptname" + uuid.NewString()
		fileName   = testrunner.GetTestDataPath("resource", "genesyscloud_script", "test_script.json")
	)
	fullyQualifiedPath, _ := filepath.Abs(fileName)
	scriptResource := fmt.Sprintf(`
resource "genesyscloud_script" "%s" {
    script_name       = "%s"
    filename          = "%s"
	file_content_hash = filesha256("%s")
}
`, resourceId, name, fileName, fullyQualifiedPath)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: scriptResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "script_name", name),
					resource.TestCheckResourceAttr("genesyscloud_script."+resourceId, "filename", fileName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_script." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					//"script_name",
					"filename",
					"file_content_hash",
				},
			},
		},
	})
}
