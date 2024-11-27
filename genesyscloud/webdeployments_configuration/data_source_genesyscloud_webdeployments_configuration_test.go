package webdeployments_configuration

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWebDeploymentsConfiguration(t *testing.T) {
	var (
		configurationName        = "Basic Configuration " + util.RandString(8)
		configurationDescription = "Basic config description"
		resourcePath             = "genesyscloud_webdeployments_configuration.basic"
		dataPath                 = "data.genesyscloud_webdeployments_configuration.basic-data"
		nameAttrReference        = resourcePath + ".name"
	)

	cleanupWebDeploymentsConfiguration(t, "Test Configuration ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: basicConfigurationResource(configurationName, configurationDescription) +
					basicConfigurationDataSource(nameAttrReference),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourcePath, "id", dataPath, "id"),
				),
			},
		},
	})
}

func basicConfigurationDataSource(name string) string {
	return fmt.Sprintf(`
	data "genesyscloud_webdeployments_configuration" "basic-data" {
		name = %s
	}
	`, name)
}
