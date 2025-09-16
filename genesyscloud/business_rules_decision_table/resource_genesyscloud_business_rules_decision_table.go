package business_rules_decision_table

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	platformclientv2 "github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// getAllBusinessRulesDecisionTables retrieves all Genesys Cloud business rules decision tables
func getAllBusinessRulesDecisionTables(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getBusinessRulesDecisionTableProxy(clientConfig)

	log.Printf("Retrieving all Business Rules Decision Tables")

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	tables, resp, err := proxy.getAllBusinessRulesDecisionTables(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all business rules decision tables error: %s", err), resp)
	}

	if tables.Entities != nil {
		for _, table := range *tables.Entities {
			if table.Id != nil {
				resources[*table.Id] = &resourceExporter.ResourceMeta{BlockLabel: *table.Name}
			}
		}
	}

	log.Printf("Successfully retrieved all Business Rules Decision Tables")
	return resources, nil
}

// createBusinessRulesDecisionTable creates a Genesys Cloud business rules decision table
func createBusinessRulesDecisionTable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	// Build the create request
	createRequest := buildCreateRequest(d)
	tableName := d.Get("name").(string)

	log.Printf("Creating business rules decision table with name: %s", tableName)

	// Create the decision table (version 1 is created automatically)
	tableVersion, resp, err := proxy.createBusinessRulesDecisionTable(ctx, createRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create business rules decision table error: %s", err), resp)
	}

	log.Printf("Successfully created business rules decision table with ID: %s", *tableVersion.Id)

	// Set the resource ID (version ID and table ID are the same)
	d.SetId(*tableVersion.Id)

	return readBusinessRulesDecisionTable(ctx, d, meta)
}

// readBusinessRulesDecisionTable reads a Genesys Cloud business rules decision table
func readBusinessRulesDecisionTable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceBusinessRulesDecisionTable(), constants.ConsistencyChecks(), ResourceType)

	tableId := d.Id()
	log.Printf("Reading business rules decision table %s", tableId)

	return util.WithRetriesForReadCustomTimeout(ctx, 1*time.Minute, d, func() *retry.RetryError {
		table, resp, err := proxy.getBusinessRulesDecisionTable(ctx, tableId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Business rules decision table %s not found", tableId)
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read business rules decision table %s | error: %s", tableId, err), resp))
		}

		// Set the basic fields
		resourcedata.SetNillableValue(d, "name", table.Name)
		resourcedata.SetNillableValue(d, "description", table.Description)
		resourcedata.SetNillableReferenceDivision(d, "division_id", table.Division)

		// Set columns directly from the table (always available)
		if table.Columns != nil {
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

			// Get schema ID for column type detection
			var schemaID string
			if table.Latest != nil {
				// Fetch the actual latest version to get the contract (for schema_id)
				latestVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, int(*table.Latest.Version))
				if err == nil && latestVersion.Contract != nil && latestVersion.Contract.ParentSchema != nil {
					schemaID = *latestVersion.Contract.ParentSchema.Id
				}
			}

			columns := flattenColumns(table.Columns, queueLookup, schemaLookup, schemaID, ctx)
			d.Set("columns", []interface{}{columns})
		}

		// Get schema_id from the latest version's contract
		if table.Latest != nil {
			// Fetch the actual latest version to get the contract (for schema_id)
			latestVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, int(*table.Latest.Version))
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

		log.Printf("Read business rules decision table %s %s", tableId, *table.Name)
		return cc.CheckState(d)
	})
}

// updateBusinessRulesDecisionTable updates a Genesys Cloud business rules decision table
func updateBusinessRulesDecisionTable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	tableId := d.Id()
	log.Printf("Updating Business Rules Decision Table: %s", tableId)

	// Get current table state for validation
	currentTable, resp, err := proxy.getBusinessRulesDecisionTable(ctx, tableId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get current business rules decision table %s error: %s", tableId, err), resp)
	}

	// Validate column updates (only allowed on version 1 draft)
	if d.HasChange("columns") {
		if err := validateColumnUpdate(ctx, proxy, currentTable); err != nil {
			return diag.FromErr(err)
		}
	}

	// Create update request
	updateRequest := &platformclientv2.Updatedecisiontablerequest{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	// Add columns if they're being updated
	if d.HasChange("columns") {
		columns := d.Get("columns").([]interface{})
		if len(columns) > 0 {
			columnsRequest := buildSdkUpdateColumns(columns[0].(map[string]interface{}))
			updateRequest.Columns = columnsRequest
		}
	}

	// Update the decision table
	_, resp, err = proxy.updateBusinessRulesDecisionTable(ctx, tableId, updateRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update business rules decision table %s error: %s", tableId, err), resp)
	}

	log.Printf("Successfully updated Business Rules Decision Table: %s", tableId)
	return readBusinessRulesDecisionTable(ctx, d, meta)
}

// deleteBusinessRulesDecisionTable deletes a Genesys Cloud business rules decision table
func deleteBusinessRulesDecisionTable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	tableId := d.Id()
	log.Printf("Deleting Business Rules Decision Table: %s", tableId)

	// Delete the decision table
	resp, err := proxy.deleteBusinessRulesDecisionTable(ctx, tableId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete business rules decision table %s error: %s", tableId, err), resp)
	}

	log.Printf("Successfully deleted Business Rules Decision Table: %s", tableId)

	// Wait for deletion to complete with retries
	diag := util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		// Try to get the table to see if it still exists
		_, resp, err := proxy.getBusinessRulesDecisionTable(ctx, tableId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted business rules decision table %s", tableId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting business rules decision table %s | error: %s", tableId, err), resp))
		}

		// Table still exists, retry
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("business rules decision table %s still exists", tableId), resp))
	})

	// Clear the ID after successful deletion
	if !diag.HasError() {
		d.SetId("")
	}

	return diag
}

// validateColumnUpdate validates that column updates are allowed
func validateColumnUpdate(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, table *platformclientv2.Decisiontable) error {
	if table.Latest == nil {
		return fmt.Errorf("cannot determine table version")
	}

	// Get the actual latest version to check its details
	latestVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, *table.Id, int(*table.Latest.Version))
	if err != nil {
		return fmt.Errorf("failed to get latest version details: %s", err)
	}

	// Check if this is version 1
	if latestVersion.Version == nil {
		return fmt.Errorf("column updates are only allowed on version 1, but version is nil")
	}
	if *latestVersion.Version != 1 {
		return fmt.Errorf("column updates are only allowed on version 1, current version is %d", *latestVersion.Version)
	}

	// Check if version 1 is published
	if table.Published != nil && table.Published.Version != nil && *table.Published.Version == 1 {
		return fmt.Errorf("version 1 is published, column updates are not allowed")
	}

	// Check if current version is in draft status
	if latestVersion.Status == nil {
		return fmt.Errorf("column updates are only allowed when version 1 is in 'Draft' status, but status is nil")
	}
	if *latestVersion.Status != "Draft" {
		return fmt.Errorf("column updates are only allowed when version 1 is in 'Draft' status, current status is '%s'", *latestVersion.Status)
	}

	return nil
}
