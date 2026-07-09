package journey_action_map

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"
	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/fileserver"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

func TestAccResourceJourneyActionMapActionMediaTypes(t *testing.T) {
	var (
		uniqueId         = uuid.NewString()[:8]
		actionMapLabel   = "test_action_map_" + uniqueId
		segmentLabel     = "test_segment_" + uniqueId
		flowLabel        = "test_flow_" + uniqueId
		actionMapName    = "tf_test_action_map_" + uniqueId
		actionMapNameUpd = "tf_test_action_map_" + uniqueId + "_upd"
		segmentName      = "tf_test_segment_" + uniqueId
		flowFilePath     = filepath.Join(testrunner.GetTestDataPath(testrunner.ResourceTestType, ResourceType), "action_media_types_journey_action_map_dependency_flow.yaml")
	)

	SetupJourneyActionMap(t, "action_media_types", sdkConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
		Steps: []resource.TestStep{
			{
				// Step 1: Create action map with webMessagingOffer (no flow)
				Config: generateSegmentConfig(segmentLabel, segmentName) +
					generateActionMapWebMessagingConfig(actionMapLabel, actionMapName, segmentLabel, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "display_name", actionMapName),
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "action.0.media_type", "webMessagingOffer"),
				),
			},
			{
				// Step 2: Update with web_messaging_offer_fields + flow
				Config: generateSegmentConfig(segmentLabel, segmentName) +
					generateFlowConfig(flowLabel, flowFilePath) +
					generateActionMapWebMessagingWithFlowConfig(actionMapLabel, actionMapNameUpd, segmentLabel, flowLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "display_name", actionMapNameUpd),
				),
			},
			{
				// Step 3: Remove web_messaging_offer_fields
				Config: generateSegmentConfig(segmentLabel, segmentName) +
					generateActionMapWebMessagingConfig(actionMapLabel, actionMapNameUpd, segmentLabel, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "action.0.media_type", "webMessagingOffer"),
				),
			},
			{
				// Step 4: Switch to architectFlow media type
				Config: generateSegmentConfig(segmentLabel, segmentName) +
					generateFlowConfig(flowLabel, flowFilePath) +
					generateActionMapArchitectFlowConfig(actionMapLabel, actionMapNameUpd, segmentLabel, flowLabel, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "action.0.media_type", "architectFlow"),
				),
			},
			{
				// Step 5: Add flow_request_mappings
				Config: generateSegmentConfig(segmentLabel, segmentName) +
					generateFlowConfig(flowLabel, flowFilePath) +
					generateActionMapArchitectFlowConfig(actionMapLabel, actionMapNameUpd, segmentLabel, flowLabel, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "action.0.architect_flow_fields.0.flow_request_mappings.#", "2"),
				),
			},
			{
				// ImportState
				ResourceName:      "genesyscloud_journey_action_map." + actionMapLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceJourneyActionMapActionMediaTypesWithTriggerConditions(t *testing.T) {
	runJourneyActionMapTestCase(t, "action_media_types_with_trigger_conditions")
}

func TestAccResourceJourneyActionMapOptionalAttributes(t *testing.T) {
	runJourneyActionMapTestCase(t, "basic_optional_attributes")
}

func TestAccResourceJourneyActionMapRequiredAttributes(t *testing.T) {
	var (
		uniqueId         = uuid.NewString()[:8]
		actionMapLabel   = "test_req_map_" + uniqueId
		segmentLabel     = "test_req_seg_" + uniqueId
		flowLabel        = "test_req_flow_" + uniqueId
		actionMapName    = "tf_test_req_map_" + uniqueId
		actionMapNameUpd = "tf_test_req_map_" + uniqueId + "_upd"
		segmentName      = "tf_test_req_seg_" + uniqueId
		flowFilePath     = filepath.Join(testrunner.GetTestDataPath(testrunner.ResourceTestType, ResourceType), "basic_required_attributes_journey_action_map_dependency_flow.yaml")
	)

	SetupJourneyActionMap(t, "basic_required_attributes", sdkConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
		Steps: []resource.TestStep{
			{
				// Step 1: Create with required attributes only
				Config: generateRequiredSegmentConfig(segmentLabel, segmentName) +
					generateActionMapWebMessagingConfig(actionMapLabel, actionMapName, segmentLabel, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "display_name", actionMapName),
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "action.0.media_type", "webMessagingOffer"),
				),
			},
			{
				// Step 2: Update with architectFlow + delay activation
				Config: generateRequiredSegmentConfig(segmentLabel, segmentName) +
					generateFlowConfig(flowLabel, flowFilePath) +
					generateActionMapRequiredUpdate(actionMapLabel, actionMapNameUpd, segmentLabel, flowLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "display_name", actionMapNameUpd),
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "action.0.media_type", "architectFlow"),
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "activation.0.type", "delay"),
					resource.TestCheckResourceAttr("genesyscloud_journey_action_map."+actionMapLabel, "activation.0.delay_in_seconds", "60"),
				),
			},
			{
				// ImportState
				ResourceName:      "genesyscloud_journey_action_map." + actionMapLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceJourneyActionMapScheduleGroups(t *testing.T) {
	runJourneyActionMapTestCase(t, "schedule_groups")
}

func runJourneyActionMapTestCaseWithFileServer(t *testing.T, testCaseName string, port int) {
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	server := fileserver.Start(httpServerExitDone, testrunner.GetTestDataPath(testrunner.ResourceTestType, ResourceType), port)

	runJourneyActionMapTestCase(t, testCaseName)

	fileserver.ShutDown(server, httpServerExitDone)
}

func runJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	SetupJourneyActionMap(t, testCaseName, sdkConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceJourneyTestSteps(ResourceType, testCaseName, nil),
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
			continue
		}

		return fmt.Errorf("unexpected error: %s", err)
	}
	return nil
}

