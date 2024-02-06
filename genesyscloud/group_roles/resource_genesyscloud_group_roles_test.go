package group_roles

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
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
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1),
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
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					generateResourceRoles("genesyscloud_auth_role."+roleResource2+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1),
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource2, "genesyscloud_auth_division."+divResource),
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
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_group_roles."+groupRoleResource, "genesyscloud_auth_role."+roleResource1, "genesyscloud_auth_division."+divResource),
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

func validateResourceRole(resourceName string, roleResourceName string, divisions ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		roleResource, ok := state.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourceName)
		}
		roleID := roleResource.Primary.ID

		homeDivID, err := genesyscloud.GetHomeDivisionID()
		if err != nil {
			return fmt.Errorf("failed to retrieve home division ID: %v", err)
		}

		if len(divisions) > 0 && divisions[0] != "*" {
			// Get the division IDs from state
			divisionIDs := make([]string, len(divisions))
			for i, divResourceName := range divisions {
				divResource, ok := state.RootModule().Resources[divResourceName]
				if !ok {
					return fmt.Errorf("failed to find %s in state", divResourceName)
				}
				divisionIDs[i] = divResource.Primary.ID
			}
			divisions = divisionIDs
		}

		resourceAttrs := resourceState.Primary.Attributes
		numRolesAttr, _ := resourceAttrs["roles.#"]
		numRoles, _ := strconv.Atoi(numRolesAttr)
		for i := 0; i < numRoles; i++ {
			if resourceAttrs["roles."+strconv.Itoa(i)+".role_id"] == roleID {
				numDivsAttr, _ := resourceAttrs["roles."+strconv.Itoa(i)+".division_ids.#"]
				numDivs, _ := strconv.Atoi(numDivsAttr)
				stateDivs := make([]string, numDivs)
				for j := 0; j < numDivs; j++ {
					stateDivs[j] = resourceAttrs["roles."+strconv.Itoa(i)+".division_ids."+strconv.Itoa(j)]
				}

				extraDivs := lists.SliceDifference(stateDivs, divisions)
				if len(extraDivs) > 0 {
					if len(extraDivs) > 1 || extraDivs[0] != homeDivID {
						return fmt.Errorf("unexpected divisions found for role %s in state: %v", roleID, extraDivs)
					}
				}

				missingDivs := lists.SliceDifference(divisions, stateDivs)
				if len(missingDivs) > 0 {
					return fmt.Errorf("missing expected divisions for role %s in state: %v", roleID, missingDivs)
				}

				// Found expected role and divisions
				return nil
			}
		}
		return fmt.Errorf("Missing expected role for resource %s in state: %s", resourceID, roleID)
	}
}

func generateResourceRoles(skillID string, divisionIds ...string) string {
	var divAttr string
	if len(divisionIds) > 0 {
		divAttr = "division_ids = [" + strings.Join(divisionIds, ",") + "]"
	}
	return fmt.Sprintf(`roles {
		role_id = %s
		%s
	}
	`, skillID, divAttr)
}
