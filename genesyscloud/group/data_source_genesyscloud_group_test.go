package group

import (
	"context"
	"fmt"
	"log"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	mu sync.Mutex
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
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
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
			return testVerifyUsersDestroyed(state)
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
	return func(s *terraform.State) error {
		maxAttempts := 30
		fmt.Printf("Fetching user with ID: %s\n", id)
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
	fmt.Printf("User ID: %s\n", id)
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

func testVerifyUsersDestroyed(state *terraform.State) error {
	usersAPI := platformclientv2.NewUsersApi()

	diagErr := util.WithRetries(context.Background(), 20*time.Second, func() *retry.RetryError {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "genesyscloud_user" {
				continue
			}
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			_, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")

			if err != nil {
				if util.IsStatus404(resp) {
					continue
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_user", fmt.Sprintf("Unexpected error: %s", err), resp))
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_user", fmt.Sprintf("User (%s) still exists", rs.Primary.ID), resp))
		}
		return nil
	})

	if diagErr != nil {
		return fmt.Errorf(fmt.Sprintf("%v", diagErr))
	}

	// Success. All users destroyed
	return nil
}
