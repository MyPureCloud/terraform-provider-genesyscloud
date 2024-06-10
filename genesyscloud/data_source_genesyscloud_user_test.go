package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	var (
		userResource   = "test-user"
		userDataSource = "test-user-data"
		randomString   = uuid.NewString()
		userEmail      = "John_Doe" + randomString + "@example.com"
		userName       = "John_Doe" + randomString
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
