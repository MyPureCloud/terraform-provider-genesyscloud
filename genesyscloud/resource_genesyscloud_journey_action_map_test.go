package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestAccResourceJourneyActionMapActionMediaTypes(t *testing.T) {
	runJourneyActionMapTestCase(t, "action_media_types")
}

func TestAccResourceJourneyActionMapOptionalAttributes(t *testing.T) {
	runJourneyActionMapTestCase(t, "basic_optional_attributes")
}

func TestAccResourceJourneyActionMapRequiredAttributes(t *testing.T) {
	runJourneyActionMapTestCase(t, "basic_required_attributes")
}

func TestAccResourceJourneyActionMapScheduleGroups(t *testing.T) {
	runJourneyActionMapTestCase(t, "schedule_groups")
}

func runJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_action_map"
	setupJourneyActionMap(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps:             testrunner.GenerateResourceTestSteps(resourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
	})
}

func setupJourneyActionMap(t *testing.T, testCaseName string) {
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	testCasePrefix := testrunner.TestObjectIdPrefix + testCaseName
	cleanupJourneySegments(testCasePrefix)
	cleanupArchitectScheduleGroups(testCasePrefix)
	cleanupArchitectSchedules(testCasePrefix)
	cleanupJourneyActionMaps(testCasePrefix)
}

func cleanupJourneyActionMaps(idPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		actionMaps, _, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
		if getErr != nil {
			return
		}

		if actionMaps.Entities == nil || len(*actionMaps.Entities) == 0 {
			break
		}

		for _, actionMap := range *actionMaps.Entities {
			if actionMap.DisplayName != nil && strings.HasPrefix(*actionMap.DisplayName, idPrefix) {
				_, delErr := journeyApi.DeleteJourneyActionmap(*actionMap.Id)
				if delErr != nil {
					diag.Errorf("failed to delete journey action map %s (%s): %s", *actionMap.Id, *actionMap.DisplayName, delErr)
					return
				}
				log.Printf("Deleted journey action map %s (%s)", *actionMap.Id, *actionMap.DisplayName)
			}
		}

		pageCount = *actionMaps.PageCount
	}
}

func testVerifyJourneyActionMapsDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_action_map" {
			continue
		}

		actionMap, resp, err := journeyApi.GetJourneyActionmap(rs.Primary.ID)
		if actionMap != nil {
			return fmt.Errorf("journey action map (%s) still exists", rs.Primary.ID)
		}

		if isStatus404(resp) {
			// Journey action map not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey action map destroyed
	return nil
}
