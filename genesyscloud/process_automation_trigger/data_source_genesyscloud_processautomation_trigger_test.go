package process_automation_trigger

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceProcessAutomationTrigger(t *testing.T) {
	var (
		triggerResourceLabel1 = "test-trigger1"
		triggerResourceLabel2 = "test-trigger2"

		triggerName1              = "Terraform trigger1-" + uuid.NewString()
		topicName1                = "v2.detail.events.conversation.{id}.customer.end"
		enabled1                  = "true"
		targetType1               = "Workflow"
		eventTtlSeconds1          = "60"
		description1              = "description 1"
		workflowTargetDataFormat1 = "Json"

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

	//Need to have a JSON encoded path
	matchCriteria := `([
        {
          "jsonPath": "direction",
          "operator": "Equal",
          "value": "INBOUND"
        },
        {
          "jsonPath": "mediaType",
          "operator": "Equal",
          "value": "VOICE"
        },
        {
          "jsonPath": "interactingDurationMs",
          "operator": "LessThanOrEqual",
          "value": 20000
        }
     ])`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a trigger
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
                    `, "genesyscloud_flow."+flowResourceLabel1+".id", targetType1, workflowTargetDataFormat1),
					matchCriteria,
					eventTtlSeconds1,
					description1,
				) + generateProcessAutomationTriggerDataSource(
					triggerResourceLabel2,
					triggerName1,
					"genesyscloud_processautomation_trigger."+triggerResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_processautomation_trigger."+triggerResourceLabel2, "id", "genesyscloud_processautomation_trigger."+triggerResourceLabel1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func generateProcessAutomationTriggerDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_processautomation_trigger" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
