package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/fileserver"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

const resourceName = "genesyscloud_journey_action_map"

func TestAccResourceJourneyActionMapActionMediaTypes(t *testing.T) {
	runJourneyActionMapTestCaseWithFileServer(t, "action_media_types", 8111)
}

func TestAccResourceJourneyActionMapActionMediaTypesWithTriggerConditions(t *testing.T) {
	runJourneyActionMapTestCase(t, "action_media_types_with_trigger_conditions")
}

func TestAccResourceJourneyActionMapOptionalAttributes(t *testing.T) {
	runJourneyActionMapTestCase(t, "basic_optional_attributes")
}

func TestAccResourceJourneyActionMapRequiredAttributes(t *testing.T) {
	runJourneyActionMapTestCaseWithFileServer(t, "basic_required_attributes", 8112)
}

func TestAccResourceJourneyActionMapScheduleGroups(t *testing.T) {
	runJourneyActionMapTestCase(t, "schedule_groups")
}

func runJourneyActionMapTestCaseWithFileServer(t *testing.T, testCaseName string, port int) {
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	server := fileserver.Start(httpServerExitDone, testrunner.GetTestDataPath(testrunner.ResourceTestType, resourceName), port)

	runJourneyActionMapTestCase(t, testCaseName)

	fileserver.ShutDown(server, httpServerExitDone)
}

func runJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	setupJourneyActionMap(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(resourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
	})
}

func setupJourneyActionMap(t *testing.T, testCaseName string) {
	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	testCasePrefix := testrunner.TestObjectIdPrefix + testCaseName
	cleanupJourneySegments(testCasePrefix)
	cleanupArchitectScheduleGroups(testCasePrefix)
	cleanupArchitectSchedules(testCasePrefix)
	cleanupFlows(testCasePrefix)
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

		if IsStatus404(resp) {
			// Journey action map not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey action map destroyed
	return nil
}
