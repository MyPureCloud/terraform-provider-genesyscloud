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
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func TestAccResourceGroupBasic(t *testing.T) {
	var (
		groupResourceLabel1   = "test-group1"
		groupName             = "terraform-" + uuid.NewString()
		groupDesc1            = "Terraform Group Description 1"
		groupDesc2            = "Terraform Group Description 2"
		typeOfficial          = "official" // Default
		visPublic             = "public"   // Default
		visMembers            = "members"
		testUserResourceLabel = "user_resource1"
		testUserName          = "nameUser1" + uuid.NewString()
		testUserEmail         = uuid.NewString() + "@group.com"
		userID                string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic group
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) +
					GenerateGroupResource(
						groupResourceLabel1,
						groupName,
						strconv.Quote(groupDesc1),
						util.NullValue, // Default type
						util.NullValue, // Default visibility
						util.NullValue, // Default rules_visible
						"roles_enabled = false",
						GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "type", typeOfficial),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "description", groupDesc1),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "visibility", visPublic),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "rules_visible", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "roles_enabled", util.FalseValue),
				),
			},
			{
				// Update group
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateGroupResource(
					groupResourceLabel1,
					groupName,
					strconv.Quote(groupDesc2),
					strconv.Quote(typeOfficial), // Cannot change type
					strconv.Quote(visMembers),
					util.FalseValue,
					"roles_enabled = true",
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "type", typeOfficial),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "description", groupDesc2),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "visibility", visMembers),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "rules_visible", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "roles_enabled", util.TrueValue),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+testUserResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+testUserResourceLabel)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),

				PreventPostDestroyRefresh: true,
			},
			{
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateGroupResource(
					groupResourceLabel1,
					groupName,
					strconv.Quote(groupDesc2),
					strconv.Quote(typeOfficial), // Cannot change type
					strconv.Quote(visMembers),
					util.FalseValue,
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				// Import/Read
				ResourceName:      "genesyscloud_group." + groupResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(60 * time.Second)
			return testVerifyGroupsAndUsersDestroyed(state)
		},
	})
}

func TestAccResourceGroupAddresses(t *testing.T) {
	var (
		groupResourceLabel1   = "test-group-addr"
		groupName             = "TF Group" + uuid.NewString()
		addrPhone1            = "+13174269078"
		addrPhone2            = "+441434634996"
		addrPhoneExt          = "4321"
		addrPhoneExt2         = "4320"
		typeGroupRing         = "GROUPRING"
		typeGroupPhone        = "GROUPPHONE"
		testUserResourceLabel = "user_resource1"
		testUserName          = "nameUser1" + uuid.NewString()
		testUserEmail         = uuid.NewString() + "@groupadd.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel1,
					groupName,
					generateGroupAddress(
						strconv.Quote(addrPhone1),
						typeGroupRing,
						util.NullValue, // No extension
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.number", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.type", typeGroupRing),
				),
			},
			{
				// Update phone number & type
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel1,
					groupName,
					generateGroupAddress(
						strconv.Quote(addrPhone2),
						typeGroupPhone,
						util.NullValue,
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.type", typeGroupPhone),
				),
			},
			{
				// Remove number and set extension
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel1,
					groupName,
					generateGroupAddress(
						util.NullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt),
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.type", typeGroupPhone),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.extension", addrPhoneExt),
				),
			},
			{
				// Update the extension
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel1,
					groupName,
					generateGroupAddress(
						util.NullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt2),
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.type", typeGroupPhone),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResourceLabel1, "addresses.0.extension", addrPhoneExt2),
				),
			},
			{
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel1,
					groupName,
					generateGroupAddress(
						util.NullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt2),
					),
					GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				),
				// Import/Read
				ResourceName:            "genesyscloud_group." + groupResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"addresses"},
				Destroy:                 true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(60 * time.Second)
			return testVerifyGroupsAndUsersDestroyed(state)
		},
	})
}

func TestAccResourceGroupMembers(t *testing.T) {
	t.Parallel()
	var (
		groupResourceLabel    = "test-group-members"
		groupName             = "Terraform Test Group-" + uuid.NewString()
		userResourceLabel1    = "group-user1"
		userResourceLabel2    = "group-user2"
		userEmail1            = "terraform1-" + uuid.NewString() + "@groupmem.com"
		userEmail2            = "terraform2-" + uuid.NewString() + "@groupmem.com"
		userName1             = "Johnny Terraform"
		userName2             = "Ryan Terraform"
		testUserResourceLabel = "user_resource1"
		testUserName          = "nameUser1" + uuid.NewString()
		testUserEmail         = uuid.NewString() + "@groupmem.com"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create group with an owner and a member
				Config: GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResourceLabel1+".id"),
					generateGroupMembers("genesyscloud_user."+userResourceLabel2+".id"),
				) + generateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResourceLabel2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel1, "owner_ids"),
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel2, "member_ids"),
				),
			},
			{
				// Make the owner a member
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResourceLabel1+".id"),
					generateGroupMembers(
						"genesyscloud_user."+userResourceLabel1+".id",
						"genesyscloud_user."+userResourceLabel2+".id",
					),
				) + generateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResourceLabel2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel1, "owner_ids"),
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel1, "member_ids"),
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel2, "member_ids"),
				),
			},
			{
				// Remove a member and change the owner
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResourceLabel2+".id"),
					generateGroupMembers(
						"genesyscloud_user."+userResourceLabel1+".id",
					),
				) + generateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResourceLabel2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel2, "owner_ids"),
					validateGroupMember("genesyscloud_group."+groupResourceLabel, "genesyscloud_user."+userResourceLabel1, "member_ids"),
				),
			},
			{
				// Remove all members while deleting the user
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
					GenerateGroupOwners("genesyscloud_user."+userResourceLabel2+".id"),
					"member_ids = []",
				) + generateBasicUserResource(
					userResourceLabel2,
					userEmail2,
					userName2,
				),
			},
			{
				ResourceName:      "genesyscloud_user." + testUserResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(60 * time.Second)
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

func validateGroupMember(groupResourcePath string, userResourcePath string, attrName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		groupResource, ok := state.RootModule().Resources[groupResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find group %s in state", groupResourcePath)
		}
		groupID := groupResource.Primary.ID

		userResource, ok := state.RootModule().Resources[userResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourcePath)
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
func generateUserWithCustomAttrs(resourceLabel string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceLabel, email, name, strings.Join(attrs, "\n"))
}

// Basic user with minimum required fields
func generateBasicUserResource(resourceLabel string, email string, name string) string {
	return generateUserResource(resourceLabel, email, name, util.NullValue, util.NullValue, util.NullValue, util.NullValue, util.NullValue, "", "")
}

func generateUserResource(
	resourceLabel string,
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
	`, resourceLabel, email, name, state, title, department, manager, acdAutoAnswer, profileSkills, certifications)
}
