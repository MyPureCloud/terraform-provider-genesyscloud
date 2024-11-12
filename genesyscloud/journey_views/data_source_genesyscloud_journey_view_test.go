package journey_views

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceJourneyViewBasic(t *testing.T) { //LLG TODO
	var (
		journeyResource = "test-journey"
		journeyName     = "Terraform Test Journey-" + uuid.NewString()
		journeyDesc     = "This is a test"

		journeyDataSource = "test-journey-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateJourneyViewResource(
					journeyResource,
					journeyName,
					journeyDesc,
				) + generateJourneyViewDataSource(
					journeyDataSource,
					"genesyscloud_routing_queue."+journeyResource+".name",
					"genesyscloud_routing_queue."+journeyResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_jjjoruney_view."+journeyDataSource,
						"id", "genesyscloud_journey_view."+journeyResource, "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceJourneyViewCaching(t *testing.T) {
	var (
		journey1ResourceId = "journey1"
		journeyName1       = "terraform test journey " + uuid.NewString()
		journey2ResourceId = "journey2"
		journeyName2       = "terraform test journey " + uuid.NewString()
		journey3ResourceId = "journey3"
		journeyName3       = "terraform test journey " + uuid.NewString()
		duration           = "1"
		elementsBlock      = ""
		dataSource1Id      = "data-1"
		dataSource2Id      = "data-2"
		dataSource3Id      = "data-3"
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
					journey1ResourceId,
					journeyName1,
					duration,
					elementsBlock,
				) + generateJourneyView( // journey resource
					journey2ResourceId,
					journeyName2,
					duration,
					elementsBlock,
				) + generateJourneyView( // journey resource
					journey3ResourceId,
					journeyName3,
					duration,
					elementsBlock,
				) + generateJourneyViewDataSource( // journey data source
					dataSource1Id,
					strconv.Quote(journeyName1),
					"genesyscloud_journey_viewe."+journey1ResourceId,
				) + generateJourneyViewDataSource( // journey data source
					dataSource2Id,
					strconv.Quote(journeyName2),
					"genesyscloud_journey_view."+journey2ResourceId,
				) + generateJourneyViewDataSource( // queue data source
					dataSource3Id,
					strconv.Quote(journeyName3),
					"genesyscloud_journey_view."+journey3ResourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_journey_view."+journey1ResourceId, "id",
						"data.genesyscloud_journey_view."+dataSource1Id, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_journey_view."+journey2ResourceId, "id",
						"data.genesyscloud_journey_view."+dataSource2Id, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_journey_view."+journey3ResourceId, "id",
						"data.genesyscloud_journey_view."+dataSource3Id, "id"),
				),
			},
		},
		CheckDestroy: testVerifyJourneyViewsDestroyed,
	})
}

func generateJourneyViewDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_journey_view" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
