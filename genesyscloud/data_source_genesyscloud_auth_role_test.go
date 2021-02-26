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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateAuthRoleResource(
					roleResource,
					roleName,
					roleDesc,
				) + generateAuthRoleDataSource(roleDataSource, "genesyscloud_auth_role."+roleResource+".name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_auth_role."+roleDataSource, "id", "genesyscloud_auth_role."+roleResource, "id"),
				),
			},
		},
	})
}

func generateAuthRoleDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_role" "%s" {
		name = %s
	}
	`, resourceID, name)
}
