package business_rules_decision_table

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceBusinessRulesDecisionTableHappyPath(t *testing.T) {
	t.Parallel()
	os.Setenv("GENESYSCLOUD_SDK_DEBUG", "true")
	os.Setenv("GENESYSCLOUD_SDK_DEBUG_FORMAT", "Text")
	os.Setenv("GENESYSCLOUD_SDK_DEBUG_FILE_PATH", "./test_debug.log")
	os.Setenv("GENESYSCLOUD_SDK_CLIENT_POOL_DEBUG", "true")

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
				// Step 1: Create complex decision table with routing queue and rows using all literal types
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						tableResourceLabel,
						tableName1,
						tableDesc1,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRows(queueResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					// Verify basic resource attributes
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "name", tableName1),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "description", tableDesc1),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "version", "1"),

					// Verify complex column structure (7 inputs + 2 outputs for comprehensive test)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.#", "7"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.#", "2"),

					// Verify column IDs are set (computed fields) - all 7 inputs + 2 outputs
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.2.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.3.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.4.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.5.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.6.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.id"),

					// Verify input columns (comprehensive schema)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.expression.0.contractual.0.schema_property_key", "customer_type"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.expression.0.contractual.0.schema_property_key", "customer_name"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.2.expression.0.contractual.0.schema_property_key", "priority_level"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.3.expression.0.contractual.0.schema_property_key", "score"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.4.expression.0.contractual.0.schema_property_key", "created_date"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.5.expression.0.contractual.0.schema_property_key", "last_updated"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.6.expression.0.contractual.0.schema_property_key", "is_active"),

					// Verify first output column (transfer_queue with queue reference)
					resource.TestCheckResourceAttrPair(
						"genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.defaults_to.0.value",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.value.0.schema_property_key", "transfer_queue"),

					// Verify second output column (skill)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.defaults_to.0.special", "Null"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.value.0.schema_property_key", "skill"),

					// Verify rows are present with all literal types
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.#", "7"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.#", "2"),

					// CRITICAL: Verify inputs are in the EXACT order we configured them
					// This test will fail if the provider doesn't maintain consistent ordering
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.schema_property_key", "customer_type"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.value", "VIP"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.type", "string"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.schema_property_key", "customer_name"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.value", "John Doe"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.type", "string"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.schema_property_key", "priority_level"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.literal.0.value", "5"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.literal.0.type", "integer"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.schema_property_key", "score"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.literal.0.value", "85.5"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.literal.0.type", "number"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.schema_property_key", "created_date"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.literal.0.value", "2023-01-15"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.literal.0.type", "date"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.schema_property_key", "last_updated"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.literal.0.value", "2023-01-15T10:30:00.000Z"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.literal.0.type", "datetime"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.schema_property_key", "is_active"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.literal.0.value", "true"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.literal.0.type", "boolean"),

					// Verify outputs are also in correct order
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.0.schema_property_key", "transfer_queue"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.0.literal.0.value",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.0.literal.0.type", "string"),

					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.schema_property_key", "skill"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.literal.0.value", "Premium Support"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.literal.0.type", "string"),

					// Verify computed fields
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.row_id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.row_index"),
				),
			},
			{
				// Step 2: Update with special values to test all literal types
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						tableResourceLabel,
						tableName2,
						tableDesc2,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRowsWithSpecials(queueResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					// Verify updated basic resource attributes
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "name", tableName2),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "description", tableDesc2),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "version", "2"),

					// Verify complex column structure is maintained (7 inputs + 2 outputs)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.#", "7"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.#", "2"),

					// Verify column IDs are still set after update (computed fields) - all 7 inputs + 2 outputs
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.1.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.2.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.3.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.4.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.5.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.6.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.1.id"),

					// Verify queue reference is still valid
					resource.TestCheckResourceAttrPair(
						"genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.0.defaults_to.0.value",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),

					// Verify rows with special values are present after update
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.#", "7"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.#", "2"),

					// Verify special values in row inputs (using valid comparator combinations)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.value", "Standard"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.value", "Jane Smith"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.literal.0.value", "Wildcard"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.literal.0.type", "special"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.literal.0.value", "Wildcard"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.literal.0.type", "special"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.literal.0.value", "CurrentTime"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.literal.0.type", "special"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.literal.0.value", "CurrentTime"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.literal.0.type", "special"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.literal.0.value", "Null"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.literal.0.type", "special"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.literal.0.value", "Standard Support"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.literal.0.type", "string"),
				),
			},
			{
				// Step 3: Update rows (add, update, delete) - keep comprehensive schema
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						tableResourceLabel,
						tableName2,
						tableDesc2,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateUpdatedRows(queueResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					// Verify updated basic attributes
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "name", tableName2),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "description", tableDesc2),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "version", "3"),

					// Verify updated rows (should have 2 rows now with comprehensive schema)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.#", "7"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.#", "7"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.outputs.#", "2"),

					// Verify first row (Alice Johnson with Premium enum)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.value", "Premium"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.value", "Alice Johnson"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.literal.0.value", "8"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.2.literal.0.type", "integer"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.literal.0.value", "92.3"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.3.literal.0.type", "number"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.literal.0.value", "2023-02-20"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.4.literal.0.type", "date"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.literal.0.value", "2023-02-20T14:45:30.000Z"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.5.literal.0.type", "datetime"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.literal.0.value", "false"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.6.literal.0.type", "boolean"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.literal.0.value", "Updated Support"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.1.literal.0.type", "string"),

					// Verify second row (Bob Wilson with VIP enum)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.0.literal.0.value", "VIP"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.0.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.1.literal.0.value", "Bob Wilson"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.1.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.2.literal.0.value", "3"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.2.literal.0.type", "integer"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.3.literal.0.value", "67.8"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.3.literal.0.type", "number"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.4.literal.0.value", "2023-03-10"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.4.literal.0.type", "date"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.5.literal.0.value", "2023-03-10T09:15:45.000Z"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.5.literal.0.type", "datetime"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.6.literal.0.value", "true"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.inputs.6.literal.0.type", "boolean"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.outputs.1.literal.0.value", "Standard Support"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.1.outputs.1.literal.0.type", "string"),
				),
			},
			{
				// Step 4: Test import functionality
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
			"customer_name" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "customer_name",
				"description" : "Customer name",
				"minLength" : 1,
				"maxLength" : 100
			},
			"priority_level" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/integer"
					}
				],
				"title" : "priority_level",
				"description" : "Priority level (1-10)",
				"minimum" : 1,
				"maximum" : 10
			},
			"score" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/number"
					}
				],
				"title" : "score",
				"description" : "Customer score",
				"minimum" : 0.0,
				"maximum" : 100.0
			},
			"created_date" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/date"
					}
				],
				"title" : "created_date",
				"description" : "Customer creation date"
			},
			"last_updated" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/datetime"
					}
				],
				"title" : "last_updated",
				"description" : "Last update timestamp"
			},
			"is_active" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/boolean"
					}
				],
				"title" : "is_active",
				"description" : "Whether customer is active"
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

