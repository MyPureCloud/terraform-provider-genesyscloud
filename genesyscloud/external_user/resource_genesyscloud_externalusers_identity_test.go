package external_user

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	userResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceExternalUser(t *testing.T) {
	var (
		randomizer         = uuid.NewString()
		userName           = "TestUser" + randomizer
		userEmail          = randomizer + "@website.com"
		externalKey        = randomizer
		UpdatedExternalKey = "updated" + randomizer
		authorityName      = "msft"
		userResoureLabel   = "sample_user"
		resourceLabel      = "sample_external_user"
		resourcePath       = ResourceType + "." + resourceLabel
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user first
				Config: userResource.GenerateBasicUserResource(
					userResoureLabel,
					userEmail,
					userName,
				),
			},
			{
				// Give the user some time to register in the API
				PreConfig: func() {
					time.Sleep(5 * time.Second)
				},
				// Create external user identity
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
			{
				// Update

				Config: userResource.GenerateBasicUserResource(
					userResoureLabel,
					userEmail,
					userName,
				) + generateExternalUserIdentity(resourceLabel, "genesyscloud_user."+userResoureLabel+".id", authorityName, UpdatedExternalKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "external_key", UpdatedExternalKey),
					resource.TestCheckResourceAttr(resourcePath, "authority_name", authorityName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_externalusers_identity." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	},
	)
}
