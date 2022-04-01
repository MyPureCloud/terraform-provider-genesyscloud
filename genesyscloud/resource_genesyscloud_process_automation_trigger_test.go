package genesyscloud

import (
	"fmt"
    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "testing"
    "strconv"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func TestAccResourceProcessAutomationTrigger(t *testing.T) {
	var (
        triggerResource1 = "test-trigger1"

        triggerName1                = "Terraform trigger1-" + uuid.NewString()
        topicName1                  = "v2.detail.events.conversation.{id}.customer.end"
        enabled1                    = "true"
        targetId1                   = "ae1e0cde-875d-4d13-a498-615e7a9fe956"
        targetType1                 = "Workflow"
        match_criteria_json_path1   = "mediaType"
        match_criteria_operator1    = "Equal"
        match_criteria_value1       = "CHAT"
        eventTtlSeconds1            = "60"

        triggerName2                = "Terraform trigger2-" + uuid.NewString()
//         topicName2                  = "v2.detail.events.conversation.{id}.customer.start"
        enabled2                    = "false"
//         targetId2                   = "ae1e0cde-875d-4d13-a498-615e7a9fe956"
//         targetType2                 = "Workflow"
        match_criteria_json_path2   = "disconnectType"
        match_criteria_operator2    = "NotEqual"
        match_criteria_value2       = "CLIENT"
        eventTtlSeconds2            = "120"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create a trigger
				Config: generateProcessAutomationTriggerResource(
					triggerResource1,
					triggerName1,
					topicName1,
					enabled1,
					targetId1,
					targetType1,
                    fmt.Sprintf(`jsonencode([%s])`, generateJsonObject(
                            generateJsonProperty("jsonPath", strconv.Quote(match_criteria_json_path1)),
                            generateJsonProperty("operator", strconv.Quote(match_criteria_operator1)),
                            generateJsonProperty("value", strconv.Quote(match_criteria_value1)),
                        ),
                    ),
					eventTtlSeconds1,
				),
				Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "name", triggerName1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "topic_name", topicName1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "enabled", enabled1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "target_id", targetId1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "target_type", targetType1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "event_ttl_seconds", eventTtlSeconds1),
                    //TODO: figure out how to validate the match criteria
//                     validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "match_criteria", "jsonPath", match_criteria_json_path1),
//                     validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "match_criteria", "operator", match_criteria_operator1),
//                     validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "match_criteria", "value", match_criteria_value1),
                ),
			},
			{
                // Update trigger name, enabled, eventTTLSeconds and match criteria
                Config: generateProcessAutomationTriggerResource(
                   triggerResource1,
                   triggerName2,
                   topicName1,
                   enabled2,
                   targetId1,
                   targetType1,
                   fmt.Sprintf(`jsonencode([%s])`, generateJsonObject(
                           generateJsonProperty("jsonPath", strconv.Quote(match_criteria_json_path2)),
                           generateJsonProperty("operator", strconv.Quote(match_criteria_operator2)),
                           generateJsonProperty("value", strconv.Quote(match_criteria_value2)),
                       ),
                   ),
                   eventTtlSeconds2,
                ),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "name", triggerName2),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "topic_name", topicName1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "enabled", enabled2),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "target_id", targetId1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "target_type", targetType1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "event_ttl_seconds", eventTtlSeconds2),
                    //TODO: figure out how to validate the match criteria
//                     validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "match_criteria", "jsonPath", match_criteria_json_path1),
//                     validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "match_criteria", "operator", match_criteria_operator1),
//                     validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "match_criteria", "value", match_criteria_value1),
                ),
            },
            {
                // Import/Read
                ResourceName:      "genesyscloud_process_automation_trigger." + triggerResource1,
                ImportState:       true,
                ImportStateVerify: true,
            },
		},
		CheckDestroy: testVerifyProcessAutomationTriggerDestroyed,
	})
}

func generateProcessAutomationTriggerResource(resourceID, name, topic_name, enabled, target_id, target_type, match_criteria, event_ttl_seconds string) string {
	return fmt.Sprintf(`resource "genesyscloud_process_automation_trigger" "%s" {
        name = "%s"
        topic_name = "%s"
        enabled = %s
        target_id = "%s"
        target_type = "%s"
        match_criteria = %s
        event_ttl_seconds = %s
	}
	`, resourceID, name, topic_name, enabled, target_id, target_type, match_criteria, event_ttl_seconds)
}

func testVerifyProcessAutomationTriggerDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_process_automation_trigger" {
			continue
		}

		trigger, resp, err := getProcessAutomationTrigger(rs.Primary.ID, integrationAPI)
		if trigger != nil {
			return fmt.Errorf("Process automation trigger (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// Trigger not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All triggers destroyed
	return nil
}