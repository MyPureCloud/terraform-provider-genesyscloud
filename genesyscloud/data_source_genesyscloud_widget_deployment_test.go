package genesyscloud

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWidgetDeployment(t *testing.T) {
	var (
		widgegetDeploymentsResource = "widget-deployments"
		widgetDeploymentsDataSource = "widget-deployments-data"
		widgetDeploymentsName       = "Widget_deployments-"
	)
	description := "This is a test description"
	flowId := uuid.NewString()
	widgetDeployV1 := &widgetDeploymentConfig{
		resourceID:             widgegetDeploymentsResource,
		name:                   widgetDeploymentsName,
		description:            strconv.Quote(description),
		flowID:                 strconv.Quote(flowId),
		clientType:             V2,
		authenticationRequired: "true",
		disabled:               "true",
	}

	deleteWidgetDeploymentWithName(widgetDeploymentsName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateWidgetDeploymentResource(widgetDeployV1) + generateWidgetDeploymentDataSource(widgetDeploymentsDataSource, "genesyscloud_widget_deployment."+widgegetDeploymentsResource+".name", "genesyscloud_widget_deployment."+widgegetDeploymentsResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_widget_deployment."+widgetDeploymentsDataSource, "id", "genesyscloud_widget_deployment."+widgegetDeploymentsResource, "id"),
				),
			},
		},
	})
}

func generateWidgetDeploymentDataSource(
	resourceID string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_widget_deployment" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
