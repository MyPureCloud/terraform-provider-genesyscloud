package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFlow(t *testing.T) {
	// TODO: Generate a real flow once the resource has been added
	t.Skip("skipping flow data source test until resource is defined")

	var (
		flowDataSource = "flow-data"
		flowName       = "test flow"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateFlowDataSource(
					flowDataSource,
					flowName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_flow."+flowDataSource, "id", "88985b85-924a-4c4b-ad1b-2c43da23b6a8"),
				),
			},
		},
	})
}

func generateFlowDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_flow" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
