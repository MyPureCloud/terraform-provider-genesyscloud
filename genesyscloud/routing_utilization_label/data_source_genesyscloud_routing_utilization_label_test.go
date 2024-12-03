package routing_utilization_label

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccDataSourceRoutingUtilizationLabel(t *testing.T) {
	var (
		resourceLabel   = "test-label"
		dataSourceLabel = "data-source-label"
		labelName       = "Terraform Label " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateRoutingUtilizationLabelResource(
					resourceLabel,
					labelName,
					"",
				) + generateRoutingUtilizationLabelDataSource(dataSourceLabel, labelName, "genesyscloud_routing_utilization_label."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_utilization_label."+dataSourceLabel, "id", "genesyscloud_routing_utilization_label."+resourceLabel, "id"),
				),
			},
		},
		CheckDestroy: validateTestLabelDestroyed,
	})
}

func generateRoutingUtilizationLabelDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
        depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}

func checkIfLabelsAreEnabled() error { // remove once the feature is globally enabled
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	_, resp, _ := api.GetRoutingUtilizationLabels(100, 1, "", "")
	if resp.StatusCode == 501 {
		return fmt.Errorf("feature is not yet implemented in this org.")
	}
	return nil
}

func generateRoutingUtilizationLabelResource(resourceLabel string, name string, dependsOnResource string) string {
	dependsOn := ""

	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on=[genesyscloud_routing_utilization_label.%s]", dependsOnResource)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
		%s
	}
	`, resourceLabel, name, dependsOn)
}
