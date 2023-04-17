package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v98/platformclientv2"
)

func TestAccResourceProcessAutomationTrigger(t *testing.T) {
	t.Parallel()
	var (
		triggerResource1 = "test-trigger1"

		triggerName1              = "Terraform trigger1-" + uuid.NewString()
		topicName1                = "v2.detail.events.conversation.{id}.customer.end"
		enabled1                  = "true"
		targetType1               = "Workflow"
		match_criteria_json_path1 = "mediaType"
		match_criteria_operator1  = "Equal"
		match_criteria_value1     = "CHAT"
		eventTtlSeconds1          = "60"
		delayBySeconds1           = "60"
		description1              = "description1"

		triggerName2              = "Terraform trigger2-" + uuid.NewString()
		enabled2                  = "false"
		match_criteria_json_path2 = "disconnectType"
		match_criteria_operator2  = "In"
		match_criteria_value2     = "CLIENT"
		eventTtlSeconds2          = "120"
		delayBySeconds2           = "90"
		description2              = "description2"

		flowResource1 = "test_flow1"
		filePath1     = "../examples/resources/genesyscloud_processautomation_trigger/trigger_workflow_example.yaml"
		flowName1     = "terraform-provider-test-" + uuid.NewString()
	)

	var homeDivisionName string
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					getHomeDivisionName("data.genesyscloud_auth_division_home.home", &homeDivisionName),
				),
			},
		},
	})

	workflowConfig1 := fmt.Sprintf(`workflow:
 name: %s
 division: %s
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
               noValue: true`, flowName1, homeDivisionName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create flow and trigger
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceEventTTL(
					triggerResource1,
					triggerName1,
					topicName1,
					enabled1,
					fmt.Sprintf(`target {
                        id = %s
                        type = "%s"
                    }
                    `, "genesyscloud_flow."+flowResource1+".id", targetType1),
					fmt.Sprintf(`match_criteria {
                        json_path = "%s"
                        operator = "%s"
                        value = "%s"
                    }
                    `, match_criteria_json_path1, match_criteria_operator1, match_criteria_value1),
					eventTtlSeconds1,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "name", triggerName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "enabled", enabled1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "event_ttl_seconds", eventTtlSeconds1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "description", description1),
					validateTargetFlowId("genesyscloud_flow."+flowResource1, "genesyscloud_processautomation_trigger."+triggerResource1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResource1, targetType1),
					validateMatchCriteriaWithValue("genesyscloud_processautomation_trigger."+triggerResource1, match_criteria_json_path1, match_criteria_operator1, match_criteria_value1, 0),
				),
			},
			{
				// Update trigger name, enabled, eventTTLSeconds and match criteria
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceEventTTL(
					triggerResource1,
					triggerName2,
					topicName1,
					enabled2,
					fmt.Sprintf(`target {
                        id = %s
                        type = "%s"
                    }
                    `, "genesyscloud_flow."+flowResource1+".id", targetType1),
					fmt.Sprintf(`match_criteria {
                        json_path = "%s"
                        operator = "%s"
                        values = ["%s"]
                    }
                    `, match_criteria_json_path2, match_criteria_operator2, match_criteria_value2),
					eventTtlSeconds2,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "name", triggerName2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "enabled", enabled2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "event_ttl_seconds", eventTtlSeconds2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "description", description2),
					validateTargetFlowId("genesyscloud_flow."+flowResource1, "genesyscloud_processautomation_trigger."+triggerResource1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResource1, targetType1),
					validateMatchCriteriaWithValues("genesyscloud_processautomation_trigger."+triggerResource1, match_criteria_json_path2, match_criteria_operator2, []string{match_criteria_value2}, 0),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_processautomation_trigger." + triggerResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyProcessAutomationTriggerDestroyed,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create flow and trigger
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceDelayBy(
					triggerResource1,
					triggerName1,
					topicName1,
					enabled1,
					fmt.Sprintf(`target {
                        id = %s
                        type = "%s"
                    }
                    `, "genesyscloud_flow."+flowResource1+".id", targetType1),
					fmt.Sprintf(`match_criteria {
                        json_path = "%s"
                        operator = "%s"
                        value = "%s"
                    }
                    `, match_criteria_json_path1, match_criteria_operator1, match_criteria_value1),
					delayBySeconds1,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "name", triggerName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "enabled", enabled1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "delay_by_seconds", delayBySeconds1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "description", description1),
					validateTargetFlowId("genesyscloud_flow."+flowResource1, "genesyscloud_processautomation_trigger."+triggerResource1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResource1, targetType1),
					validateMatchCriteriaWithValue("genesyscloud_processautomation_trigger."+triggerResource1, match_criteria_json_path1, match_criteria_operator1, match_criteria_value1, 0),
				),
			},
			{
				// Update trigger name, enabled, eventTTLSeconds and match criteria
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceDelayBy(
					triggerResource1,
					triggerName2,
					topicName1,
					enabled2,
					fmt.Sprintf(`target {
                        id = %s
                        type = "%s"
                    }
                    `, "genesyscloud_flow."+flowResource1+".id", targetType1),
					fmt.Sprintf(`match_criteria {
                        json_path = "%s"
                        operator = "%s"
                        values = ["%s"]
                    }
                    `, match_criteria_json_path2, match_criteria_operator2, match_criteria_value2),
					delayBySeconds2,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "name", triggerName2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "enabled", enabled2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "delay_by_seconds", delayBySeconds2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResource1, "description", description2),
					validateTargetFlowId("genesyscloud_flow."+flowResource1, "genesyscloud_processautomation_trigger."+triggerResource1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResource1, targetType1),
					validateMatchCriteriaWithValues("genesyscloud_processautomation_trigger."+triggerResource1, match_criteria_json_path2, match_criteria_operator2, []string{match_criteria_value2}, 0),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_processautomation_trigger." + triggerResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyProcessAutomationTriggerDestroyed,
	})
}

