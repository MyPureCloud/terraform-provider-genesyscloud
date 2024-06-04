package group

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceGroup(t *testing.T) {
	var (
		groupResource    = "test-group-members"
		groupDataSource  = "group-data"
		groupName        = "test group" + uuid.NewString()
		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@examplegroup.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) +
					GenerateGroupResource(
						groupResource,
						groupName,
						util.NullValue, // No description
						util.NullValue, // Default type
						util.NullValue, // Default visibility
						util.NullValue, // Default rules_visible
						GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
					) + generateGroupDataSource(
					groupDataSource,
					groupName,
					"genesyscloud_group."+groupResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_group."+groupDataSource, "id", "genesyscloud_group."+groupResource, "id"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for resources to get deleted properly
						return nil
					},
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
