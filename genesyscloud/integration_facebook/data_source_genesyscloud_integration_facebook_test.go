package integration_facebook

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the integration facebook Data Source
*/

func TestAccDataSourceIntegrationFacebook(t *testing.T) {
	t.Parallel()
	var (
		testResource1       = "test_sample"
		testResource2       = "test_sample"
		name1               = "test_sample"
		supportedContentId1 = "6b3d7fb2-c276-415c-a5c7-d18eba936c68"
		pageAccessToken1    = uuid.NewString()
		messagingSettingId1 = "2c4e3b8e-3c9f-45c9-82cd-4bb54c8f18f0"
		appId               = ""
		appSecret           = ""
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFacebookIntegrationResource(
					testResource1,
					name1,
					supportedContentId1,
					messagingSettingId1,
					pageAccessToken1,
					"",
					"",
					appId,
					appSecret,
				) + generateIntegrationFacebookDataSource(
					testResource2,
					name1,
					"genesyscloud_integration_facebook."+testResource1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_facebook."+testResource2, "id", "genesyscloud_integration_facebook."+testResource1, "id"),
				),
			},
		},
	})
}

func generateIntegrationFacebookDataSource(
	resourceId string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_facebook" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceId, name, dependsOnResource)
}
