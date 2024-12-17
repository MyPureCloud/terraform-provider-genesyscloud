package process_automation_trigger

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func TestAccResourceProcessAutomationTrigger(t *testing.T) {
	var (
		triggerResourceLabel1 = "test-trigger1"

		triggerName1                      = "Terraform trigger1-" + uuid.NewString()
		topicName1                        = "v2.detail.events.conversation.{id}.customer.end"
		enabled1                          = "true"
		targetType1                       = "Workflow"
		workflowTargetSettingsDataFormat1 = "Json"
		eventTtlSeconds1                  = "60"
		delayBySeconds1                   = "60"
		description1                      = "description1"

		triggerName2                      = "Terraform trigger2-" + uuid.NewString()
		enabled2                          = "false"
		eventTtlSeconds2                  = "120"
		delayBySeconds2                   = "90"
		description2                      = "description2"
		workflowTargetSettingsDataFormat2 = "TopLevelPrimitives"

		flowResourceLabel1 = "test_flow1"
		filePath1          = "../../examples/resources/genesyscloud_processautomation_trigger/trigger_workflow_example.yaml"
		flowName1          = "terraform-provider-test-" + uuid.NewString()
	)
	var homeDivisionName string
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					util.GetHomeDivisionName("data.genesyscloud_auth_division_home.home", &homeDivisionName),
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

	matchCriteria1 := `[
				{
				  "jsonPath": "mediaType",
				  "operator": "Equal",
				  "value": "CHAT"
				}
	]`

	matchCriteria2 := `[
		{
			"jsonPath": "mediaType",
			"operator": "Equal",
			"value": "VOICE"
		}
	]`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow and trigger
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceEventTTL(
					triggerResourceLabel1,
					triggerName1,
					topicName1,
					enabled1,
					fmt.Sprintf(`target {
                        id = %s
                        type = "%s"
						workflow_target_settings {
							data_format = "%s"
						}
                    }
                    `, "genesyscloud_flow."+flowResourceLabel1+".id", targetType1, workflowTargetSettingsDataFormat1),
					matchCriteria1,
					eventTtlSeconds1,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "name", triggerName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "enabled", enabled1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "event_ttl_seconds", eventTtlSeconds1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "description", description1),
					validateTargetFlowId("genesyscloud_flow."+flowResourceLabel1, "genesyscloud_processautomation_trigger."+triggerResourceLabel1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResourceLabel1, targetType1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "target.0.workflow_target_settings.0.data_format", workflowTargetSettingsDataFormat1),
					testAccCheckMatchCriteria("genesyscloud_processautomation_trigger."+triggerResourceLabel1, matchCriteria1),
				),
			},
			{
				// Update trigger name, enabled, eventTTLSeconds and match criteria
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceEventTTL(
					triggerResourceLabel1,
					triggerName2,
					topicName1,
					enabled2,
					fmt.Sprintf(`target {
			            id = %s
			            type = "%s"
						workflow_target_settings {
							data_format = "%s"
						}
			        }
			        `, "genesyscloud_flow."+flowResourceLabel1+".id", targetType1, workflowTargetSettingsDataFormat1),
					matchCriteria2,
					eventTtlSeconds2,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "name", triggerName2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "enabled", enabled2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "event_ttl_seconds", eventTtlSeconds2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "description", description2),
					validateTargetFlowId("genesyscloud_flow."+flowResourceLabel1, "genesyscloud_processautomation_trigger."+triggerResourceLabel1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResourceLabel1, targetType1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "target.0.workflow_target_settings.0.data_format", workflowTargetSettingsDataFormat1),
					testAccCheckMatchCriteria("genesyscloud_processautomation_trigger."+triggerResourceLabel1, matchCriteria2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_processautomation_trigger." + triggerResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyProcessAutomationTriggerDestroyed,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow and trigger
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceDelayBy(
					triggerResourceLabel1,
					triggerName1,
					topicName1,
					enabled1,
					fmt.Sprintf(`target {
	                    id = %s
	                    type = "%s"
						workflow_target_settings {
							data_format = "%s"
						}
	                }
	                `, "genesyscloud_flow."+flowResourceLabel1+".id", targetType1, workflowTargetSettingsDataFormat2),
					matchCriteria1,
					delayBySeconds1,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "name", triggerName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "enabled", enabled1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "delay_by_seconds", delayBySeconds1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "description", description1),
					validateTargetFlowId("genesyscloud_flow."+flowResourceLabel1, "genesyscloud_processautomation_trigger."+triggerResourceLabel1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResourceLabel1, targetType1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "target.0.workflow_target_settings.0.data_format", workflowTargetSettingsDataFormat2),
					testAccCheckMatchCriteria("genesyscloud_processautomation_trigger."+triggerResourceLabel1, matchCriteria1),
				),
			},
			{
				// Update trigger name, enabled, eventTTLSeconds and match criteria
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					workflowConfig1,
					false,
				) + generateProcessAutomationTriggerResourceDelayBy(
					triggerResourceLabel1,
					triggerName2,
					topicName1,
					enabled2,
					fmt.Sprintf(`target {
	                    id = %s
	                    type = "%s"
						workflow_target_settings {
							data_format = "%s"
						}
	                }
	                `, "genesyscloud_flow."+flowResourceLabel1+".id", targetType1, workflowTargetSettingsDataFormat2),
					matchCriteria2,
					delayBySeconds2,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "name", triggerName2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "topic_name", topicName1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "enabled", enabled2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "delay_by_seconds", delayBySeconds2),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "description", description2),
					validateTargetFlowId("genesyscloud_flow."+flowResourceLabel1, "genesyscloud_processautomation_trigger."+triggerResourceLabel1),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResourceLabel1, targetType1),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel1, "target.0.workflow_target_settings.0.data_format", workflowTargetSettingsDataFormat2),
					testAccCheckMatchCriteria("genesyscloud_processautomation_trigger."+triggerResourceLabel1, matchCriteria2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_processautomation_trigger." + triggerResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyProcessAutomationTriggerDestroyed,
	})
}

