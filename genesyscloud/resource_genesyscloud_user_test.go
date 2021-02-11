package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	userResource1 = "test-user1"
	email1        = "terraform-" + uuid.NewString() + "@example.com"
	email2        = "terraform-" + uuid.NewString() + "@example.com"
	name1         = "John Doe"
	name2         = "Jim Smith"
)

func TestAccResourceUserBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserResource(
					userResource1,
					email1,
					name1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", name1),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "password"),
					testDefaultUserDivision("genesyscloud_user."+userResource1),
				),
			},
			{
				// Update
				Config: generateUserResource(
					userResource1,
					email2,
					name2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", name2),
					testDefaultUserDivision("genesyscloud_user."+userResource1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + userResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func testVerifyUsersDestroyed(state *terraform.State) error {
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_user" {
			continue
		}
		user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "active")
		if user != nil {
			return fmt.Errorf("User (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
			// User not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All users destroyed
	return nil
}

// Default user division should be home division
func testDefaultUserDivision(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		authAPI := platformclientv2.NewAuthorizationApi()
		homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
		if err != nil {
			return fmt.Errorf("Failed to query home division: %s", err)
		}

		homeDivID := *homeDiv.Id

		r := state.RootModule().Resources[resource]
		if r == nil {
			return fmt.Errorf("User %s not found in state", resource)
		}

		a := r.Primary.Attributes

		if a["division_id"] != homeDivID {
			return fmt.Errorf("expected user's division to be home division %s", homeDivID)
		}

		return nil
	}
}

func generateUserResource(resourceID string, email string, name string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
	}`, resourceID, email, name)
}
