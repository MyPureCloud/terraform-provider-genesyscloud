package journey_action_map

import (
	"fmt"
	"path"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/fileserver"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

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
	server := fileserver.Start(httpServerExitDone, path.Join("../", testrunner.GetTestDataPath(testrunner.ResourceTestType, ResourceType)), port)

	runJourneyActionMapTestCase(t, testCaseName)

	fileserver.ShutDown(server, httpServerExitDone)
}

func runJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	SetupJourneyActionMap(t, testCaseName, sdkConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(ResourceType, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
	})
}

func testVerifyJourneyActionMapsDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
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
