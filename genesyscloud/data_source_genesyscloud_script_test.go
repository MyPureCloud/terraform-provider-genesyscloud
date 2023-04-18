package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceScript(t *testing.T) {
	var (
		scriptDataSource = "script-data"
		resourceId       = "script"
		name             = "tfscript" + uuid.NewString()
		filePath         = testrunner.GetTestDataPath("resource", "genesyscloud_script", "test_script.json")
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
					"",
				) + generateScriptDataSource(
					scriptDataSource,
					name,
					resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_script."+scriptDataSource, "id",
						"genesyscloud_script."+resourceId, "id"),
				),
			},
		},
	})
}

func generateScriptDataSource(dataSourceID, name, resourceId string) string {
	return fmt.Sprintf(`data "genesyscloud_script" "%s" {
		name = "%s"
		depends_on = [genesyscloud_script.%s]
	}
	`, dataSourceID, name, resourceId)
}
