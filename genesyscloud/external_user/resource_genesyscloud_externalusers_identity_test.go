package external_user

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	userResource "terraform-provider-genesyscloud/genesyscloud/user"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalUser(t *testing.T) {
	var (
		userName         = "TestUser" + uuid.NewString()
		userEmail        = uuid.NewString() + "@website.com"
		externalKey      = "microsoftlogin"
		authorityName    = "msft"
		userResoureLabel = "sample_user"
		resourceLabel    = "sample_external_user"
		resourcePath     = ResourceType + "." + resourceLabel
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: userResource.GenerateBasicUserResource(
					userResoureLabel,
					userEmail,
					userName,
				) + generateExternalUserIdentity(resourceLabel, "genesyscloud_user."+userResoureLabel+".id", authorityName, externalKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "external_key", externalKey),
					resource.TestCheckResourceAttr(resourcePath, "authority_name", authorityName),
				),
			},
		},
	},
	)
}

func generateExternalUserIdentity(resourceLabel, userId, authorityName, externalKey string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalusers_identity" "%s" {
        user_id = %s
        authority_name = "%s"
        external_key = "%s"
	}
	`, resourceLabel, userId, authorityName, externalKey)
}