// --- Helper functions for inline test configs ---

func generateSegmentConfig(label, name string) string {
	return fmt.Sprintf(`
resource "genesyscloud_journey_segment" "%s" {
  display_name            = "%s"
  color                   = "#008000"
  should_display_to_agent = true
  journey {
    patterns {
      criteria {
        key                = "page.hostname"
        values             = ["something_else"]
        operator           = "equal"
        should_ignore_case = false
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
      event_name   = "EventName"
    }
  }
}
`, label, name)
}

func generateFlowConfig(label, filePath string) string {
	return architectFlow.GenerateFlowResource(
		label,
		filePath,
		"",
		false,
		`substitutions = {
			flow_name            = "tf_test_flow_`+label+`"
			default_language     = "en-us"
			greeting             = "Hello World"
			menu_disconnect_name = "Disconnect"
		}`,
	)
}

func generateActionMapWebMessagingConfig(label, name, segmentLabel, webMsgFields string) string {
	return fmt.Sprintf(`
resource "genesyscloud_journey_action_map" "%s" {
  display_name          = "%s"
  trigger_with_segments = [genesyscloud_journey_segment.%s.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "webMessagingOffer"
    %s
  }
  start_date = "2022-07-04T12:00:00.000000"
}
`, label, name, segmentLabel, webMsgFields)
}

func generateActionMapWebMessagingWithFlowConfig(label, name, segmentLabel, flowLabel string) string {
	return fmt.Sprintf(`
resource "genesyscloud_journey_action_map" "%s" {
  display_name          = "%s"
  trigger_with_segments = [genesyscloud_journey_segment.%s.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "webMessagingOffer"
    web_messaging_offer_fields {
      offer_text        = "Offer text"
      architect_flow_id = genesyscloud_flow.%s.id
    }
  }
  start_date = "2022-07-04T12:00:00.000000"
}
`, label, name, segmentLabel, flowLabel)
}

func generateActionMapArchitectFlowConfig(label, name, segmentLabel, flowLabel string, withMappings bool) string {
	mappings := ""
	if withMappings {
		mappings = `
      flow_request_mappings {
        name           = "Name_1"
        attribute_type = "String"
        mapping_type   = "Lookup"
        value          = "session.id"
      }
      flow_request_mappings {
        name           = "Name_2"
        attribute_type = "Integer"
        mapping_type   = "HardCoded"
        value          = "999"
      }`
	}
	return fmt.Sprintf(`
resource "genesyscloud_journey_action_map" "%s" {
  display_name          = "%s"
  trigger_with_segments = [genesyscloud_journey_segment.%s.id]
  activation {
    type = "immediate"
  }
  action {
    media_type = "architectFlow"
    architect_flow_fields {
      architect_flow_id = genesyscloud_flow.%s.id
      %s
    }
  }
  start_date = "2022-07-04T12:00:00.000000"
}
`, label, name, segmentLabel, flowLabel, mappings)
}

func generateRequiredSegmentConfig(label, name string) string {
	return fmt.Sprintf(`
resource "genesyscloud_journey_segment" "%s" {
  display_name            = "%s"
  color                   = "#008000"
  should_display_to_agent = false
  journey {
    patterns {
      criteria {
        key                = "page.title"
        values             = ["Title"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
}
`, label, name)
}

func generateActionMapRequiredUpdate(label, name, segmentLabel, flowLabel string) string {
	return fmt.Sprintf(`
resource "genesyscloud_journey_action_map" "%s" {
  display_name          = "%s"
  trigger_with_segments = [genesyscloud_journey_segment.%s.id]
  activation {
    type             = "delay"
    delay_in_seconds = 60
  }
  action {
    media_type = "architectFlow"
    architect_flow_fields {
      architect_flow_id = genesyscloud_flow.%s.id
    }
  }
  start_date = "2022-07-05T15:30:00.000000"
}
`, label, name, segmentLabel, flowLabel)
}
