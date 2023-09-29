package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceOrgAuthorizationPairing(t *testing.T) {
	t.Parallel()
	var (
		orgAuthorizationPairingResource = "test-orgauthorization-pairing"
		userResource1                   = "test-user-1"
		userResource2                   = "test-user-2"
		email1                          = "terraform-1-" + uuid.NewString() + "@example.com"
		email2                          = "terraform-2-" + uuid.NewString() + "@example.com"
		userName1                       = "test user " + uuid.NewString()
		userName2                       = "test user " + uuid.NewString()
		groupResource1                  = "test-group-1"
		groupResource2                  = "test-group-2"
		groupName1                      = "TF Group" + uuid.NewString()
		groupName2                      = "TF Group" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// 1 user and 1 group
			{
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + GenerateBasicGroupResource(
					groupResource1,
					groupName1,
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
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + GenerateBasicUserResource(
					userResource2,
					email2,
					userName2,
				) + GenerateBasicGroupResource(
					groupResource1,
					groupName1,
				) + GenerateBasicGroupResource(
					groupResource2,
					groupName2,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id, genesyscloud_user.%s.id]
  group_ids = [genesyscloud_group.%s.id, genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, userResource1, userResource2, groupResource1, groupResource2),
				Check: resource.ComposeTestCheckFunc(
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource1, "id"),
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource2, "id"),
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource1, "id"),
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.#", "2"),
				),
			},
			// 1 user
			{
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id]
}`, orgAuthorizationPairingResource, userResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.0", "genesyscloud_user."+userResource1, "id"),
				),
			},
			// 2 users
			{
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				) + GenerateBasicUserResource(
					userResource2,
					email2,
					userName2,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  user_ids  = [genesyscloud_user.%s.id, genesyscloud_user.%s.id]
}`, orgAuthorizationPairingResource, userResource1, userResource2),
				Check: resource.ComposeTestCheckFunc(
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource1, "id"),
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids",
						"genesyscloud_user."+userResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "user_ids.#", "2"),
				),
			},
			// 1 group
			{
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName1,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  group_ids = [genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, groupResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.0", "genesyscloud_group."+groupResource1, "id"),
				),
			},
			// 2 groups
			{
				Config: GenerateBasicGroupResource(
					groupResource1,
					groupName1,
				) + GenerateBasicGroupResource(
					groupResource2,
					groupName2,
				) + fmt.Sprintf(`resource "genesyscloud_orgauthorization_pairing" "%s" {
  group_ids = [genesyscloud_group.%s.id, genesyscloud_group.%s.id]
}`, orgAuthorizationPairingResource, groupResource1, groupResource2),
				Check: resource.ComposeTestCheckFunc(
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource1, "id"),
					validateResourceAttributeInArray("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids",
						"genesyscloud_group."+groupResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_orgauthorization_pairing."+orgAuthorizationPairingResource, "group_ids.#", "2"),
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
