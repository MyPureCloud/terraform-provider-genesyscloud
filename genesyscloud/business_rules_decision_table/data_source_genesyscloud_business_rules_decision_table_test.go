package business_rules_decision_table

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
Test Class for the business rules decision table Data Source
*/

func TestAccDataSourceBusinessRulesDecisionTable(t *testing.T) {
	t.Parallel()

	enabled, businessRulesDecisionTableResp, queueResp := businessRulesDecisionTableFtIsEnabled()
	if !enabled {
		t.Skipf("Skipping test as required permissions are not configured, decision table: %s, queues: %s", businessRulesDecisionTableResp.Status, queueResp.Status)
		return
	}

	var (
		schemaResourceLabel = "test-schema"
		schemaName          = "tf_schema_" + uuid.NewString()
		schemaDescription   = "Test schema for decision table data source testing"

		tableResourceLabel = "test-decision-table"
		tableName          = "tf_decision_table_" + uuid.NewString()
		tableDescription   = "Test decision table for data source testing"

		queueResourceLabel = "test-queue"
		queueName          = "tf_test_queue_" + uuid.NewString()

		tableDataSourceLabel = "business_rules_decision_table_data_source_1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create everything in one step: schema, queue, decision table, and data source
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResourceWithQueues(
						tableResourceLabel,
						tableName,
						tableDescription,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						queueResourceLabel,
					) +
					generateBusinessRulesDecisionTableDataSource(
						tableDataSourceLabel,
						tableName,
						"genesyscloud_business_rules_decision_table."+tableResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					// Verify the data source returns the same ID as the resource
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "id",
						"genesyscloud_business_rules_decision_table."+tableResourceLabel, "id",
					),
					// Verify all computed fields are populated
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "name", tableName,
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "description", tableDescription,
					),
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "division_id",
						"data.genesyscloud_auth_division_home.home", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "schema_id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel, "id",
					),
					// Verify complex column structure
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.#", "1",
					),
					// Verify input columns exist (2 inputs: customer_type, priority)
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.#", "2",
					),
					// Verify output columns exist (2 outputs: transfer_queue, skill)
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.#", "2",
					),

					// Verify column IDs are set (computed fields)
					resource.TestCheckResourceAttrSet("data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.0.id"),
					resource.TestCheckResourceAttrSet("data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.1.id"),
					resource.TestCheckResourceAttrSet("data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.0.id"),
					resource.TestCheckResourceAttrSet("data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.1.id"),

					// Verify first input column properties (customer_type)
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.0.defaults_to.0.special", "Wildcard",
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.0.expression.0.contractual.0.schema_property_key", "customer_type",
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.0.expression.0.comparator", "Equals",
					),
					// Verify second input column properties (priority)
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.1.defaults_to.0.special", "Wildcard",
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.1.expression.0.contractual.0.schema_property_key", "priority",
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.inputs.1.expression.0.comparator", "Equals",
					),

					// Verify first output column properties (transfer_queue with queue reference)
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.0.defaults_to.0.value",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.0.value.0.schema_property_key", "transfer_queue",
					),
					// Verify second output column properties (skill)
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.1.defaults_to.0.special", "Null",
					),
					resource.TestCheckResourceAttr(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "columns.0.outputs.1.value.0.schema_property_key", "skill",
					),

					// Verify version fields are populated
					resource.TestCheckResourceAttrSet(
						"data.genesyscloud_business_rules_decision_table."+tableDataSourceLabel, "latest_version",
					),
					// published_version might be null for new tables, so we don't check it
				),
			},
		},
	})
}

// generateBusinessRulesDecisionTableDataSource generates a data source for testing
func generateBusinessRulesDecisionTableDataSource(
	dataSourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_business_rules_decision_table" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, dataSourceLabel, name, dependsOnResource)
}
