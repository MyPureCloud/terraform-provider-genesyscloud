package oauth_client

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOAuthClient(t *testing.T) {
	var (
		oauthClientDataSourceLabel = "oauth-client"

		clientResourceLabel1 = "test-client"
		clientName1          = "terraform1-" + uuid.NewString()
		clientDesc1          = "terraform test client1"
		tokenSec1            = "300"
		redirectURI1         = "https://example.com/auth1"
		grantTypeClientCreds = "CLIENT-CREDENTIALS"

		roleResourceLabel1 = "admin-role"
		roleName1          = "admin" // Must use a role already assigned to the TF OAuth client
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAuthRoleDataSource(
					roleResourceLabel1,
					strconv.Quote(roleName1),
					"",
				) + generateOauthClient(
					clientResourceLabel1,
					clientName1,
					clientDesc1,
					grantTypeClientCreds,
					tokenSec1,
					util.NullValue, // Default state
					util.GenerateStringArray(strconv.Quote(redirectURI1)),
					util.NullValue, // No scopes for client creds
					generateOauthClientRoles("data.genesyscloud_auth_role."+roleResourceLabel1+".id", util.NullValue),
				) + generateOAuthClientDataSource(
					oauthClientDataSourceLabel,
					"genesyscloud_oauth_client."+clientResourceLabel1+".name",
					"genesyscloud_oauth_client."+clientResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_oauth_client."+oauthClientDataSourceLabel, "id", "genesyscloud_oauth_client."+clientResourceLabel1, "id"),
				),
			},
		},
	})
}

func generateOAuthClientDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_oauth_client" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}

func generateAuthRoleDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_role" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
