package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	var (
		userResource   = "test-user"
		userDataSource = "test-user-data"
		userEmail      = "terraform-" + uuid.NewString() + "@example.com"
		userName       = "John Data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Search by email
				Config: generateBasicUserResource(
					userResource,
					userEmail,
					userName,
				) + generateUserDataSource(userDataSource, "genesyscloud_user."+userResource+".email", nullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_user."+userDataSource, "id", "genesyscloud_user."+userResource, "id"),
				),
			},
			{
				// Search by name
				Config: generateBasicUserResource(
					userResource,
					userEmail,
					userName,
				) + generateUserDataSource(userDataSource, nullValue, "genesyscloud_user."+userResource+".name"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_user."+userDataSource, "id", "genesyscloud_user."+userResource, "id"),
				),
			},
		},
	})
}

func generateUserDataSource(
	resourceID string,
	email string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_user" "%s" {
        email = %s
		name = %s
	}
	`, resourceID, email, name)
}