func TestAccResourceBusinessRulesDecisionTableInvalidLiteralValues(t *testing.T) {
	t.Parallel()

	enabled, businessRulesDecisionTableResp, queueResp := businessRulesDecisionTableFtIsEnabled()
	if !enabled {
		t.Skipf("Skipping test as required permissions are not configured, decision table: %s, queues: %s", businessRulesDecisionTableResp.Status, queueResp.Status)
		return
	}

	var (
		schemaResourceLabel = "test-schema"
		schemaName          = "tf_schema_" + uuid.NewString()[:8]
		schemaDescription   = "Test schema for invalid literal testing"
		queueResourceLabel  = "test-queue"
		queueName           = "tf_test_queue_" + uuid.NewString()[:8]
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Test invalid integer value
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table with invalid integer",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRowsWithInvalidLiteral(queueResourceLabel, "integer", "abc"),
					),
				ExpectError: regexp.MustCompile("Failed to add rows: failed to convert row 1: value 'abc' is not a valid integer"),
			},
			{
				// Test invalid number value
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table with invalid number",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRowsWithInvalidLiteral(queueResourceLabel, "number", "not-a-number"),
					),
				ExpectError: regexp.MustCompile("Failed to add rows: failed to convert row 1: value 'not-a-number' is not a valid number"),
			},
			{
				// Test invalid boolean value
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table with invalid boolean",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRowsWithInvalidLiteral(queueResourceLabel, "boolean", "maybe"),
					),
				ExpectError: regexp.MustCompile("Failed to add rows: failed to convert row 1: value 'maybe' is not a valid boolean"),
			},
			{
				// Test invalid date value
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table with invalid date",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRowsWithInvalidLiteral(queueResourceLabel, "date", "not-a-date"),
					),
				ExpectError: regexp.MustCompile("Failed to add rows: failed to convert row 1: value 'not-a-date' is not a valid date"),
			},
			{
				// Test invalid datetime value
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table with invalid datetime",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRowsWithInvalidLiteral(queueResourceLabel, "datetime", "not-a-datetime"),
					),
				ExpectError: regexp.MustCompile("Failed to add rows: failed to convert row 1: value 'not-a-datetime' is not a valid datetime"),
			},
		},
	})
}

