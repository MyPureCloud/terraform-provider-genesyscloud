package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWebDeploymentsDeployment(t *testing.T) {
	var (
		deploymentName        = "BasicDeployment" + randString(8)
		deploymentDescription = "Basic Deployment description"
		fullResourceName      = "genesyscloud_webdeployments_deployment.basic"
		fullDataSourceName    = "data.genesyscloud_webdeployments_deployment.basic-data"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: basicDeploymentDataSource(deploymentName, deploymentDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(fullResourceName, "id", fullDataSourceName, "id"),
				),
			},
		},
	})
}

func basicDeploymentDataSource(deploymentName string, deploymentDescr string) string {
	minimalConfigName := "Minimal Config " + uuid.NewString()
	return fmt.Sprintf(`

	resource "genesyscloud_webdeployments_configuration" "minimal" {
		name             = "%s"
		languages        = ["en-us"]
		default_language = "en-us"
	}

	resource "genesyscloud_webdeployments_deployment" "basic" {
		name = "%s"
		description = "%s"
		allow_all_domains = true
		configuration {
			id = "${genesyscloud_webdeployments_configuration.minimal.id}"
			version = "${genesyscloud_webdeployments_configuration.minimal.version}"
		}
	}
	
	data "genesyscloud_webdeployments_deployment" "basic-data" {
		depends_on=[genesyscloud_webdeployments_deployment.basic]
		name = "%s"
	}
	`, minimalConfigName, deploymentName, deploymentDescr, deploymentName)
}
