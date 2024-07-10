package group_roles

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	mu sync.Mutex
)

func TestAccResourceGroupRolesMembership(t *testing.T) {
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
		userID            string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create group with 1 role in default division
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + group.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + authRole.GenerateAuthRoleResource(
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
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+testUserResource]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+testUserResource)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),
			},
			{
				// Create another role and division and add to the group
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + group.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + authRole.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + authRole.GenerateAuthRoleResource(
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
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + group.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + authRole.GenerateAuthRoleResource(
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
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + group.GenerateBasicGroupResource(
					groupResource1,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + authRole.GenerateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateGroupRoles(
					groupRoleResource,
					groupResource1,
				) + genesyscloud.GenerateAuthDivisionBasic(divResource, divName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_group_roles."+groupRoleResource, "roles.%"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for resources to get deleted properly
						return nil
					},
				),

				PreventPostDestroyRefresh: true,
			},
			{
				Config: generateGroupRoles(
					groupRoleResource,
					groupResource1,
				),
				// Import/Read
				ResourceName:      "genesyscloud_group_roles." + groupRoleResource,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
				Check: resource.ComposeTestCheckFunc(
					checkUserDeleted(userID),
				),
			},
		},
		CheckDestroy: testVerifyGroupsAndUsersDestroyed,
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

// TODO: Duplicating this code within the function to not break a cyclic dependency
func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}

func checkUserDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching user with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 30
		for i := 0; i < maxAttempts; i++ {

			deleted, err := isUserDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("user %s was not deleted properly", id)
	}
}

func isUserDeleted(id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()

	usersAPI := platformclientv2.NewUsersApi()
	// Attempt to get the user
	_, response, err := usersAPI.GetUser(id, nil, "", "")

	// Check if the user is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // User is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If user is found, it means the user is not deleted
	return false, nil
}

func testVerifyGroupsAndUsersDestroyed(state *terraform.State) error {
	groupsAPI := platformclientv2.NewGroupsApi()
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_group" {
			group, resp, err := groupsAPI.GetGroup(rs.Primary.ID)
			if group != nil {
				return fmt.Errorf("Group (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// Group not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
		if rs.Type == "genesyscloud_user" {
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if user != nil {
				return fmt.Errorf("User (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// User not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}

	}
	return nil
}
