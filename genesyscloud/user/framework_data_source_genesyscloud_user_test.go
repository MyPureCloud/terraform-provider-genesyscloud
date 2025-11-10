package user

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure test resources are initialized for Framework tests
func init() {
	if frameworkResources == nil || frameworkDataSources == nil {
		initTestResources()
	}
}

func TestAccFrameworkDataSourceUser(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel   = "test-user-resource"
		userDataSourceLabel = "test-user-data-source"
		randomString        = uuid.NewString()
		userEmail           = "framework_user_" + randomString + "@example.com"
		userName            = "Framework_User_" + randomString
		userID              string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
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
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[ResourceType+"."+userResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", ResourceType+"."+userResourceLabel)
						}
						// Verify user ID is set
						if rs.Primary.ID == "" {
							return fmt.Errorf("user ID is empty")
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
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for proper cleanup
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
