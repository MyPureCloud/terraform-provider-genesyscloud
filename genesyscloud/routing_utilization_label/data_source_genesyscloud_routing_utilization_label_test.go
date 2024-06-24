package routing_utilization_label

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingUtilizationLabel(t *testing.T) {
	var (
		resourceName   = "test-label"
		dataSourceName = "data-source-label"
		labelName      = "Terraform Label " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
			if err := CheckIfLabelsAreEnabled(); err != nil {
				t.Skipf("%v", err) // be sure to skip the test and not fail it
			}
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingUtilizationLabelResource(
					resourceName,
					labelName,
					"",
				) + generateRoutingUtilizationLabelDataSource(dataSourceName, labelName, "genesyscloud_routing_utilization_label."+resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_utilization_label."+dataSourceName, "id", "genesyscloud_routing_utilization_label."+resourceName, "id"),
				),
			},
		},
		CheckDestroy: validateTestLabelDestroyed,
	})
}

func generateRoutingUtilizationLabelDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
