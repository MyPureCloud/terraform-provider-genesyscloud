package workforcemanagement_business_unit_activitycode

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccDataSourceBusinessUnitActivityCode(t *testing.T) {
	var (
		buResourceLabel = "bu_1"
		buName          = "TestBU" + uuid.NewString()
		resourceLabel   = "activity_code_1"
		dataSourceLabel = "activity_code_data"
		name            = "Test Activity Code " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTestBusinessUnit(buResourceLabel, buName) +
					GenerateBusinessUnitActivityCodeResource(resourceLabel, businessUnitIdRef(buResourceLabel), name, "Meeting") +
					generateBusinessUnitActivityCodeDataSource(dataSourceLabel, businessUnitIdRef(buResourceLabel), name, ResourceType+"."+resourceLabel),
				Check: resource.TestCheckResourceAttrPair(
					"data."+ResourceType+"."+dataSourceLabel, "id",
					ResourceType+"."+resourceLabel, "id",
				),
			},
		},
	})
}

func generateBusinessUnitActivityCodeDataSource(dataSourceLabel string, businessUnitId string, name string, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
	business_unit_id = %s
	name             = "%s"
	depends_on       = [%s]
}
`, ResourceType, dataSourceLabel, businessUnitId, name, dependsOn)
}
