package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEmergencyGroup(t *testing.T) {
	var (
		emergencyGroupResourceID   = "e-group-1"
		emergencyGroupDataSourceID = "e-group-data"
		name                       = "CX as Code Emergency Group" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateArchitectEmergencyGroupResource(emergencyGroupResourceID,
					name,
					nullValue,
					"",
					falseValue,
					"",
				) + generateEmergencyGroupDataSource(
					emergencyGroupDataSourceID,
					name,
					"genesyscloud_architect_emergencygroup."+emergencyGroupResourceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_emergencygroup."+emergencyGroupDataSourceID, "id",
						"genesyscloud_architect_emergencygroup."+emergencyGroupResourceID, "id"),
				),
			},
		},
	})
}

func generateEmergencyGroupDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_emergencygroup" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
