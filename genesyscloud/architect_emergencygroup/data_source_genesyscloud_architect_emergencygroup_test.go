package architect_emergencygroup

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectEmergencyGroup(t *testing.T) {
	var (
		emergencyGroupResourceLabel   = "e-group-1"
		emergencyGroupDataSourceLabel = "e-group-data"
		name                          = "CX as Code Emergency Group" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateArchitectEmergencyGroupResource(emergencyGroupResourceLabel,
					name,
					util.NullValue,
					"",
					util.FalseValue,
					"",
				) + generateEmergencyGroupDataSource(
					emergencyGroupDataSourceLabel,
					name,
					"genesyscloud_architect_emergencygroup."+emergencyGroupResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_emergencygroup."+emergencyGroupDataSourceLabel, "id",
						"genesyscloud_architect_emergencygroup."+emergencyGroupResourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyEmergencyGroupDestroyed,
	})
}

func generateEmergencyGroupDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_emergencygroup" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
