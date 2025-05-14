package journey_view_schedule

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceJourneyViewSchedule(t *testing.T) {
	var (
		journeyViewResourceLabel = "test_journey_view_resource"
		journeyViewName          = "test_journey_view_name"
		journeyViewDuration      = "P1Y"
		scheduleResourceLabel    = "test_journey_view_schedule"
		frequencyDaily           = "Daily"
		frequencyWeekly          = "Weekly"
		journeyViewConfig        = generateJourneyViewResource(journeyViewResourceLabel, journeyViewName, journeyViewDuration)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: journeyViewConfig + generateJourneyViewScheduleResource(
					journeyViewResourceLabel,
					scheduleResourceLabel,
					frequencyDaily),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+scheduleResourceLabel, "frequency", frequencyDaily),
					resource.TestCheckResourceAttrPair(ResourceType+"."+scheduleResourceLabel, "journey_view_id",
						"genesyscloud_journey_views."+journeyViewResourceLabel, "id"),
				),
			},
			{
				// Update
				Config: journeyViewConfig + generateJourneyViewScheduleResource(
					journeyViewResourceLabel,
					scheduleResourceLabel,
					frequencyWeekly,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+scheduleResourceLabel, "frequency", frequencyWeekly),
					resource.TestCheckResourceAttrPair(ResourceType+"."+scheduleResourceLabel, "journey_view_id",
						"genesyscloud_journey_views."+journeyViewResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + scheduleResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyJourneyViewScheduleDestroyed,
	})
}

func generateJourneyViewResource(journeyViewResourceLabel string, journeyViewName string, journeyViewDuration string) string {
	return fmt.Sprintf(`resource "genesyscloud_journey_views" "%s" {
		name = "%s"
		duration = "%s"
	}
	`, journeyViewResourceLabel, journeyViewName, journeyViewDuration)
}

func generateJourneyViewScheduleResource(journeyViewResourceLabel string, scheduleResourceLabel string, scheduleFrequency string) string {
	return fmt.Sprintf(`resource "genesyscloud_journey_view_schedule" "%s" {
		journey_view_id = genesyscloud_journey_views.%s.id
		frequency = "%s"
	}
	`, scheduleResourceLabel, journeyViewResourceLabel, scheduleFrequency)
}

func testVerifyJourneyViewScheduleDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		schedule, resp, err := journeyApi.GetJourneyViewSchedules(rs.Primary.ID)
		if schedule != nil {
			return fmt.Errorf("journey view schedule %s still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Schedule not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success.
	return nil
}
