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
		resourceLabel   = "library"
		dataSourceLabel = "library_data"
		name            = "Library " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: GenerateResponseManagementLibraryResource(
					resourceLabel,
					name,
				) + generateResponseManagementLibraryDataSource(
					dataSourceLabel,
					name,
					"genesyscloud_responsemanagement_library."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_responsemanagement_library."+dataSourceLabel, "id",
						"genesyscloud_responsemanagement_library."+resourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateResponseManagementLibraryDataSource(
	dataSourceLabel string,
	name string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_responsemanagement_library" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, dataSourceLabel, name, dependsOn)
}
