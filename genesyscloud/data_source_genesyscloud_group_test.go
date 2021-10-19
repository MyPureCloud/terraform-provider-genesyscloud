package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroup(t *testing.T) {
	// TODO: Generate a real script once the resource has been added
	t.Skip("skipping group data source test until resource is defined")

	var (
		groupDataSource = "group-data"
		groupName       = "test group"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateGroupDataSource(
					groupDataSource,
					groupName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_group."+groupDataSource, "id", ""),
				),
			},
		},
	})
}

func generateGroupDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_group" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
