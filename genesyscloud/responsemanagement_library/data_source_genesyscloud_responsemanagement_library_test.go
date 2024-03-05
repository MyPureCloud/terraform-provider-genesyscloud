package responsemanagement_library

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceResponseManagementLibrary(t *testing.T) {
	var (
		resourceId   = "library"
		dataSourceId = "library_data"
		name         = "Library " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: GenerateResponseManagementLibraryResource(
					resourceId,
					name,
				) + generateResponseManagementLibraryDataSource(
					dataSourceId,
					name,
					"genesyscloud_responsemanagement_library."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_responsemanagement_library."+dataSourceId, "id",
						"genesyscloud_responsemanagement_library."+resourceId, "id",
					),
				),
			},
		},
	})
}

func generateResponseManagementLibraryDataSource(
	dataSourceId string,
	name string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_responsemanagement_library" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, dataSourceId, name, dependsOn)
}
