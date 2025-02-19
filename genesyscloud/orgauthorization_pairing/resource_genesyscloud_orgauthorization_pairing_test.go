package orgauthorization_pairing

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOrgAuthorizationPairing(t *testing.T) {
	var (
		orgAuthorizationPairingResourceLabel = "test-orgauthorization-pairing"
		userResourceLabel1                   = "test-user-1"
		userResourceLabel2                   = "test-user-2"
		email1                               = "terraform-1-" + uuid.NewString() + "@authpair.com"
		email2                               = "terraform-2-" + uuid.NewString() + "@authpair.com"
		userName1                            = "test user " + uuid.NewString()
		userName2                            = "test user " + uuid.NewString()
		groupResourceLabel1                  = "test-group-1"
		groupResourceLabel2                  = "test-group-2"
		groupName1                           = "TF Group" + uuid.NewString()
		groupName2                           = "TF Group" + uuid.NewString()
		testUserResourceLabel                = "user_resource1"
		testUserName                         = "nameUser1" + uuid.NewString()
		testUserEmail                        = uuid.NewString() + "@authpair.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			// 1 user and 1 group
			{
				PreConfig: func() {
					time.Sleep(45 * time.Second)
				},
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + generateBasicGroupResource(
					groupResourceLabel1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id]
  group_ids = [genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResourceLabel, userResourceLabel1, groupResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids.0", "genesyscloud_user."+userResourceLabel1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids.0", "genesyscloud_group."+groupResourceLabel1, "id"),
				),
			},
			// 2 users and 2 groups
			{
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + user.GenerateBasicUserResource(
					userResourceLabel2,
					email2,
					userName2,
				) + generateBasicGroupResource(
					groupResourceLabel1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + generateBasicGroupResource(
					groupResourceLabel2,
					groupName2,
					generateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id, genesyscloud_user.%s.id]
  group_ids = [genesyscloud_group.%s.id, genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResourceLabel, userResourceLabel1, userResourceLabel2, groupResourceLabel1, groupResourceLabel2),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids",
						"genesyscloud_user."+userResourceLabel1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids",
						"genesyscloud_user."+userResourceLabel2, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids",
						"genesyscloud_group."+groupResourceLabel1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids",
						"genesyscloud_group."+groupResourceLabel2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids.#", "2"),
				),
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
			},
			// 1 user
			{
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id]
}`, orgAuthorizationPairingResourceLabel, userResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids.0", "genesyscloud_user."+userResourceLabel1, "id"),
				),
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
			},
			// 2 users
			{
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				) + user.GenerateBasicUserResource(
					userResourceLabel2,
					email2,
					userName2,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id, genesyscloud_user.%s.id]
}`, orgAuthorizationPairingResourceLabel, userResourceLabel1, userResourceLabel2),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper creation
						return nil
					},
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids",
						"genesyscloud_user."+userResourceLabel1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids",
						"genesyscloud_user."+userResourceLabel2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "user_ids.#", "2"),
				),
			},
			// 1 group
			{
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + generateBasicGroupResource(
					groupResourceLabel1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  group_ids = [genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResourceLabel, groupResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids.0", "genesyscloud_group."+groupResourceLabel1, "id"),
				),
			},
			// 2 groups
			{
				Config: generateUserWithCustomAttrs(testUserResourceLabel, testUserEmail, testUserName) + generateBasicGroupResource(
					groupResourceLabel1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + generateBasicGroupResource(
					groupResourceLabel2,
					groupName2,
					generateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  group_ids = [genesyscloud_group.%s.id, genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResourceLabel, groupResourceLabel1, groupResourceLabel2),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids",
						"genesyscloud_group."+groupResourceLabel1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids",
						"genesyscloud_group."+groupResourceLabel2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResourceLabel, "group_ids.#", "2"),
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds for resources to get deleted properly
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_orgauthorization_pairing." + orgAuthorizationPairingResourceLabel,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func generateBasicGroupResource(resourceLabel string, name string, nestedBlocks ...string) string {
	return generateGroupResource(resourceLabel, name, util.NullValue, util.NullValue, util.NullValue, util.TrueValue, nestedBlocks...)
}

func generateGroupResource(
	resourceLabel string,
	name string,
	desc string,
	groupType string,
	visibility string,
	rulesVisible string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_group" "%s" {
		name = "%s"
		description = %s
		type = %s
		visibility = %s
		rules_visible = %s
        %s
	}
	`, resourceLabel, name, desc, groupType, visibility, rulesVisible, strings.Join(nestedBlocks, "\n"))
}

func generateGroupOwners(userIDs ...string) string {
	return fmt.Sprintf(`owner_ids = [%s]
	`, strings.Join(userIDs, ","))
}

// TODO: Duplicating this code within the function to not break a cyclic dependency
func generateUserWithCustomAttrs(resourceLabel string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceLabel, email, name, strings.Join(attrs, "\n"))
}
