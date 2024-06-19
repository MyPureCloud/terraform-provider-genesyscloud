package group

import (
	"fmt"
	"log"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	sdkConfig *platformclientv2.Configuration
	mu        sync.Mutex
)

func TestAccDataSourceGroup(t *testing.T) {
	var (
		groupResource    = "test-group-members"
		groupDataSource  = "group-data"
		groupName        = "test group" + uuid.NewString()
		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@examplegroup.com"
		userID           string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) +
					GenerateGroupResource(
						groupResource,
						groupName,
						util.NullValue, // No description
						util.NullValue, // Default type
						util.NullValue, // Default visibility
						util.NullValue, // Default rules_visible
						GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
					) + generateGroupDataSource(
					groupDataSource,
					groupName,
					"genesyscloud_group."+groupResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_group."+groupDataSource, "id", "genesyscloud_group."+groupResource, "id"),
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
				ResourceName:      "genesyscloud_user." + testUserResource,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					checkUserDeleted(userID),
				),
			},
		},
	})
}

func generateGroupDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_group" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}

func checkUserDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching user with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 18
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

	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)
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
