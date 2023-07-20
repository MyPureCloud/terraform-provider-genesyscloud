package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUserRolesMembership(t *testing.T) {
	t.Parallel()
	var (
		empRoleDataSrc   = "employee-role"
		empRoleName      = "employee"
		userRoleResource = "test-user-roles2"
		userResource1    = "test-user"
		email1           = "terraform-" + uuid.NewString() + "@example.com"
		userName1        = "Role Terraform"
		roleResource1    = "test-role-1"
		roleResource2    = "test-role-2"
		roleName1        = "Terraform User Role Test1" + uuid.NewString()
		roleName2        = "Terraform User Role Test2" + uuid.NewString()
		roleDesc         = "Terraform user roles test"
		divResource      = "test-division"
		divName          = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 role in default division
				// Also add employee role reference as new user's automatically get this role
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateUserRoles(
					userRoleResource,
					userResource1,
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					generateResourceRoles("data.genesyscloud_auth_role."+empRoleDataSrc+".id"),
				) + generateDefaultAuthRoleDataSource(
					empRoleDataSrc,
					strconv.Quote(empRoleName),
				),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource1),
				),
			},
			{
				// Create another role and division and add to the user
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				) + generateUserRoles(
					userRoleResource,
					userResource1,
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					generateResourceRoles("genesyscloud_auth_role."+roleResource2+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + generateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource1),
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource2, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove a role from the user and modify division
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateUserRoles(
					userRoleResource,
					userResource1,
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + generateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource1, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove all roles from the user
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateUserRoles(
					userRoleResource,
					userResource1,
				) + generateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.%"),
				),
			},
		},
	})
}

func generateUserRoles(resourceID string, userResource string, roles ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user_roles" "%s" {
		user_id = genesyscloud_user.%s.id
		%s
	}
	`, resourceID, userResource, strings.Join(roles, "\n"))
}
