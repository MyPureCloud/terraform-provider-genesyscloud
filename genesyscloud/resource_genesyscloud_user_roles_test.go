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
		userRoleResource = "test-user-roles"
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create user with 1 role in default division
				// Also add employee role reference as new user's automatically get this role
				Config: generateBasicUserResource(
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
				Config: generateBasicUserResource(
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
				Config: generateBasicUserResource(
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
				Config: generateBasicUserResource(
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
					resource.TestCheckNoResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles"),
				),
			},
		},
	})
}

//func TestUserRolesLoop(t *testing.T) {
//	for {
//		os.Remove("/Users/ronanwatkins/genesys_src/repos/terraform-provider-genesyscloud/genesyscloud/sdk_debug.log")
//		TestUserRoles(t)
//	}
//}

func TestUserRoles(t *testing.T) {
	config := fmt.Sprintf(`# Built-in roles, not managed by terraform
data "genesyscloud_auth_role" "employee" {
  name = "employee"
}
data "genesyscloud_auth_role" "admin" {
  name = "admin"
}

# Custom roles
resource "genesyscloud_auth_role" "merge_role" {
  name                = "%s"
  description         = "Merge-only role"
  permission_policies {
    domain      = "externalContacts"
    entity_name = "contact"
    action_set  = ["view"]
  }
  permission_policies {
    domain      = "externalContacts"
    entity_name = "identity"
    action_set  = ["merge"]
  }
}
resource "genesyscloud_auth_role" "promote_role" {
  name                = "%s"
  description         = "Promote-only role"
  permission_policies {
    domain      = "externalContacts"
    entity_name = "contact"
    action_set  = ["view"]
  }
  permission_policies {
    domain      = "externalContacts"
    entity_name = "identity"
    action_set  = ["promote"]
  }
}
resource "genesyscloud_auth_role" "restricted_role" {
  name                = "%s"
  description         = "No relate permissions role"
}

# Users and their role associations
resource "genesyscloud_user" "employee" {
  email    = "%s"
  name     = "employee"
  password = "Test1234!"
}
resource "genesyscloud_user_roles" "employee-roles" {
  user_id = genesyscloud_user.employee.id
  roles {
    role_id = data.genesyscloud_auth_role.employee.id
  }
}

resource "genesyscloud_user" "merge_user" {
  email    = "%s"
  name     = "merge_user"
  password = "Test1234!"
}
resource "genesyscloud_user_roles" "merge_user-roles" {
  user_id = genesyscloud_user.merge_user.id
  roles {
    role_id = genesyscloud_auth_role.merge_role.id
  }
}

resource "genesyscloud_user" "promote_user" {
  email    = "%s"
  name     = "promote_user"
  password = "Test1234!"
}
resource "genesyscloud_user_roles" "promote_user-roles" {
  user_id = genesyscloud_user.promote_user.id
  roles {
    role_id = genesyscloud_auth_role.promote_role.id
  }
}

resource "genesyscloud_user" "restricted_user" {
  email    = "%s"
  name     = "restricted_user"
  password = "Test1234!"
}
resource "genesyscloud_user_roles" "restricted_user-roles" {
  user_id = genesyscloud_user.restricted_user.id
  roles {
    role_id = genesyscloud_auth_role.restricted_role.id
  }
}`, "Merge Role"+uuid.NewString(), "Promote Role"+uuid.NewString(), "Restricted Role"+uuid.NewString(), uuid.NewString()+"employee@relatetest.com", uuid.NewString()+"merge_user@relatetest.com", uuid.NewString()+"promote_user@relatetest.com", uuid.NewString()+"restricted_user@relatetest.com")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create user with 1 role in default division
				// Also add employee role reference as new user's automatically get this role
				Config: config,
				//Check: resource.ComposeTestCheckFunc(
				//	validateResourceRole("genesyscloud_user_roles.restricted_user-roles", "genesyscloud_auth_role.restricted_role"),
				//),
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
