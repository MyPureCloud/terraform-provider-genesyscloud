package journey_views

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceJourneyViewBasic(t *testing.T) {
	var (
		journeyResourceLabel   = "test-journey"
		journeyName            = "TerraformTestJourney-" + uuid.NewString()
		duration               = "P1Y"
		elementsBlock          = ""
		chartsBlock            = ""
		journeyDataSourceLabel = "test-journey-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateJourneyView(
					journeyResourceLabel,
					journeyName,
					duration,
					elementsBlock,
					chartsBlock,
				) + generateJourneyViewDataSource(
					journeyDataSourceLabel,
					journeyName,
					"genesyscloud_journey_views."+journeyResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_journey_views."+journeyDataSourceLabel,
						"id", "genesyscloud_journey_views."+journeyResourceLabel, "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceJourneyViewMultiple(t *testing.T) {
	var (
		journeyResourceLabel1 = "journey1"
		journeyName1          = "terraform test journey " + uuid.NewString()
		journeyResourceLabel2 = "journey2"
		journeyName2          = "terraform test journey " + uuid.NewString()
		journeyResourceLabel3 = "journey3"
		journeyName3          = "terraform test journey " + uuid.NewString()
		duration              = "P1Y"
		elementsBlock         = ""
		dataSourceLabel1      = "data-1"
		dataSourceLabel2      = "data-2"
		dataSourceLabel3      = "data-3"
		chartsBlock           = ""
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(45 * time.Second)
				},
				Config: generateJourneyView( // journey resource
					journeyResourceLabel1,
					journeyName1,
					duration,
					elementsBlock,
					chartsBlock,
				) + generateJourneyView( // journey resource
					journeyResourceLabel2,
					journeyName2,
					duration,
					elementsBlock,
					chartsBlock,
				) + generateJourneyView( // journey resource
					journeyResourceLabel3,
					journeyName3,
					duration,
					elementsBlock,
					chartsBlock,
				) + generateJourneyViewDataSource( // journey data source
					dataSourceLabel1,
					journeyName1,
					"genesyscloud_journey_views."+journeyResourceLabel1,
				) + generateJourneyViewDataSource( // journey data source
					dataSourceLabel2,
					journeyName2,
					"genesyscloud_journey_views."+journeyResourceLabel2,
				) + generateJourneyViewDataSource( // journey data source
					dataSourceLabel3,
					journeyName3,
					"genesyscloud_journey_views."+journeyResourceLabel3,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_journey_views."+journeyResourceLabel1, "id",
						"data.genesyscloud_journey_views."+dataSourceLabel1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_journey_views."+journeyResourceLabel2, "id",
						"data.genesyscloud_journey_views."+dataSourceLabel2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_journey_views."+journeyResourceLabel3, "id",
						"data.genesyscloud_journey_views."+dataSourceLabel3, "id"),
				),
			},
		},
		CheckDestroy: testVerifyJourneyViewsDestroyed,
	})
}

func generateJourneyViewDataSource(
	dataSourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_journey_views" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, dataSourceLabel, name, dependsOnResource)
}