func TestAccResourceBusinessRulesDecisionTableRequiredFieldsValidation(t *testing.T) {
	t.Parallel()

	enabled, businessRulesDecisionTableResp, queueResp := businessRulesDecisionTableFtIsEnabled()
	if !enabled {
		t.Skipf("Skipping test as required permissions are not configured, decision table: %s, queues: %s", businessRulesDecisionTableResp.Status, queueResp.Status)
		return
	}

	var (
		schemaResourceLabel = "test-schema"
		schemaName          = "tf_schema_" + uuid.NewString()[:8]
		schemaDescription   = "Test schema for required fields testing"
		queueResourceLabel  = "test-queue"
		queueName           = "tf_test_queue_" + uuid.NewString()[:8]
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Test missing name
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"", // Empty name
						"Test decision table",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateRows(queueResourceLabel),
					),
				ExpectError: regexp.MustCompile(`expected length of name to be in the range`),
			},
			{
				// Test missing schema_id
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					fmt.Sprintf(`resource "genesyscloud_business_rules_decision_table" "test-decision-table" {
		name = "tf_decision_table_%s"
		description = "Test decision table"
		division_id = data.genesyscloud_auth_division_home.home.id
		// schema_id is missing
		%s
		%s
	}
	`, uuid.NewString()[:8], generateColumns(queueResourceLabel), generateRows(queueResourceLabel)),
				ExpectError: regexp.MustCompile("The argument \"schema_id\" is required"),
			},
			{
				// Test invalid schema_id (non-existent)
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table",
						"data.genesyscloud_auth_division_home.home.id",
						"\"invalid-schema-id-12345\"", // Invalid schema_id
						generateColumns(queueResourceLabel),
						generateRows(queueResourceLabel),
					),
				ExpectError: regexp.MustCompile("Failed to create business rules decision table: API Error: 400 - is not a valid UUID"),
			},
		},
	})
}

