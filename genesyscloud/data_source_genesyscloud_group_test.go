package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroup(t *testing.T) {
	var (
		groupResource   = "test-group-members"
		groupDataSource = "group-data"
		groupName       = "test group" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateGroupResource(
					groupResource,
					groupName,
					nullValue, // No description
					nullValue, // Default type
					nullValue, // Default visibility
					nullValue, // Default rules_visible
				) + generateGroupDataSource(
					groupDataSource,
					groupName,
					"genesyscloud_group."+groupResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_group."+groupDataSource, "id", "genesyscloud_group."+groupResource, "id"),
				),
			},
		},
	})
}

func generateGroupDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_group" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
