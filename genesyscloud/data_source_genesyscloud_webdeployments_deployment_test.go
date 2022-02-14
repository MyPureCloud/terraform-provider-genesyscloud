package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWebDeploymentsDeployment(t *testing.T) {
	var (
		deploymentName        = "Basic Deployment " + randString(8)
		deploymentDescription = "Basic Deployment description"
		fullResourceName      = "genesyscloud_webdeployments_deployment.basic"
		fullDataSourceName    = "data.genesyscloud_webdeployments_deployment.basic-data"
		resourceNameReference = fullResourceName + ".name"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: basicDeploymentResource(deploymentName, deploymentDescription) +
					basicDeploymentDataSource(resourceNameReference),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(fullResourceName, "id", fullDataSourceName, "id"),
				),
			},
		},
	})
}

func basicDeploymentDataSource(name string) string {
	return fmt.Sprintf(`
	data "genesyscloud_webdeployments_deployment" "basic-data" {
		name = %s
	}
	`, name)
}
