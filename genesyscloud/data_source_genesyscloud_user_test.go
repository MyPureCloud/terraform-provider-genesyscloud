package genesyscloud

import (
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

func TestAccDataSourceUser(t *testing.T) {
	var (
		userResource   = "test-user"
		userDataSource = "test-user-data"
		randomString   = uuid.NewString()
		userEmail      = "John_Doe" + randomString + "@example.com"
		userName       = "John_Doe" + randomString
		userID         string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by email
				Config: GenerateBasicUserResource(
					userResource,
					userEmail,
					userName,
				) + generateUserDataSource(
					userDataSource,
					"genesyscloud_user."+userResource+".email",
					util.NullValue,
					"genesyscloud_user."+userResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_user."+userDataSource, "id", "genesyscloud_user."+userResource, "id"),
					resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources["genesyscloud_user."+userResource]
							if !ok {
								return fmt.Errorf("not found: %s", "genesyscloud_user."+userResource)
							}
							userID = rs.Primary.ID
							log.Printf("User ID: %s\n", userID) // Print user ID
							return nil
						},
					),
				),
			},
			{
				// Search by name
				Config: GenerateBasicUserResource(
					userResource,
					userEmail,
					userName,
				) + generateUserDataSource(
					userDataSource,
					util.NullValue,
					"genesyscloud_user."+userResource+".name",
					"genesyscloud_user."+userResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_user."+userDataSource, "id", "genesyscloud_user."+userResource, "id"),
					checkUserDeleted(userID),
				),
			},
		},
	})
}

func generateUserDataSource(
	resourceID string,
	email string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_user" "%s" {
        email = %s
		name = %s
        depends_on=[%s]
	}
	`, resourceID, email, name, dependsOnResource)
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
