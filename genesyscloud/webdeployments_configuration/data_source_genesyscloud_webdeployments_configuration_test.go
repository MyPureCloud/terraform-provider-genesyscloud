package webdeployments_configuration

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

func TestAccDataSourceWebDeploymentsConfiguration(t *testing.T) {
	var (
		configurationName        = "Basic Configuration " + gcloud.RandString(8)
		configurationDescription = "Basic config description"
		fullResourceName         = "genesyscloud_webdeployments_configuration.basic"
		fullDataSourceName       = "data.genesyscloud_webdeployments_configuration.basic-data"
		resourceNameReference    = fullResourceName + ".name"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
