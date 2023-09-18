package genesyscloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGroupRolesMembership(t *testing.T) {
	t.Parallel()
	var (
		groupRoleResource = "test-group-roles1"
		groupResource1    = "test-group"
		groupName         = "terraform-" + uuid.NewString()
		roleResource1     = "test-role-1"
		roleResource2     = "test-role-2"
		roleName1         = "Terraform Group Role Test1" + uuid.NewString()
		roleName2         = "Terraform Group Role Test2" + uuid.NewString()
		roleDesc          = "Terraform Group roles test"
		divResource       = "test-division"
		divName           = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create group with 1 role in default division
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1),
				),
			},
			{
				// Create another role and division and add to the group
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource2+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + generateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1),
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource2, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove a role from the group and modify division
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + generateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove all roles from the group
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
				) + generateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_group_roles."+groupRoleResource, "roles.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_group_roles." + groupRoleResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateGroupRoles(resourceID string, groupResource string, roles ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_group_roles" "%s" {
		group_id = genesyscloud_group.%s.id
		%s
	}
	`, resourceID, groupResource, strings.Join(roles, "\n"))
}
