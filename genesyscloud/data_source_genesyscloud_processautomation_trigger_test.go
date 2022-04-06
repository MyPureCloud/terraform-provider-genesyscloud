package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
	"strconv"
)

func TestAccDataSourceProcessAutomationTrigger(t *testing.T) {
	var (
        triggerResource1 = "test-trigger1"
        triggerResource2 = "test-trigger2"

        triggerName1                = "Terraform trigger1-" + uuid.NewString()
        topicName1                  = "v2.detail.events.conversation.{id}.customer.end"
        enabled1                    = "true"
        targetId1                   = "ae1e0cde-875d-4d13-a498-615e7a9fe956"
        targetType1                 = "Workflow"
        match_criteria_json_path    = "mediaType"
        match_criteria_operator     = "Equal"
        match_criteria_value        = "CHAT"
        eventTtlSeconds1            = "60"
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
					fmt.Sprintf(`jsonencode(%s)`, generateJsonObject(
                            generateJsonProperty("id", strconv.Quote(targetId1)),
                            generateJsonProperty("type", strconv.Quote(targetType1)),
                        ),
                    ),
                    fmt.Sprintf(`jsonencode([%s])`, generateJsonObject(
                            generateJsonProperty("jsonPath", strconv.Quote(match_criteria_json_path)),
                            generateJsonProperty("operator", strconv.Quote(match_criteria_operator)),
                            generateJsonProperty("value", strconv.Quote(match_criteria_value)),
                        ),
                    ),
					eventTtlSeconds1,
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
