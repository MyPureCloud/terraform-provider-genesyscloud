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

	// os.Setenv("GENESYSCLOUD_SDK_DEBUG", "true")
	// os.Setenv("GENESYSCLOUD_SDK_DEBUG_FORMAT", "Text")
	// os.Setenv("GENESYSCLOUD_SDK_DEBUG_FILE_PATH", "./test_debug.log")
	// os.Setenv("GENESYSCLOUD_SDK_CLIENT_POOL_DEBUG", "true")

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
					generateBusinessRulesDecisionTableResource(
						tableResourceLabel,
						tableName,
						tableDescription,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumns(queueResourceLabel),
						generateBasicRows(queueResourceLabel),
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
