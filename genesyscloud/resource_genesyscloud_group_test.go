package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceGroupBasic(t *testing.T) {
	t.Parallel()
	var (
		groupResource1 = "test-group1"
		groupName      = "terraform-" + uuid.NewString()
		groupDesc1     = "Terraform Group Description 1"
		groupDesc2     = "Terraform Group Description 2"
		typeOfficial   = "official" // Default
		visPublic      = "public"   // Default
		visMembers     = "members"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic group
				Config: generateGroupResource(
					groupResource1,
					groupName,
					strconv.Quote(groupDesc1),
					nullValue, // Default type
					nullValue, // Default visibility
					nullValue, // Default rules_visible
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "type", typeOfficial),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "description", groupDesc1),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "visibility", visPublic),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "rules_visible", trueValue),
				),
			},
			{
				// Update group
				Config: generateGroupResource(
					groupResource1,
					groupName,
					strconv.Quote(groupDesc2),
					strconv.Quote(typeOfficial), // Cannot change type
					strconv.Quote(visMembers),
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "type", typeOfficial),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "description", groupDesc2),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "visibility", visMembers),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "rules_visible", falseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_group." + groupResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGroupsDestroyed,
	})
}

func TestAccResourceGroupAddresses(t *testing.T) {
	t.Parallel()
	var (
		groupResource1 = "test-group-addr"
		groupName      = "TF Group" + uuid.NewString()
		addrPhone1     = "+13174269078"
		addrPhone2     = "+441434634996"
		addrPhoneExt   = "4321"
		addrPhoneExt2  = "4320"
		typeGroupRing  = "GROUPRING"
		typeGroupPhone = "GROUPPHONE"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						strconv.Quote(addrPhone1),
						typeGroupRing,
						nullValue, // No extension
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.number", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupRing),
				),
			},
			{
				// Update phone number & type
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						strconv.Quote(addrPhone2),
						typeGroupPhone,
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupPhone),
				),
			},
			{
				// Remove number and set extension
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						nullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupPhone),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.extension", addrPhoneExt),
				),
			},
			{
				// Update the extension
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName,
					generateGroupAddress(
						nullValue,
						typeGroupPhone,
						strconv.Quote(addrPhoneExt2),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "name", groupName),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.type", typeGroupPhone),
					resource.TestCheckResourceAttr("genesyscloud_group."+groupResource1, "addresses.0.extension", addrPhoneExt2),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_group." + groupResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"addresses"},
			},
		},
		CheckDestroy: testVerifyGroupsDestroyed,
	})
}

func TestAccResourceGroupMembers(t *testing.T) {
	t.Parallel()
	var (
		groupResource = "test-group-members"
		groupName     = "Terraform Test Group-" + uuid.NewString()
		userResource1 = "group-user1"
		userResource2 = "group-user2"
		userEmail1    = "terraform1-" + uuid.NewString() + "@example.com"
		userEmail2    = "terraform2-" + uuid.NewString() + "@example.com"
		userName1     = "Johnny Terraform"
		userName2     = "Ryan Terraform"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create group with an owner and a member
				Config: GenerateBasicGroupResource(
					groupResource,
					groupName,
					generateGroupOwners("genesyscloud_user."+userResource1+".id"),
					generateGroupMembers("genesyscloud_user."+userResource2+".id"),
				) + GenerateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + GenerateBasicUserResource(
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
				Config: GenerateBasicGroupResource(
					groupResource,
					groupName,
					generateGroupOwners("genesyscloud_user."+userResource1+".id"),
					generateGroupMembers(
						"genesyscloud_user."+userResource1+".id",
						"genesyscloud_user."+userResource2+".id",
					),
				) + GenerateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + GenerateBasicUserResource(
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
				Config: GenerateBasicGroupResource(
					groupResource,
					groupName,
					generateGroupOwners("genesyscloud_user."+userResource2+".id"),
					generateGroupMembers(
						"genesyscloud_user."+userResource1+".id",
					),
				) + GenerateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + GenerateBasicUserResource(
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
				Config: GenerateBasicGroupResource(
					groupResource,
					groupName,
					generateGroupOwners("genesyscloud_user."+userResource2+".id"),
					"member_ids = []",
				) + GenerateBasicUserResource(
					userResource2,
					userEmail2,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_group."+groupResource, "member_ids.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_group." + groupResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGroupsDestroyed,
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
		} else if IsStatus404(resp) {
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
