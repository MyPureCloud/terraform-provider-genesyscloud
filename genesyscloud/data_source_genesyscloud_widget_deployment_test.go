package genesyscloud

import (
	"fmt"
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

	widgetDeployV1 := &widgetDeploymentConfig{
		resourceID:             widgegetDeploymentsResource,
		name:                   widgetDeploymentsName + uuid.NewString(),
		description:            "This is a test description",
		flowID:                 uuid.NewString(),
		clientType:             "v1",
		authenticationRequired: "true",
		disabled:               "true",
		webChatSkin:            "basic",
		authenticationUrl:      "https://localhost",
	}

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteWidgetDeploymentWithName(widgetDeploymentsName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateWidgetDeployV1(widgetDeployV1) + generateWidgetDeploymentDataSource(widgetDeploymentsDataSource, "genesyscloud_widget_deployment."+widgegetDeploymentsResource+".name", "genesyscloud_widget_deployment."+widgegetDeploymentsResource),
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
