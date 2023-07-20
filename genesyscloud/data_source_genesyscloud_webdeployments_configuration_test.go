package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWebDeploymentsConfiguration(t *testing.T) {
	var (
		configurationName        = "Basic Configuration " + randString(8)
		configurationDescription = "Basic config description"
		fullResourceName         = "genesyscloud_webdeployments_configuration.basic"
		fullDataSourceName       = "data.genesyscloud_webdeployments_configuration.basic-data"
		resourceNameReference    = fullResourceName + ".name"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: basicConfigurationResource(configurationName, configurationDescription) +
					basicConfigurationDataSource(resourceNameReference),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(fullResourceName, "id", fullDataSourceName, "id"),
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
