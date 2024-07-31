package group

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceGroupBasic(t *testing.T) {
	var (
		groupResource1   = "test-group1"
		groupName        = "terraform-" + uuid.NewString()
		groupDesc1       = "Terraform Group Description 1"
		groupDesc2       = "Terraform Group Description 2"
		typeOfficial     = "official" // Default
		visPublic        = "public"   // Default
		visMembers       = "members"
		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@group.com"
		userID           string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic group
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) +
					GenerateGroupResource(
						groupResource1,
						groupName,
						strconv.Quote(groupDesc1),
						util.NullValue, // Default type
						util.NullValue, // Default visibility
						util.NullValue, // Default rules_visible
						"roles_enabled = false",
						GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "type", typeOfficial),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "description", groupDesc1),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "visibility", visPublic),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "rules_visible", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "roles_enabled", util.FalseValue),
				),
			},
			{
				// Update group
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateGroupResource(
					groupResource1,
					groupName,
					strconv.Quote(groupDesc2),
					strconv.Quote(typeOfficial), // Cannot change type
					strconv.Quote(visMembers),
					util.FalseValue,
					"roles_enabled = true",
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "type", typeOfficial),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "description", groupDesc2),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "visibility", visMembers),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "rules_visible", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "roles_enabled", util.TrueValue),
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

				PreventPostDestroyRefresh: true,
			},
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateGroupResource(
					groupResource1,
					groupName,
					strconv.Quote(groupDesc2),
					strconv.Quote(typeOfficial), // Cannot change type
					strconv.Quote(visMembers),
					util.FalseValue,
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				// Import/Read
				ResourceName:      "genesyscloud_group." + groupResource1,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: testVerifyGroupsAndUsersDestroyed,
	})
}

func TestAccResourceGroupAddresses(t *testing.T) {
	var (
		groupResource1   = "test-group-addr"
		groupName        = "TF Group" + uuid.NewString()
		addrPhone1       = "+13174269078"
		addrPhone2       = "+441434634996"
		addrPhoneExt     = "4321"
		addrPhoneExt2    = "4320"
		typeGroupRing    = "GROUPRING"
		typeGroupPhone   = "GROUPPHONE"
		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@groupadd.com"
		userID           string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						strconv.Quote(addrPhone1),
						typeGroupRing,
						util.NullValue, // No extension
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.number", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupRing),
				),
			},
			{
				// Update phone number & type
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						strconv.Quote(addrPhone2),
						typeGroupPhone,
						util.NullValue,
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupPhone),
				),
			},
			{
				// Remove number and set extension
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						util.NullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt),
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupPhone),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.extension", addrPhoneExt),
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
				// Update the extension
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						util.NullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt2),
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupPhone),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.extension", addrPhoneExt2),
				),

				PreventPostDestroyRefresh: true,
			},
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						util.NullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt2),
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				),
				// Import/Read
				ResourceName:            "genesyscloud_group." + groupResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"addresses"},
				Destroy:                 true,
				Check: resource.ComposeTestCheckFunc(
					checkUserDeleted(userID),
				),
			},
		},
		CheckDestroy: testVerifyGroupsAndUsersDestroyed,
	})
}

func TestAccResourceGroupMembers(t *testing.T) {
	t.Parallel()
	var (
		groupResource    = "test-group-members"
		groupName        = "Terraform Test Group-" + uuid.NewString()
		userResource1    = "group-user1"
		userResource2    = "group-user2"
		userEmail1       = "terraform1-" + uuid.NewString() + "@groupmem.com"
		userEmail2       = "terraform2-" + uuid.NewString() + "@groupmem.com"
		userName1        = "Johnny Terraform"
		userName2        = "Ryan Terraform"
		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@groupmem.com"
		userID           string
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create group with an owner and a member
				Config: GenerateBasicGroupResource(
					groupResource,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResource1+".id"),
					generateGroupMembers("genesyscloud_user."+userResource2+".id"),
				) + generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResource2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource1, "owner_ids"),
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource2, "member_ids"),
				),
			},
			{
				// Make the owner a member
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResource1+".id"),
					generateGroupMembers(
						"genesyscloud_user."+userResource1+".id",
						"genesyscloud_user."+userResource2+".id",
					),
				) + generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResource2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource1, "owner_ids"),
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource1, "member_ids"),
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource2, "member_ids"),
				),
			},
			{
				// Remove a member and change the owner
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResource2+".id"),
					generateGroupMembers(
						"genesyscloud_user."+userResource1+".id",
					),
				) + generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResource2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource2, "owner_ids"),
					validateGroupMember("genesyscloud_group."+groupResource, "genesyscloud_user."+userResource1, "member_ids"),
				),
			},
			{
				// Remove all members while deleting the user
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResource,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResource2+".id"),
					"member_ids = []",
				) + generateBasicUserResource(
					userResource2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_group."+groupResource, "member_ids.%"),
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
				PreventPostDestroyRefresh: true,
			},
			{
				ResourceName:      "genesyscloud_user." + testUserResource,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyGroupsAndUsersDestroyed(state)
		},
	})
}

func testVerifyGroupsDestroyed(state *terraform.State) error {
	groupsAPI := platformclientv2.NewGroupsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_group" {
			continue
		}

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
	// Success. All groups destroyed
	return nil
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
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_user" {
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue //Try one more time
			}
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if user != nil {
				return fmt.Errorf("User Resource (%s) still exists", rs.Primary.ID)
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

func validateGroupMember(groupResourceName string, userResourceName string, attrName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		groupResource, ok := state.RootModule().Resources[groupResourceName]
		if !ok {
			return fmt.Errorf("Failed to find group %s in state", groupResourceName)
		}
		groupID := groupResource.Primary.ID

		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
		}
		userID := userResource.Primary.ID

		numMembersAttr, ok := groupResource.Primary.Attributes[attrName+".#"]
		if !ok {
			return fmt.Errorf("No %s found for group %s in state", attrName, groupID)
		}

		numMembers, _ := strconv.Atoi(numMembersAttr)
		for i := 0; i < numMembers; i++ {
			if groupResource.Primary.Attributes[attrName+"."+strconv.Itoa(i)] == userID {
				// Found user
				return nil
			}
		}

		return fmt.Errorf("%s %s not found for group %s in state", attrName, userID, groupID)
	}
}

// Duplicating this code within the function to not break a cyclid dependency
func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}

// Basic user with minimum required fields
func generateBasicUserResource(resourceID string, email string, name string) string {
	return generateUserResource(resourceID, email, name, util.NullValue, util.NullValue, util.NullValue, util.NullValue, util.NullValue, "", "")
}

func generateUserResource(
	resourceID string,
	email string,
	name string,
	state string,
	title string,
	department string,
	manager string,
	acdAutoAnswer string,
	profileSkills string,
	certifications string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		state = %s
		title = %s
		department = %s
		manager = %s
		acd_auto_answer = %s
		profile_skills = [%s]
		certifications = [%s]
	}
	`, resourceID, email, name, state, title, department, manager, acdAutoAnswer, profileSkills, certifications)
}
