package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOAuthClient(t *testing.T) {
	var (
		oauthClientDataSource = "oauth-client"

		clientResource1      = "test-client"
		clientName1          = "terraform1-" + uuid.NewString()
		clientDesc1          = "terraform test client1"
		tokenSec1            = "300"
		redirectURI1         = "https://example.com/auth1"
		grantTypeClientCreds = "CLIENT-CREDENTIALS"

		roleResource1 = "admin-role"
		roleName1     = "admin" // Must use a role already assigned to the TF OAuth client
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateAuthRoleDataSource(
					roleResource1,
					strconv.Quote(roleName1),
					"",
				) + generateOauthClient(
					clientResource1,
					clientName1,
					clientDesc1,
					grantTypeClientCreds,
					tokenSec1,
					nullValue, // Default state
					generateStringArray(strconv.Quote(redirectURI1)),
					nullValue, // No scopes for client creds
					generateOauthClientRoles("data.genesyscloud_auth_role."+roleResource1+".id", nullValue),
				) + generateOAuthClientDataSource(
					oauthClientDataSource,
					"genesyscloud_oauth_client."+clientResource1+".name",
					"genesyscloud_oauth_client."+clientResource1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_oauth_client."+oauthClientDataSource, "id", "genesyscloud_oauth_client."+clientResource1, "id"),
				),
			},
		},
	})
}

func generateOAuthClientDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_oauth_client" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
