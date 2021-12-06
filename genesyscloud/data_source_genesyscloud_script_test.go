package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceScript(t *testing.T) {
	// TODO: Generate a real script once the resource has been added
	t.Skip("skipping script data source test until resource is defined")

	var (
		scriptDataSource = "script-data"
		scriptName       = "test script"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateScriptDataSource(
					scriptDataSource,
					scriptName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_script."+scriptDataSource, "id", ""),
				),
			},
		},
	})
}

func generateScriptDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_script" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