func TestAccResourceProcessAutomationTriggerValues(t *testing.T) {
	var (
		triggerResourceLabel = "test-trigger-" + uuid.NewString()

		triggerName                      = "Terraform trigger1-" + uuid.NewString()
		topicName                        = "v2.detail.events.conversation.{id}.customer.end"
		enabled                          = "true"
		targetType                       = "Workflow"
		workflowTargetSettingsDataFormat = "Json"
		eventTtlSeconds                  = "60"
		description                      = "description1"

		flowResourceLabel = "test_flow"
		filePath          = "../../examples/resources/genesyscloud_processautomation_trigger/trigger_workflow_example.yaml"
		flowName          = "terraform-provider-test-" + uuid.NewString()
	)
	var homeDivisionName string
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					util.GetHomeDivisionName("data.genesyscloud_auth_division_home.home", &homeDivisionName),
				),
			},
		},
	})

	workflowConfig := fmt.Sprintf(`workflow:
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
               noValue: true`, flowName, homeDivisionName)

	matchCriteria1 := `[
				{
				  "jsonPath": "mediaType",
				  "operator": "In",
				  "values": ["id1", "id2"]
				}
	]`

	matchCriteria2 := `[
		{
			"jsonPath": "mediaType",
			"operator": "In",
			"values": ["id1", "id2", "id3"]
		}
	]`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow and trigger
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel,
					filePath,
					workflowConfig,
					false,
				) + generateProcessAutomationTriggerResourceEventTTL(
					triggerResourceLabel,
					triggerName,
					topicName,
					enabled,
					fmt.Sprintf(`target {
                        id = %s
                        type = "%s"
						workflow_target_settings {
							data_format = "%s"
						}
                    }
                    `, "genesyscloud_flow."+flowResourceLabel+".id", targetType, workflowTargetSettingsDataFormat),
					matchCriteria1,
					eventTtlSeconds,
					description,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "name", triggerName),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "topic_name", topicName),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "enabled", enabled),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "event_ttl_seconds", eventTtlSeconds),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "description", description),
					validateTargetFlowId("genesyscloud_flow."+flowResourceLabel, "genesyscloud_processautomation_trigger."+triggerResourceLabel),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResourceLabel, targetType),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "target.0.workflow_target_settings.0.data_format", workflowTargetSettingsDataFormat),
					testAccCheckMatchCriteria("genesyscloud_processautomation_trigger."+triggerResourceLabel, matchCriteria1),
				),
			},
			{
				// Update match criteria
				Config: architect_flow.GenerateFlowResource(
					flowResourceLabel,
					filePath,
					workflowConfig,
					false,
				) + generateProcessAutomationTriggerResourceEventTTL(
					triggerResourceLabel,
					triggerName,
					topicName,
					enabled,
					fmt.Sprintf(`target {
			            id = %s
			            type = "%s"
						workflow_target_settings {
							data_format = "%s"
						}
			        }
			        `, "genesyscloud_flow."+flowResourceLabel+".id", targetType, workflowTargetSettingsDataFormat),
					matchCriteria2,
					eventTtlSeconds,
					description,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "name", triggerName),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "topic_name", topicName),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "enabled", enabled),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "event_ttl_seconds", eventTtlSeconds),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "description", description),
					validateTargetFlowId("genesyscloud_flow."+flowResourceLabel, "genesyscloud_processautomation_trigger."+triggerResourceLabel),
					validateTargetType("genesyscloud_processautomation_trigger."+triggerResourceLabel, targetType),
					resource.TestCheckResourceAttr("genesyscloud_processautomation_trigger."+triggerResourceLabel, "target.0.workflow_target_settings.0.data_format", workflowTargetSettingsDataFormat),
					testAccCheckMatchCriteria("genesyscloud_processautomation_trigger."+triggerResourceLabel, matchCriteria2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_processautomation_trigger." + triggerResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyProcessAutomationTriggerDestroyed,
	})
}

