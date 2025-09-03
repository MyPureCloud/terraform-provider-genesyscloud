package business_rules_decision_table

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// dataSourceBusinessRulesDecisionTableRead reads a Genesys Cloud business rules decision table by name
func dataSourceBusinessRulesDecisionTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query for business rules decision tables by name. Retry in case new table is not yet indexed by search.
	// As table names are non-unique, fail in case of multiple results.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		tables, retryable, resp, err := proxy.getBusinessRulesDecisionTablesByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error getting business rules decision table %s | error: %v", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no business rules decision table found with name %s", name), resp))
		}

		if len(*tables) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("ambiguous business rules decision table name: %s", name), resp))
		}

		table := (*tables)[0]
		d.SetId(*table.Id)

		// Set the basic fields
		resourcedata.SetNillableValue(d, "name", table.Name)
		resourcedata.SetNillableValue(d, "description", table.Description)
		if table.Division != nil {
			d.Set("division_id", table.Division.Id)
		}

		// Set columns - try to get from table first, then from latest version if needed
		var columns interface{}
		var schemaID string

		// Get providers from proxy (allows injection of mock providers during testing)
		queueLookup := proxy.GetQueueLookupProvider()
		schemaLookup := proxy.GetSchemaLookupProvider()

		// If no providers are set in proxy, create default ones
		if queueLookup == nil {
			queueLookup = NewDefaultQueueLookupProvider(sdkConfig)
		}
		if schemaLookup == nil {
			schemaLookup = NewDefaultSchemaLookupProvider(sdkConfig)
		}

		log.Printf("Debug: Table Columns field is nil: %v", table.Columns == nil)

		// Try to get columns from the table response first
		if table.Columns != nil {
			log.Printf("Debug: Found columns in table response")
			columns = buildTerraformColumns(table.Columns, queueLookup, schemaLookup, schemaID, ctx)
			log.Printf("Debug: Built columns from table: %+v", columns)
		}

		// If columns weren't found in table response, try to get them from the latest version
		if columns == nil && table.Latest != nil {
			log.Printf("Debug: Trying to get columns from latest version")
			latestVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, *table.Id, int(*table.Latest.Version))
			if err == nil && latestVersion.Columns != nil {
				log.Printf("Debug: Found columns in latest version")
				// Get schema ID for column type detection
				if latestVersion.Contract != nil && latestVersion.Contract.ParentSchema != nil {
					schemaID = *latestVersion.Contract.ParentSchema.Id
				}
				columns = buildTerraformColumns(latestVersion.Columns, queueLookup, schemaLookup, schemaID, ctx)
				log.Printf("Debug: Built columns from version: %+v", columns)
			} else {
				log.Printf("Debug: No columns found in latest version or error: %v", err)
			}
		}

		// Set columns if we found them and they have actual content
		if columns != nil {
			// Check if the columns map has actual content (not just empty inputs/outputs)
			if columnsMap, ok := columns.(map[string]interface{}); ok {
				hasInputs := columnsMap["inputs"] != nil && len(columnsMap["inputs"].([]interface{})) > 0
				hasOutputs := columnsMap["outputs"] != nil && len(columnsMap["outputs"].([]interface{})) > 0

				if hasInputs || hasOutputs {
					log.Printf("Debug: Setting columns in data source: %+v", columns)
					d.Set("columns", []interface{}{columns})
				} else {
					log.Printf("Debug: Columns map is empty, not setting columns")
				}
			} else {
				log.Printf("Debug: Columns is not a map, not setting columns")
			}
		} else {
			log.Printf("Debug: No columns found to set")
		}

		// Get schema_id from the latest version's contract
		if table.Latest != nil {
			// Fetch the actual latest version to get the contract (for schema_id)
			latestVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, *table.Id, int(*table.Latest.Version))
			if err == nil && latestVersion.Contract != nil && latestVersion.Contract.ParentSchema != nil {
				d.Set("schema_id", latestVersion.Contract.ParentSchema.Id)
			}
		}

		// Set version information
		if table.Latest != nil && table.Latest.Version != nil {
			d.Set("latest_version", *table.Latest.Version)
		}
		if table.Published != nil && table.Published.Version != nil {
			d.Set("published_version", *table.Published.Version)
		}

		log.Printf("Successfully read Business Rules Decision Table by name: %s", name)
		return nil
	})
}
