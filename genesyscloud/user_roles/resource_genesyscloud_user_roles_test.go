package user_roles

import (
	"fmt"
	"strconv"
	"strings"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/user"
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
		empRoleDataSourceLabel = "employee-role"
		empRoleName            = "employee"
		userRoleResourceLabel  = "test-user-roles2"
		userResourceLabel1     = "test-user"
		email1                 = "terraform-" + uuid.NewString() + "@example.com"
		userName1              = "Role Terraform"
		roleResourceLabel1     = "test-role-1"
		roleResourceLabel2     = "test-role-2"
		roleName1              = "Terraform User Role Test1" + uuid.NewString()
		roleName2              = "Terraform User Role Test2" + uuid.NewString()
		roleDesc               = "Terraform user roles test"
		divResourceLabel       = "test-division"
		divName                = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 role in default division
				// Also add employee role reference as new user's automatically get this role
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}\n" + user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResourceLabel,
					userResourceLabel1,
					generateResourceRoles("genesyscloud_auth_role."+roleResourceLabel1+".id", "data.genesyscloud_auth_division_home.home.id"),
					generateResourceRoles("data.genesyscloud_auth_role."+empRoleDataSourceLabel+".id"),
				) + authRole.GenerateDefaultAuthRoleDataSource(
					empRoleDataSourceLabel,
					strconv.Quote(empRoleName),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResourceLabel, "roles.0.division_ids.#", "1"),
					resource.TestCheckResourceAttrPair("genesyscloud_user_roles."+userRoleResourceLabel, "roles.0.division_ids.0",
						"data.genesyscloud_auth_division_home.home", "id"),
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResourceLabel, "roles.1.division_ids.#", "0"),
					validateResourceRole("genesyscloud_user_roles."+userRoleResourceLabel, "genesyscloud_auth_role."+roleResourceLabel1),
				),
			},
			{
				// Create another role and division and add to the user
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}\n" + user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc,
				) + authRole.GenerateAuthRoleResource(
					roleResourceLabel2,
					roleName2,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResourceLabel,
					userResourceLabel1,
					generateResourceRoles("genesyscloud_auth_role."+roleResourceLabel1+".id"),
					generateResourceRoles("genesyscloud_auth_role."+roleResourceLabel2+".id",
						"genesyscloud_auth_division."+divResourceLabel+".id", "data.genesyscloud_auth_division_home.home.id"),
				) + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResourceLabel, "roles.1.division_ids.#", "0"),
					resource.TestCheckResourceAttr("genesyscloud_user_roles."+userRoleResourceLabel, "roles.0.division_ids.#", "2"),
					validateResourceRole("genesyscloud_user_roles."+userRoleResourceLabel, "genesyscloud_auth_role."+roleResourceLabel1),
					validateResourceRole("genesyscloud_user_roles."+userRoleResourceLabel, "genesyscloud_auth_role."+roleResourceLabel2,
						"genesyscloud_auth_division."+divResourceLabel, "data.genesyscloud_auth_division_home.home"),
				),
			},
			{
				// Remove a role from the user and modify division
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResourceLabel,
					userResourceLabel1,
					generateResourceRoles("genesyscloud_auth_role."+roleResourceLabel1+".id", "genesyscloud_auth_division."+divResourceLabel+".id"),
				) + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName),
				Check: resource.ComposeTestCheckFunc(
					validateResourceRole("genesyscloud_user_roles."+userRoleResourceLabel, "genesyscloud_auth_role."+roleResourceLabel1, "genesyscloud_auth_division."+divResourceLabel),
				),
			},
			{
				// Remove all roles from the user
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + authRole.GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc,
				) + GenerateUserRoles(
					userRoleResourceLabel,
					userResourceLabel1,
				) + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_user_roles."+userRoleResourceLabel, "roles.%"),
				),
			},
		},
	})
}

func validateResourceRole(resourcePath string, roleResourcePath string, divisions ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Failed to find %s in state", resourcePath)
		}
		resourceLabel := resourceState.Primary.ID

		roleResource, ok := state.RootModule().Resources[roleResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourcePath)
		}
		roleID := roleResource.Primary.ID

		homeDivID, err := util.GetHomeDivisionID()
		if err != nil {
			return fmt.Errorf("failed to retrieve home division ID: %v", err)
		}

		if len(divisions) > 0 && divisions[0] != "*" {
			// Get the division IDs from state
			divisionIDs := make([]string, len(divisions))
			for i, divResourcePath := range divisions {
				divResource, ok := state.RootModule().Resources[divResourcePath]
				if !ok {
					return fmt.Errorf("failed to find %s in state", divResourcePath)
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
		return fmt.Errorf("Missing expected role for resource %s in state: %s", resourceLabel, roleID)
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