func generateProcessAutomationTriggerResourceEventTTL(resourceLabel, name, topic_name, enabled, target, match_criteria, event_ttl_seconds, description string) string {
	return fmt.Sprintf(`resource "genesyscloud_processautomation_trigger" "%s" {
        name = "%s"
        topic_name = "%s"
        enabled = %s
        %s
		match_criteria=jsonencode(%s)
        event_ttl_seconds = %s
		description = "%s"
	}
	`, resourceLabel, name, topic_name, enabled, target, match_criteria, event_ttl_seconds, description)
}

func generateProcessAutomationTriggerResourceDelayBy(resourceLabel, name, topic_name, enabled, target, match_criteria, delay_by_seconds, description string) string {
	return fmt.Sprintf(`resource "genesyscloud_processautomation_trigger" "%s" {
        name = "%s"
        topic_name = "%s"
        enabled = %s
        %s
        match_criteria=jsonencode(%s)
		delay_by_seconds = %s
		description = "%s"
	}
	`, resourceLabel, name, topic_name, enabled, target, match_criteria, delay_by_seconds, description)
}

func testVerifyProcessAutomationTriggerDestroyed(state *terraform.State) error {
	integrationAPI := platformclientv2.NewIntegrationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_processautomation_trigger" {
			continue
		}

		trigger, resp, err := getProcessAutomationTrigger(rs.Primary.ID, integrationAPI)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				//This is to be expected.  We have an error where we dont find what we are looking
				return nil
			} else {
				return fmt.Errorf("Error occurred while trying to getProcessAutomationTrigger %s Err: %s", rs.Primary.ID, err)
			}
		}

		if trigger != nil {
			return fmt.Errorf("Process automation trigger (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			return nil
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}

	// Success. All triggers destroyed
	return nil
}

func validateTargetFlowId(flowResourcePath string, triggerResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find flow %s in state", flowResourcePath)
		}
		triggerResource, ok := state.RootModule().Resources[triggerResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find trigger %s in state", triggerResourcePath)
		}

		flowID := flowResource.Primary.ID

		if flowID != triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".id"] {
			return fmt.Errorf("Flow in trigger was not created correctly. Expect: %s, Actual: %s", flowID, triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".id"])
		}

		return nil
	}
}

func validateTargetType(triggerResourcePath string, typeVal string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		triggerResource, ok := state.RootModule().Resources[triggerResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find trigger %s in state", triggerResourcePath)
		}

		if typeVal != triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".type"] {
			return fmt.Errorf("Type in trigger target was not created correctly. Expect: %s, Actual: %s", typeVal, triggerResource.Primary.Attributes["target."+strconv.Itoa(0)+".type"])
		}

		return nil
	}
}

func testAccCheckMatchCriteria(resourcePath string, targetMatchCriteriaJson string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourcePath]

		if !ok {
			return fmt.Errorf("Resource Not found: %s", resourcePath)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID is not set")
		}

		//Retrieve the match criteria
		resourceMatchCriteriaJson := rs.Primary.Attributes["match_criteria"]

		//Convert the resource and target skill condition to []map. This is an intermediary format.
		var resourceMatchCriteriaMap []map[string]interface{}
		var targetMatchCriteriaMap []map[string]interface{}

		if err := json.Unmarshal([]byte(resourceMatchCriteriaJson), &resourceMatchCriteriaMap); err != nil {
			return fmt.Errorf("error converting resource match criteria from JSON to a Map: %s", err)
		}

		if err := json.Unmarshal([]byte(targetMatchCriteriaJson), &targetMatchCriteriaMap); err != nil {
			return fmt.Errorf("error converting target match criteria to a Map: %s", err)
		}

		//Convert the resource and target maps back to a string so they have the exact same format.
		r, err := json.Marshal(resourceMatchCriteriaMap)
		if err != nil {
			return fmt.Errorf("error converting the resource map back from a Map to JSON: %s", err)
		}
		t, err := json.Marshal(targetMatchCriteriaMap)
		if err != nil {
			return fmt.Errorf("error converting the target map back from a Map to JSON: %s", err)
		}

		//Checking to see if our 2 JSON strings are exactly equal.
		resourceStr := string(r)
		target := string(t)
		if resourceStr != target {
			return fmt.Errorf("resource match criteria does not match match criteria passed in. Expected: %s Actual: %s", resourceStr, target)
		}

		return nil
	}
}
