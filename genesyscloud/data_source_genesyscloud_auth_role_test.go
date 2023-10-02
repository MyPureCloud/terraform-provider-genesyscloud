package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthRole(t *testing.T) {
	var (
		roleResource   = "auth-role"
		roleDataSource = "auth-role-data"
		roleName       = "Terraform Role-" + uuid.NewString()
		roleDesc       = "Terraform test role"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateAuthRoleResource(
					roleResource,
					roleName,
					roleDesc,
				) + generateAuthRoleDataSource(
					roleDataSource,
					"genesyscloud_auth_role."+roleResource+".name",
					"genesyscloud_auth_role."+roleResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_auth_role."+roleDataSource, "id", "genesyscloud_auth_role."+roleResource, "id"),
				),
			},
		},
	})
}

func generateAuthRoleDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_role" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}

func generateDefaultAuthRoleDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_role" "%s" {
		name = %s
	}
	`, resourceID, name)
}
