package user_roles

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 role in default division
				// Also add employee role reference as new user's automatically get this role
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}\n" + genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "data.genesyscloud_auth_division_home.home.id"),
					generateResourceRoles("data.genesyscloud_auth_role."+empRoleDataSrc+".id"),
				) + authRole.GenerateDefaultAuthRoleDataSource(
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
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}\n" + genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + authRole.GenerateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					generateResourceRoles("genesyscloud_auth_role."+roleResource2+".id",
						"genesyscloud_auth_division."+divResource+".id", "data.genesyscloud_auth_division_home.home.id"),
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
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
				Config: genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
					generateResourceRoles("genesyscloud_auth_role."+roleResource1+".id", "genesyscloud_auth_division."+divResource+".id"),
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_user_roles."+userRoleResource, "genesyscloud_auth_role."+roleResource1, "genesyscloud_auth_division."+divResource),
				),
			},
			{
				// Remove all roles from the user
				Config: genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResource,
					userResource1,
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_user_roles."+userRoleResource, "roles.%"),
				),
			},
		},
	})
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

		homeDivID, err := util.GetHomeDivisionID()
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
