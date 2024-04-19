package responsemanagement_responseasset

import (
	"fmt"
	"log"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceResponseManagementResponseAsset(t *testing.T) {
	var (
		resourceId   = "resp_asset"
		testDirName  = "test_responseasset_data"
		fileName     = filepath.Join(testDirName, "yeti-img.png")
		dataSourceId = "resp_asset_data"
	)

	defer func() {
		err := cleanupResponseAssets(testDirName)
		if err != nil {
			log.Printf("error cleaning up response assets: %v. Dangling assets may exist.", err)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateResponseManagementResponseAssetDataSource(dataSourceId, fileName, "genesyscloud_responsemanagement_responseasset."+resourceId) +
					GenerateResponseManagementResponseAssetResource(resourceId, fileName, util.NullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_responsemanagement_responseasset."+dataSourceId, "id",
						"genesyscloud_responsemanagement_responseasset."+resourceId, "id"),
				),
			},
		},
	})
}

func generateResponseManagementResponseAssetDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_responsemanagement_responseasset" "%s" {
    name       = "%s"
    depends_on = [%s]
}
`, id, name, dependsOn)
}
