package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceAuthRoleBasic(t *testing.T) {
	var (
		roleResource1 = "auth-role1"
		roleName1     = "Terraform Role-" + uuid.NewString()
		roleDesc1     = "Terraform test role"
		roleDesc2     = "Terraform test role updated"
		perm1         = "group_creation"
		perm2         = "admin"
		directoryDom  = "directory"
		userEntity    = "user"
		groupEntity   = "group"
		allAction     = "*"
		addAction     = "add"
		editAction    = "edit"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc1,
					generateRolePermissions(strconv.Quote(perm1)),
					generateRolePermPolicy(directoryDom, userEntity, strconv.Quote(addAction)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResource1, "name", roleName1),
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResource1, "description", roleDesc1),
					validateRolePermissions("genesyscloud_auth_role."+roleResource1, perm1),
					validatePermissionPolicy("genesyscloud_auth_role."+roleResource1, directoryDom, userEntity, addAction),
				),
			},
			{
				// Update
				Config: generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc2,
					generateRolePermissions(strconv.Quote(perm1), strconv.Quote(perm2)),
					generateRolePermPolicy(directoryDom, userEntity, strconv.Quote(allAction)),
					generateRolePermPolicy(directoryDom, groupEntity, strconv.Quote(addAction), strconv.Quote(editAction)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResource1, "name", roleName1),
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResource1, "description", roleDesc2),
					validateRolePermissions("genesyscloud_auth_role."+roleResource1, perm1, perm2),
					validatePermissionPolicy("genesyscloud_auth_role."+roleResource1, directoryDom, userEntity, allAction),
					validatePermissionPolicy("genesyscloud_auth_role."+roleResource1, directoryDom, groupEntity, addAction, editAction),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_role." + roleResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyRolesDestroyed,
	})
}

func generateAuthRoleResource(
	resourceID string,
	name string,
	description string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_role" "%s" {
		name = "%s"
		description = "%s"
		%s
	}
	`, resourceID, name, description, strings.Join(nestedBlocks, "\n"))
}

func generateRolePermissions(permissions ...string) string {
	return fmt.Sprintf(`
		permissions = [%s]
	`, strings.Join(permissions, ","))
}

func generateRolePermPolicy(domain string, entityName string, actions ...string) string {
	return fmt.Sprintf(` permission_policies {
		domain = "%s"
		entity_name = "%s"
		action_set = [%s]
	}
	`, domain, entityName, strings.Join(actions, ","))
}

func testVerifyRolesDestroyed(state *terraform.State) error {
	authAPI := platformclientv2.NewAuthorizationApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_auth_role" {
			continue
		}

		role, resp, err := authAPI.GetAuthorizationRole(rs.Primary.ID, nil)
		if role != nil {
			return fmt.Errorf("Role (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
			// Role not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All roles destroyed
	return nil
}

func validateRolePermissions(roleResourceName string, permissions ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		roleResource, ok := state.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourceName)
		}

		numPermsAttr, _ := roleResource.Primary.Attributes["permissions.#"]
		numPerms, _ := strconv.Atoi(numPermsAttr)
		configPerms := make([]string, numPerms)
		for i := 0; i < numPerms; i++ {
			configPerms[i] = roleResource.Primary.Attributes["permissions."+strconv.Itoa(i)]
		}

		extraPerms := sliceDifference(configPerms, permissions)
		if len(extraPerms) > 0 {
			return fmt.Errorf("Unexpected permissions found for role %s in state: %v", roleResource.Primary.ID, extraPerms)
		}

		missingPerms := sliceDifference(permissions, configPerms)
		if len(missingPerms) > 0 {
			return fmt.Errorf("Missing expected permissions for role %s in state: %v", roleResource.Primary.ID, missingPerms)
		}

		// All expected permissions found
		return nil
	}
}

func validatePermissionPolicy(roleResourceName string, domain string, entityName string, actionSet ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		roleResource, ok := state.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourceName)
		}

		roleAttrs := roleResource.Primary.Attributes
		numPermsAttr, _ := roleAttrs["permission_policies.#"]
		numPerms, _ := strconv.Atoi(numPermsAttr)
		for i := 0; i < numPerms; i++ {
			if roleAttrs["permission_policies."+strconv.Itoa(i)+".domain"] == domain &&
				roleAttrs["permission_policies."+strconv.Itoa(i)+".entity_name"] == entityName {

				numActionsAttr, _ := roleAttrs["permission_policies."+strconv.Itoa(i)+".action_set.#"]
				numActions, _ := strconv.Atoi(numActionsAttr)
				stateActions := make([]string, numActions)
				for j := 0; j < numActions; j++ {
					stateActions[j] = roleAttrs["permission_policies."+strconv.Itoa(i)+".action_set."+strconv.Itoa(j)]
				}

				extraActions := sliceDifference(stateActions, actionSet)
				if len(extraActions) > 0 {
					return fmt.Errorf("Unexpected permission actions found for role %s in state: %v", roleResource.Primary.ID, extraActions)
				}

				missingActions := sliceDifference(actionSet, stateActions)
				if len(missingActions) > 0 {
					return fmt.Errorf("Missing expected permission actions for role %s in state: %v", roleResource.Primary.ID, missingActions)
				}

				// Found expected policy
				return nil
			}
		}

		return fmt.Errorf("Missing expected permission policy for role %s in state: %s %s", roleResource.Primary.ID, domain, entityName)
	}
}
