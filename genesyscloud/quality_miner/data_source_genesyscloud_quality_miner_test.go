package quality_miner

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceQualityMiner(t *testing.T) {
	var (
		minerResourceLabel = "miner"
		dsWithTypeLabel    = "miner-data-withtype"
		dsNoTypeLabel      = "miner-data-notype"
		minerName          = "tf-miner-" + uuid.NewString()[:8]
		minerType          = "Topic"
		resourcePath       = "genesyscloud_quality_miner." + minerResourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateQualityMinerResource(
					minerResourceLabel,
					minerName,
					minerType,
				),
			},
			{
				Config: GenerateQualityMinerResource(
					minerResourceLabel,
					minerName,
					minerType,
				) + generateQualityMinerDataSourceWithType(
					dsWithTypeLabel,
					resourcePath+".name",
					resourcePath+".miner_type",
					resourcePath,
				) + generateQualityMinerDataSourceNoType(
					dsNoTypeLabel,
					resourcePath+".name",
					resourcePath,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_quality_miner."+dsWithTypeLabel, "id", resourcePath, "id"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_quality_miner."+dsNoTypeLabel, "id", resourcePath, "id"),
				),
			},
		},
	})
}

func TestAccDataSourceQualityMinerNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: `data "genesyscloud_quality_miner" "ghost" {
	name = "ghost-` + uuid.NewString()[:8] + `"
}
`,
				ExpectError: regexp.MustCompile("no miner found|unable to find miner"),
			},
		},
	})
}

func generateQualityMinerDataSourceWithType(
	resourceLabel string,
	name string,
	minerType string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_quality_miner" "%s" {
	name       = %s
	miner_type = %s
	depends_on = [%s]
}
`, resourceLabel, name, minerType, dependsOnResource)
}

func generateQualityMinerDataSourceNoType(
	resourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_quality_miner" "%s" {
	name       = %s
	depends_on = [%s]
}
`, resourceLabel, name, dependsOnResource)
}
