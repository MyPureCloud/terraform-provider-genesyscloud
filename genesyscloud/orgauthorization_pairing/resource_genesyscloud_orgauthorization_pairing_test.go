package orgauthorization_pairing

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOrgAuthorizationPairing(t *testing.T) {
	var (
		orgAuthorizationPairingResource = "test-orgauthorization-pairing"
		userResource1                   = "test-user-1"
		userResource2                   = "test-user-2"
		email1                          = "terraform-1-" + uuid.NewString() + "@authpair.com"
		email2                          = "terraform-2-" + uuid.NewString() + "@authpair.com"
		userName1                       = "test user " + uuid.NewString()
		userName2                       = "test user " + uuid.NewString()
		groupResource1                  = "test-group-1"
		groupResource2                  = "test-group-2"
		groupName1                      = "TF Group" + uuid.NewString()
		groupName2                      = "TF Group" + uuid.NewString()
		testUserResource                = "user_resource1"
		testUserName                    = "nameUser1" + uuid.NewString()
		testUserEmail                   = uuid.NewString() + "@authpair.com"
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
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + generateBasicGroupResource(
					groupResource1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id]
  group_ids = [genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, userResource1, groupResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.0", "genesyscloud_user."+userResource1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.0", "genesyscloud_group."+groupResource1, "id"),
				),
			},
			// 2 users and 2 groups
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + genesyscloud.GenerateBasicUserResource(
					userResource2,
					email2,
					userName2,
				) + generateBasicGroupResource(
					groupResource1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + generateBasicGroupResource(
					groupResource2,
					groupName2,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id, genesyscloud_user.%s.id]
  group_ids = [genesyscloud_group.%s.id, genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, userResource1, userResource2, groupResource1, groupResource2),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource2, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.#", "2"),
				),
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
			},
			// 1 user
			{
				Config: genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id]
}`, orgAuthorizationPairingResource, userResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.0", "genesyscloud_user."+userResource1, "id"),
				),
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
			},
			// 2 users
			{
				Config: genesyscloud.GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + genesyscloud.GenerateBasicUserResource(
					userResource2,
					email2,
					userName2,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id, genesyscloud_user.%s.id]
}`, orgAuthorizationPairingResource, userResource1, userResource2),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper creation
						return nil
					},
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.#", "2"),
				),
			},
			// 1 group
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + generateBasicGroupResource(
					groupResource1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  group_ids = [genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, groupResource1),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.0", "genesyscloud_group."+groupResource1, "id"),
				),
			},
			// 2 groups
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + generateBasicGroupResource(
					groupResource1,
					groupName1,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + generateBasicGroupResource(
					groupResource2,
					groupName2,
					generateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  group_ids = [genesyscloud_group.%s.id, genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, groupResource1, groupResource2),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource1, "id"),
					util.ValidateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.#", "2"),
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds for resources to get deleted properly
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_orgauthorization_pairing." + orgAuthorizationPairingResource,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func generateBasicGroupResource(resourceID string, name string, nestedBlocks ...string) string {
	return generateGroupResource(resourceID, name, util.NullValue, util.NullValue, util.NullValue, util.TrueValue, nestedBlocks...)
}

func generateGroupResource(
	resourceID string,
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
	`, resourceID, name, desc, groupType, visibility, rulesVisible, strings.Join(nestedBlocks, "\n"))
}

func generateGroupOwners(userIDs ...string) string {
	return fmt.Sprintf(`owner_ids = [%s]
	`, strings.Join(userIDs, ","))
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
