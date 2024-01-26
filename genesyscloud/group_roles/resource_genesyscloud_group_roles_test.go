package group_roles

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud"
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
		testUserResource  = "user_resource1"
		testUserName      = "nameUser1" + uuid.NewString()
		testUserEmail     = uuid.NewString() + "@example.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { genesyscloud.TestAccPreCheck(t) },
		ProviderFactories: genesyscloud.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create group with 1 role in default division
				Config: genesyscloud.GenerateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + genesyscloud.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					genesyscloud.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + genesyscloud.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
					genesyscloud.GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					genesyscloud.ValidateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1),
				),
			},
			{
				// Create another role and division and add to the group
				Config: genesyscloud.GenerateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + genesyscloud.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					genesyscloud.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + genesyscloud.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + genesyscloud.GenerateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
					genesyscloud.GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					genesyscloud.GenerateResourceRoles("genesyscloud_auth_role."+roleResource2+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					genesyscloud.ValidateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1),
					genesyscloud.ValidateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource2, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove a role from the group and modify division
				Config: genesyscloud.GenerateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + genesyscloud.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					genesyscloud.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + genesyscloud.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
					genesyscloud.GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					genesyscloud.ValidateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove all roles from the group
				Config: genesyscloud.GenerateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + genesyscloud.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					genesyscloud.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + genesyscloud.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
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
