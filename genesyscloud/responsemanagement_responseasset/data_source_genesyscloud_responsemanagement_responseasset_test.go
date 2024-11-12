package responsemanagement_responseasset

import (
	"fmt"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceResponseManagementResponseAsset(t *testing.T) {
	var (
		resourceLabel   = "resp_asset"
		testDirName     = "test_responseasset_data"
		fileName        = filepath.Join(testDirName, "yeti-img-asset.png")
		dataSourceLabel = "resp_asset_data"
	)
	cleanupResponseAssets("yeti")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fileName, util.NullValue) +
					generateResponseManagementResponseAssetDataSource(dataSourceLabel, fileName, "genesyscloud_responsemanagement_responseasset."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_responsemanagement_responseasset."+dataSourceLabel, "id",
						"genesyscloud_responsemanagement_responseasset."+resourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyResponseAssetDestroyed,
	})
}

func generateResponseManagementResponseAssetDataSource(dataSourceLabel string, name string, dependsOn string) string {
	normalizeFilePath, _ := testrunner.NormalizeFileName(name)
	return fmt.Sprintf(`
data "genesyscloud_responsemanagement_responseasset" "%s" {
    name       = "%s"
    depends_on = [%s]
}
`, dataSourceLabel, normalizeFilePath, dependsOn)
}
