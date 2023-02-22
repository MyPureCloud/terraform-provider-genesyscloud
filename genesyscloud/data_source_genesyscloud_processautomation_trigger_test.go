package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceProcessAutomationTrigger(t *testing.T) {
	var (
		triggerResource1 = "test-trigger1"
		triggerResource2 = "test-trigger2"

		triggerName1             = "Terraform trigger1-" + uuid.NewString()
		topicName1               = "v2.detail.events.conversation.{id}.customer.end"
		enabled1                 = "true"
		targetType1              = "Workflow"
		match_criteria_json_path = "mediaType"
		match_criteria_operator  = "Equal"
		match_criteria_value     = "CHAT"
		eventTtlSeconds1         = "60"
		description1             = "description 1"

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
				// Create a trigger
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
                    `, match_criteria_json_path, match_criteria_operator, match_criteria_value),
					eventTtlSeconds1,
					description1,
				) + generateProcessAutomationTriggerDataSource(
					triggerResource2,
					triggerName1,
					"genesyscloud_processautomation_trigger."+triggerResource1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_processautomation_trigger."+triggerResource2, "id", "genesyscloud_processautomation_trigger."+triggerResource1, "id"), // Default value would be "DISABLED"
				),
			},
		},
	})

}

func getHomeDivisionName(key string, divisionName *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		homeDivision, ok := state.RootModule().Resources[key]
		if !ok {
			return fmt.Errorf("Failed to find home division")
		}
		*divisionName = homeDivision.Primary.Attributes["name"]
		return nil
	}
}

func generateProcessAutomationTriggerDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_processautomation_trigger" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
