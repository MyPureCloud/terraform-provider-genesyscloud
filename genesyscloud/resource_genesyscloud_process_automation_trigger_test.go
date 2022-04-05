package genesyscloud

import (
	"fmt"
    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "testing"
    "strconv"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"encoding/json"
)

func TestAccResourceProcessAutomationTrigger(t *testing.T) {
	var (
        triggerResource1 = "test-trigger1"

        triggerName1                = "Terraform trigger1-" + uuid.NewString()
        topicName1                  = "v2.detail.events.conversation.{id}.customer.end"
        enabled1                    = "true"
        targetType1                 = "Workflow"
        match_criteria_json_path1   = "mediaType"
        match_criteria_operator1    = "Equal"
        match_criteria_value1       = "CHAT"
        eventTtlSeconds1            = "60"

        triggerName2                = "Terraform trigger2-" + uuid.NewString()
        enabled2                    = "false"
        match_criteria_json_path2   = "disconnectType"
        match_criteria_operator2    = "NotEqual"
        match_criteria_value2       = "CLIENT"
        eventTtlSeconds2            = "120"

		flowResource1 = "test_flow1"
        filePath1     = "../examples/resources/genesyscloud_process_automation_trigger/trigger_workflow_example.yaml"
		flowName1     = "terraform-provider-test-" + uuid.NewString()
		workflowConfig1 = fmt.Sprintf(`workflow:
 name: %s
 division: Home
 startUpRef: "/workflow/states/state[Initial State_10]"
 defaultLanguage: en-us
 variables:
     - stringVariable:
         name: Flow.dateActiveQueuesChanged
         initialValue:
           noValue: true
         isInput: true
         isOutput: false
     - stringVariable:
         name: Flow.id
         initialValue:
           noValue: true
         isInput: true
         isOutput: false
 settingsErrorHandling:
   errorHandling:
     endWorkflow:
       none: true
 states:
   - state:
       name: Initial State
       refId: Initial State_10
       actions:
         - endWorkflow:
             name: End Workflow
             exitReason:
               noValue: true`, flowName1)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create flow and trigger
                Config: generateFlowResource(
                    flowResource1,
                    filePath1,
                    workflowConfig1,
                ) + generateProcessAutomationTriggerResource(
					triggerResource1,
					triggerName1,
					topicName1,
					enabled1,
					fmt.Sprintf(`jsonencode(%s)`, generateJsonObject(
                            generateJsonProperty("id", "genesyscloud_flow."+flowResource1+".id"),
                            generateJsonProperty("type", strconv.Quote(targetType1)),
                        ),
                    ),
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
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "event_ttl_seconds", eventTtlSeconds1),
                    validateTargetFlowId("genesyscloud_flow."+flowResource1, "genesyscloud_process_automation_trigger."+triggerResource1),
                    validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "target", "type", targetType1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1,
                                                    "match_criteria",
                                                    fmt.Sprintf(`[{%s,%s,%s}]`,
                                                        fmt.Sprintf("\"jsonPath\":%s", strconv.Quote(match_criteria_json_path1)),
                                                        fmt.Sprintf("\"operator\":%s", strconv.Quote(match_criteria_operator1)),
                                                        fmt.Sprintf("\"value\":%s", strconv.Quote(match_criteria_value1)),
                                                    ),
                    ),
                ),
			},
			{
                // Update trigger name, enabled, eventTTLSeconds and match criteria
                Config: generateFlowResource(
                    flowResource1,
                    filePath1,
                    workflowConfig1,
                ) + generateProcessAutomationTriggerResource(
                   triggerResource1,
                   triggerName2,
                   topicName1,
                   enabled2,
                   fmt.Sprintf(`jsonencode(%s)`, generateJsonObject(
                           generateJsonProperty("id", "genesyscloud_flow."+flowResource1+".id"),
                           generateJsonProperty("type", strconv.Quote(targetType1)),
                       ),
                   ),
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
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1, "event_ttl_seconds", eventTtlSeconds2),
                    validateTargetFlowId("genesyscloud_flow."+flowResource1, "genesyscloud_process_automation_trigger."+triggerResource1),
                    validateValueInJsonAttr("genesyscloud_process_automation_trigger."+triggerResource1, "target", "type", targetType1),
                    resource.TestCheckResourceAttr("genesyscloud_process_automation_trigger."+triggerResource1,
                                                    "match_criteria",
                                                    fmt.Sprintf(`[{%s,%s,%s}]`,
                                                        fmt.Sprintf("\"jsonPath\":%s", strconv.Quote(match_criteria_json_path2)),
                                                        fmt.Sprintf("\"operator\":%s", strconv.Quote(match_criteria_operator2)),
                                                        fmt.Sprintf("\"value\":%s", strconv.Quote(match_criteria_value2)),
                                                    ),
                    ),
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

func generateProcessAutomationTriggerResource(resourceID, name, topic_name, enabled, target, match_criteria, event_ttl_seconds string) string {
	return fmt.Sprintf(`resource "genesyscloud_process_automation_trigger" "%s" {
        name = "%s"
        topic_name = "%s"
        enabled = %s
        target = %s
        match_criteria = %s
        event_ttl_seconds = %s
	}
	`, resourceID, name, topic_name, enabled, target, match_criteria, event_ttl_seconds)
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

func validateTargetFlowId(flowResourceName string, triggerResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourceName]
		if !ok {
			return fmt.Errorf("Failed to find flow %s in state", flowResourceName)
		}
		triggerResource, ok := state.RootModule().Resources[triggerResourceName]
		if !ok {
            return fmt.Errorf("Failed to find trigger %s in state", triggerResourceName)
        }

		flowID := flowResource.Primary.ID
		targetAttr, ok := triggerResource.Primary.Attributes["target"]

		var jsonMap map[string]interface{}
        if err := json.Unmarshal([]byte(targetAttr), &jsonMap); err != nil {
            return fmt.Errorf("Error parsing JSON for %s in state: %v", triggerResource.Primary.ID, err)
        }

		if flowID != jsonMap["id"] {
			return fmt.Errorf("Flow in trigger was not created correctly. Expect: %s, Actual: %s", flowID, jsonMap["id"])
		}

		return nil
	}
}