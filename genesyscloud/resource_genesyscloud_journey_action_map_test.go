package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/fileserver"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(resourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
	})
}

func setupJourneyActionMap(t *testing.T, testCaseName string) {
	_, err := provider.AuthorizeSdk()
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
				resp, delErr := journeyApi.DeleteJourneyActionmap(*actionMap.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError("genesyscloud_journey_action_map", fmt.Sprintf("failed to delete journey action map %s (%s): %s", *actionMap.Id, *actionMap.DisplayName, delErr), resp)
					return
				}
				log.Printf("Deleted journey action map %s (%s)", *actionMap.Id, *actionMap.DisplayName)
			}
		}

		pageCount = *actionMaps.PageCount
	}
}

func cleanupArchitectSchedules(idPrefix string) {
	architectApi := platformclientv2.NewArchitectApi()

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		architectSchedules, _, getErr := architectApi.GetArchitectSchedules(pageNum, pageSize, "", "", "", nil)
		if getErr != nil {
			return
		}

		if architectSchedules.Entities == nil || len(*architectSchedules.Entities) == 0 {
			break
		}

		for _, schedule := range *architectSchedules.Entities {
			if schedule.Name != nil && strings.HasPrefix(*schedule.Name, idPrefix) {
				resp, delErr := architectApi.DeleteArchitectSchedule(*schedule.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("failed to delete architect schedule %s (%s): %s", *schedule.Id, *schedule.Name, delErr), resp)
					return
				}
				log.Printf("Deleted architect schedule %s (%s)", *schedule.Id, *schedule.Name)
			}
		}
	}
}

func cleanupArchitectScheduleGroups(idPrefix string) {
	architectApi := platformclientv2.NewArchitectApi()

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		architectScheduleGroups, _, getErr := architectApi.GetArchitectSchedulegroups(pageNum, pageSize, "", "", "", "", nil)
		if getErr != nil {
			return
		}

		if architectScheduleGroups.Entities == nil || len(*architectScheduleGroups.Entities) == 0 {
			break
		}

		for _, scheduleGroup := range *architectScheduleGroups.Entities {
			if scheduleGroup.Name != nil && strings.HasPrefix(*scheduleGroup.Name, idPrefix) {
				resp, delErr := architectApi.DeleteArchitectSchedulegroup(*scheduleGroup.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError("genesyscloud_journey_action_map", fmt.Sprintf("failed to delete architect schedule group %s (%s): %s", *scheduleGroup.Id, *scheduleGroup.Name, delErr), resp)
					return
				}
				log.Printf("Deleted architect schedule group %s (%s)", *scheduleGroup.Id, *scheduleGroup.Name)
			}
		}
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

		if util.IsStatus404(resp) {
			// Journey action map not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey action map destroyed
	return nil
}

func cleanupFlows(idPrefix string) {
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 50
		flows, _, getErr := architectApi.GetFlows(nil, pageNum, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if getErr != nil {
			return
		}

		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}

		for _, flow := range *flows.Entities {
			if flow.Name != nil && strings.HasPrefix(*flow.Name, idPrefix) {
				resp, delErr := architectApi.DeleteFlow(*flow.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError("genesyscloud_journey_action_map", fmt.Sprintf("failed to delete flow %s (%s): %s", *flow.Id, *flow.Name, delErr), resp)
					return
				}
				log.Printf("Deleted flow %s (%s)", *flow.Id, *flow.Name)
			}
		}
	}
}