func generateProcessAutomationTriggerResourceEventTTL(resourceID, name, topic_name, enabled, target, match_criteria, event_ttl_seconds, description string) string {
	return fmt.Sprintf(`resource "genesyscloud_processautomation_trigger" "%s" {
        name = "%s"
        topic_name = "%s"
        enabled = %s
        %s
        %s
        event_ttl_seconds = %s
		description = "%s"
	}
	`, resourceID, name, topic_name, enabled, target, match_criteria, event_ttl_seconds, description)
}

func generateProcessAutomationTriggerResourceDelayBy(resourceID, name, topic_name, enabled, target, match_criteria, delay_by_seconds, description string) string {
	return fmt.Sprintf(`resource "genesyscloud_processautomation_trigger" "%s" {
        name = "%s"
        topic_name = "%s"
        enabled = %s
        %s
        %s
		delay_by_seconds = %s
		description = "%s"
	}
	`, resourceID, name, topic_name, enabled, target, match_criteria, delay_by_seconds, description)
}

func testVerifyProcessAutomationTriggerDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_processautomation_trigger" {
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

		if flowID != triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".id"] {
			return fmt.Errorf("Flow in trigger was not created correctly. Expect: %s, Actual: %s", flowID, triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".id"])
		}

		return nil
	}
}

func validateTargetType(triggerResourceName string, typeVal string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		triggerResource, ok := state.RootModule().Resources[triggerResourceName]
		if !ok {
			return fmt.Errorf("Failed to find trigger %s in state", triggerResourceName)
		}

		if typeVal != triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".type"] {
			return fmt.Errorf("Type in trigger target was not created correctly. Expect: %s, Actual: %s", typeVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".type"])
		}

		return nil
	}
}

func validateMatchCriteriaWithValue(triggerResourceName string, jsonPathVal string, operatorVal string, value string, position int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		triggerResource, ok := state.RootModule().Resources[triggerResourceName]
		if !ok {
			return fmt.Errorf("Failed to find trigger %s in state", triggerResourceName)
		}

		if jsonPathVal != triggerResource.Primary.Attributes["match_criteria."+strconv.Itoa(position)+".json_path"] {
			return fmt.Errorf("Match Criteria json_path in trigger was not created correctly. Expect: %s, Actual: %s", jsonPathVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(position)+".json_path"])
		}

		if operatorVal != triggerResource.Primary.Attributes["match_criteria."+strconv.Itoa(position)+".operator"] {
			return fmt.Errorf("Match Criteria operator in trigger was not created correctly. Expect: %s, Actual: %s", jsonPathVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(position)+".operator"])
		}

		if value != triggerResource.Primary.Attributes["match_criteria."+strconv.Itoa(position)+".value"] {
			return fmt.Errorf("Match Criteria value in trigger was not created correctly. Expect: %s, Actual: %s", jsonPathVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(position)+".value"])
		}

		return nil
	}
}

func validateMatchCriteriaWithValues(triggerResourceName string, jsonPathVal string, operatorVal string, values []string, position int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		triggerResource, ok := state.RootModule().Resources[triggerResourceName]
		if !ok {
			return fmt.Errorf("Failed to find trigger %s in state", triggerResourceName)
		}

		if jsonPathVal != triggerResource.Primary.Attributes["match_criteria."+strconv.Itoa(position)+".json_path"] {
			return fmt.Errorf("Match Criteria json_path in trigger was not created correctly. Expect: %s, Actual: %s", jsonPathVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(position)+".json_path"])
		}

		if operatorVal != triggerResource.Primary.Attributes["match_criteria."+strconv.Itoa(position)+".operator"] {
			return fmt.Errorf("Match Criteria operator in trigger was not created correctly. Expect: %s, Actual: %s", jsonPathVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(position)+".operator"])
		}

		if values[0] != triggerResource.Primary.Attributes["match_criteria."+strconv.Itoa(position)+".values."+strconv.Itoa(0)] {
			return fmt.Errorf("Match Criteria values in trigger was not created correctly. Expect: %s, Actual: %s", jsonPathVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(position)+".values."+strconv.Itoa(0)])
		}

		return nil
	}
}
