package guide

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccDataSourceGuide(t *testing.T) {
	var (
		guideResourceLabel   = "test-guide"
		guideDataSourceLabel = "guide-data"
		guideName            = "Test Guide " + uuid.NewString()
		guideSource          = "Manual"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create guide and test data source
				Config: GenerateGuideResource(
					guideResourceLabel,
					guideName,
					guideSource,
				) + generateGuideDataSource(
					guideDataSourceLabel,
					guideName,
					"genesyscloud_guide."+guideResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_guide."+guideDataSourceLabel, "id",
						"genesyscloud_guide."+guideResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("data.genesyscloud_guide."+guideDataSourceLabel, "name", guideName),
				),
			},
		},
	})
}

func generateGuideDataSource(resourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_guide" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
