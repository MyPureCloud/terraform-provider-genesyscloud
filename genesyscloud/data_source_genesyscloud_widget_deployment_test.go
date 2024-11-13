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
		widgegetDeploymentsResourceLabel = "widget-deployments"
		widgetDeploymentsDataSourceLabel = "widget-deployments-data"
		widgetDeploymentsName            = "Widget_deployments-"
	)
	description := "This is a test description"
	flowId := uuid.NewString()
	widgetDeployV1 := &widgetDeploymentConfig{
		resourceLabel:          widgegetDeploymentsResourceLabel,
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
				Config: generateWidgetDeploymentResource(widgetDeployV1) + generateWidgetDeploymentDataSource(widgetDeploymentsDataSourceLabel, "genesyscloud_widget_deployment."+widgegetDeploymentsResourceLabel+".name", "genesyscloud_widget_deployment."+widgegetDeploymentsResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_widget_deployment."+widgetDeploymentsDataSourceLabel, "id", "genesyscloud_widget_deployment."+widgegetDeploymentsResourceLabel, "id"),
				),
			},
		},
	})
}

func generateWidgetDeploymentDataSource(
	resourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_widget_deployment" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
