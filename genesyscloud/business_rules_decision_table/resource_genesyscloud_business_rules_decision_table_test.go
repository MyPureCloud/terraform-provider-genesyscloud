package business_rules_decision_table

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceBusinessRulesDecisionTable(t *testing.T) {
	t.Parallel()

	enabled, businessRulesDecisionTableResp, queueResp := businessRulesDecisionTableFtIsEnabled()
	if !enabled {
		t.Skipf("Skipping test as required permissions are not configured, decision table: %s, queues: %s", businessRulesDecisionTableResp.Status, queueResp.Status)
		return
	}

	var (
		// Resource labels
		tableResourceLabel  = "test-decision-table"
		schemaResourceLabel = "test-schema"
		queueResourceLabel  = "test-queue"

		// Table names and descriptions
		tableName1 = "TF Test DT1-" + uuid.NewString()[:8]
		tableName2 = "TF Test DT2-" + uuid.NewString()[:8]
		tableDesc1 = "Terraform test decision table1"
		tableDesc2 = "Terraform test decision table2"

		// Schema and queue properties for testing
		schemaName        = "TF Test Schema-" + uuid.NewString()[:8]
		schemaDescription = "Test schema for decision table testing"
		queueName         = "TF Test Queue-" + uuid.NewString()[:8]
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Step 1: Create complex decision table with routing queue
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResourceWithQueues(
						tableResourceLabel,
						tableName1,
						tableDesc1,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						queueResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					// Verify basic resource attributes
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "name", tableName1),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "description", tableDesc1),

					// Verify complex column structure
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.#", "2"),

					// Verify column IDs are set (computed fields)
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.id"),

					// Verify first input column (customer_type)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.defaults_to.0.special", "Wildcard"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.expression.0.contractual.0.schema_property_key", "customer_type"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.expression.0.comparator", "Equals"),

					// Verify second input column (priority)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.defaults_to.0.special", "Wildcard"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.expression.0.contractual.0.schema_property_key", "priority"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.expression.0.comparator", "Equals"),

					// Verify first output column (transfer_queue with queue reference)
					resource.TestCheckResourceAttrPair(
						"genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.defaults_to.0.value",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.value.0.schema_property_key", "transfer_queue"),

					// Verify second output column (skill)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.defaults_to.0.special", "Null"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.value.0.schema_property_key", "skill"),
				),
			},
			{
				// Step 2: Update with new name and description (keep same complex columns)
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResourceWithQueues(
						tableResourceLabel,
						tableName2,
						tableDesc2,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						queueResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					// Verify updated basic resource attributes
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "name", tableName2),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "description", tableDesc2),

					// Verify complex column structure is maintained
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.#", "2"),

					// Verify column IDs are still set after update (computed fields)
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.id"),

					// Verify queue reference is still valid
					resource.TestCheckResourceAttrPair(
						"genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.defaults_to.0.value",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
				),
			},
			{
				// Step 3: Test import functionality
				ResourceName:      "genesyscloud_business_rules_decision_table." + tableResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyBusinessRulesDecisionTablesDestroyed,
	})
}

// testVerifyBusinessRulesDecisionTablesDestroyed verifies that all decision tables are properly destroyed
func testVerifyBusinessRulesDecisionTablesDestroyed(state *terraform.State) error {
	businessRulesAPI := platformclientv2.NewBusinessRulesApi()

	// Check decision tables
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_business_rules_decision_table" {
			decisionTable, resp, err := sdkGetBusinessRulesDecisionTable(rs.Primary.ID, businessRulesAPI)
			if decisionTable != nil {
				return fmt.Errorf("Business Rules Decision Table (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// Decision table not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error checking decision table: %s", err)
			}
		}
	}
	// Success. All decision tables destroyed
	return nil
}

// sdkGetBusinessRulesDecisionTable is a helper function to get a decision table directly via SDK
func sdkGetBusinessRulesDecisionTable(tableId string, api *platformclientv2.BusinessRulesApi) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
	return api.GetBusinessrulesDecisiontable(tableId)
}

// Helper functions for generating test dependencies

func generateBusinessRulesSchemaResource(resourceLabel, name, description string) string {
	return fmt.Sprintf(`resource "genesyscloud_business_rules_schema" "%s" {
		enabled = true
		name = "%s"
		description = "%s"
		properties = jsonencode({
			"customer_type" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/enum"
					}
				],
				"title" : "customer_type",
				"description" : "Customer type for routing decisions",
				"enum" : ["VIP", "Standard", "Premium"],
				"_enumProperties" : {
					"VIP" : {
						"title" : "VIP Customer"
					},
					"Standard" : {
						"title" : "Standard Customer"
					},
					"Premium" : {
						"title" : "Premium Customer"
					}
				}
			},
			"priority" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "priority",
				"description" : "Priority level for routing",
				"minLength" : 1,
				"maxLength" : 10
			},
			"transfer_queue" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/businessRulesQueue"
					}
				],
				"title" : "transfer_queue",
				"description" : "Transfer queue for routing"
			},
			"skill" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "skill",
				"description" : "Skill for routing",
				"minLength" : 1,
				"maxLength" : 100
			}
		})
	}
	`, resourceLabel, name, description)
}

func generateHomeDivisionReference() string {
	return "\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n"
}
