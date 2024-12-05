package user

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
)

func TestAccDataSourceUser(t *testing.T) {
	var (
		userResourceLabel   = "test-user"
		userDataSourceLabel = "test-user-data"
		randomString        = uuid.NewString()
		userEmail           = "John_Doe" + randomString + "@exampleuser.com"
		userName            = "John_Doe" + randomString
		userID              string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by email
				Config: GenerateBasicUserResource(
					userResourceLabel,
					userEmail,
					userName,
				) + generateUserDataSource(
					userDataSourceLabel,
					ResourceType+"."+userResourceLabel+".email",
					util.NullValue,
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[ResourceType+"."+userResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", ResourceType+"."+userResourceLabel)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),
			},
			{
				// Search by name
				Config: GenerateBasicUserResource(
					userResourceLabel,
					userEmail,
					userName,
				) + generateUserDataSource(
					userDataSourceLabel,
					util.NullValue,
					ResourceType+"."+userResourceLabel+".name",
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyUsersDestroyed(state)
		},
	})
}

func generateUserDataSource(
	resourceLabel string,
	email string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
        email = %s
		name = %s
        depends_on=[%s]
	}
	`, ResourceType, resourceLabel, email, name, dependsOnResource)
}
