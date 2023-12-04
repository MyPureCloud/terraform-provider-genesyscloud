package genesyscloud

import (
	"strconv"
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
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}\n" + GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "data.genesyscloud_auth_division_home.home.id"),
					GenerateResourceRoles("data.genesyscloud_auth_role."+empRoleDataSrc+".id"),
				) + generateDefaultAuthRoleDataSource(
					empRoleDataSrc,
					strconv.Quote(empRoleName),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.0.division_ids.#", "1"),
					resource.TestCheckResourceAttrPair("genesyscloud_user_roles."+userRoleResource, "roles.0.division_ids.0",
						"data.genesyscloud_auth_division_home.home", "id"),
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.1.division_ids.#", "0"),
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource1),
				),
			},
			{
				// Create another role and division and add to the user
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}\n" + GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource2+".id",
						"genesyscloud_auth_division."+divResource+".id", "data.genesyscloud_auth_division_home.home.id"),
				) + GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.1.division_ids.#", "0"),
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.0.division_ids.#", "2"),
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource1),
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource2,
						"genesyscloud_auth_division."+divResource, "data.genesyscloud_auth_division_home.home"),
				),
			},
			{
				// Remove a role from the user and modify division
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
					GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + GenerateAuthDivisionBasic(divResource, divName),
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
				) + GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
				) + GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.%"),
				),
			},
		},
	})
}