func TestAccResourceBusinessRulesDecisionTableInvalidColumnReferences(t *testing.T) {
	t.Parallel()

	enabled, businessRulesDecisionTableResp, queueResp := businessRulesDecisionTableFtIsEnabled()
	if !enabled {
		t.Skipf("Skipping test as required permissions are not configured, decision table: %s, queues: %s", businessRulesDecisionTableResp.Status, queueResp.Status)
		return
	}

	var (
		schemaResourceLabel = "test-schema"
		schemaName          = "tf_schema_" + uuid.NewString()[:8]
		schemaDescription   = "Test schema for invalid column references"
		queueResourceLabel  = "test-queue"
		queueName           = "tf_test_queue_" + uuid.NewString()[:8]
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Test with column referencing non-existent schema property
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResourceWithInvalidColumn(
						"test-decision-table",
						"tf_decision_table_"+uuid.NewString()[:8],
						"Test decision table with invalid column reference",
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumnsWithInvalidPropertyKey(queueResourceLabel),
						generateRows(queueResourceLabel),
					),
				ExpectError: regexp.MustCompile("row validation failed: row 1 input 3: schema_property_key 'priority_level' not found in input columns"),
			},
		},
	})
}

func TestAccResourceBusinessRulesDecisionTableWithDefaults(t *testing.T) {
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
		tableName = "TF Test DT Defaults-" + uuid.NewString()[:8]
		tableDesc = "Terraform test decision table with defaults"

		// Schema and queue properties for testing
		schemaName        = "TF Test Schema-" + uuid.NewString()[:8]
		schemaDescription = "Test schema for decision table with defaults testing"
		queueName         = "TF Test Queue-" + uuid.NewString()[:8]
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Step 1: Create decision table with 3 inputs and 1 output, but only specify 2 inputs and 1 output in rows
				// This tests that the API doesn't return all inputs/outputs when defaults are used
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResource(
						tableResourceLabel,
						tableName,
						tableDesc,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumnsWithDefaults(queueResourceLabel),
						generateRowsWithDefaults(queueResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					// Verify basic resource attributes
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "name", tableName),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "description", tableDesc),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "version", "1"),

					// Verify column structure (3 inputs + 1 output)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.inputs.#", "3"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "columns.0.outputs.#", "1"),

					// Verify that only the specified inputs/outputs are present in the row
					// The API may not return all inputs/outputs if they use defaults
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.#", "1"),

					// Check that we have at least the inputs we specified (may be more due to defaults)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.#", "1"),

					// Verify the specific values we set (not using defaults)
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.value", "VIP"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.0.literal.0.type", "string"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.value", "5"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.inputs.1.literal.0.type", "integer"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.0.literal.0.value", "Premium Support"),
					resource.TestCheckResourceAttr("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.outputs.0.literal.0.type", "string"),

					// Verify computed fields
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.row_id"),
					resource.TestCheckResourceAttrSet("genesyscloud_business_rules_decision_table."+tableResourceLabel, "rows.0.row_index"),
				),
			},
		},
		CheckDestroy: testVerifyBusinessRulesDecisionTablesDestroyed,
	})
}

func generateBusinessRulesDecisionTableResourceWithInvalidColumn(
	resourceLabel string,
	name string,
	description string,
	divisionId string,
	schemaId string,
	columns string,
	rows string) string {
	return fmt.Sprintf(`resource "genesyscloud_business_rules_decision_table" "%s" {
		name = "%s"
		description = "%s"
		division_id = %s
		schema_id = %s
		%s
		%s
	}
	`, resourceLabel, name, description, divisionId, schemaId, columns, rows)
}

func generateColumnsWithInvalidPropertyKey(queueResourceLabel string) string {
	return `columns {
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "customer_type"
				}
				comparator = "Equals"
			}
		}
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "customer_name"
				}
				comparator = "Equals"
			}
		}
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "invalid_property_key_that_does_not_exist"
				}
				comparator = "Equals"
			}
		}
		outputs {
			defaults_to {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
			}
			value {
				schema_property_key = "transfer_queue"
				properties {
					schema_property_key = "queue"
					properties {
						schema_property_key = "id"
					}
				}
			}
		}
	}`
}

func generateHomeDivisionReference() string {
	return "\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n"
}
