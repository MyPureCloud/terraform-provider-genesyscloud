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
		log.Printf("Retrieved %d decision tables with published versions", len(*tables.Entities))
	}

	log.Printf("Successfully retrieved Business Rules Decision Tables with published versions")
	return resources, nil
}

// createBusinessRulesDecisionTable creates a Genesys Cloud business rules decision table
func createBusinessRulesDecisionTable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	// Validate required fields
	tableName := d.Get("name").(string)
	if tableName == "" {
		return util.BuildAPIDiagnosticError(ResourceType, "name is required", nil)
	}

	schemaId := d.Get("schema_id").(string)
	if schemaId == "" {
		return util.BuildAPIDiagnosticError(ResourceType, "schema_id is required", nil)
	}

	description := d.Get("description").(string)
	log.Printf("Creating business rules decision table with name: %s, description: %s", tableName, description)

	// Create the decision table
	createRequest := buildCreateRequest(d)
	if createRequest.Description != nil {
		log.Printf("DEBUG: Create request description: %s", *createRequest.Description)
	} else {
		log.Printf("DEBUG: Create request description: nil")
	}
	tableVersionResponse, resp, err := proxy.createBusinessRulesDecisionTable(ctx, createRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create business rules decision table: %s", err), resp)
	}

	// Validate API response
	if tableVersionResponse == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "Received nil response from API", resp)
	}
	if tableVersionResponse.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "Table ID is nil in API response", resp)
	}
	if tableVersionResponse.Version == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "Table version is nil in API response", resp)
	}

	tableId := *tableVersionResponse.Id
	tableVersion := int(*tableVersionResponse.Version)

	log.Printf("Successfully created business rules decision table with ID: %s", tableId)

	// Add rows (required)
	rows := d.Get("rows").([]interface{})
	if len(rows) == 0 {
		return util.BuildAPIDiagnosticError(ResourceType, "At least one row is required", nil)
	}
	log.Printf("Adding %d rows to decision table %s version %d", len(rows), tableId, tableVersion)
	err = addRowsToVersion(ctx, proxy, tableId, tableVersion, rows)
	if err != nil {
		proxy.deleteBusinessRulesDecisionTable(ctx, tableId)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add rows: %s", err), nil)
	}
	log.Printf("Successfully added %d rows to decision table %s version %d", len(rows), tableId, tableVersion)

	// Publish the version
	if err := publishDecisionTableVersion(ctx, proxy, tableId, tableVersion); err != nil {
		proxy.deleteBusinessRulesDecisionTable(ctx, tableId)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish version: %s", err), nil)
	}
	log.Printf("Successfully published decision table %s version %d", tableId, tableVersion)

	d.SetId(tableId)
	log.Printf("Created business rules decision table %s", tableId)
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

		// Get table details to find the published version and table metadata
		table, resp, err := proxy.getBusinessRulesDecisionTable(ctx, tableId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Business rules decision table %s not found", tableId)
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get decision table %s: %s", tableId, err), resp))
		}

		// Determine which version to read (always use published version)
		versionToRead := 1 // Default to version 1
		if table.Published != nil && table.Published.Version != nil {
			versionToRead = int(*table.Published.Version)
		} else {
			log.Printf("No published version found for decision table %s, attempting to use default version 1", tableId)
		}

		// Get the version details for columns and rows
		tableVersion, resp, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, versionToRead)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Business rules decision table %s version %d not found", tableId, versionToRead)
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read decision table version %d for table %s | error: %s", versionToRead, tableId, err), resp))
		}

		// Set name and description from the table (version endpoint doesn't provide these)
		resourcedata.SetNillableValue(d, "name", table.Name)
		resourcedata.SetNillableValue(d, "description", table.Description)
		resourcedata.SetNillableReferenceDivision(d, "division_id", tableVersion.Division)
		resourcedata.SetNillableValue(d, "version", &versionToRead)

		// Set schema_id from the version's contract
		if tableVersion.Contract != nil && tableVersion.Contract.ParentSchema != nil {
			resourcedata.SetNillableValue(d, "schema_id", tableVersion.Contract.ParentSchema.Id)
		}

		// Set columns from the version - error if nil
		if tableVersion.Columns == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("decision table version %d for table %s has no columns", versionToRead, tableId), nil))
		}
		columns := flattenColumns(tableVersion.Columns)
		d.Set("columns", []interface{}{columns})

		// Get stored column order from state
		var storedInputOrder []string
		var storedOutputOrder []string
		if inputOrderInterface := d.Get("input_column_order"); inputOrderInterface != nil {
			if inputOrderList, ok := inputOrderInterface.([]interface{}); ok {
				storedInputOrder = make([]string, len(inputOrderList))
				for i, v := range inputOrderList {
					storedInputOrder[i] = v.(string)
				}
			}
		}
		if outputOrderInterface := d.Get("output_column_order"); outputOrderInterface != nil {
			if outputOrderList, ok := outputOrderInterface.([]interface{}); ok {
				storedOutputOrder = make([]string, len(outputOrderList))
				for i, v := range outputOrderList {
					storedOutputOrder[i] = v.(string)
				}
			}
		}

		// Set rows from the version - error if failed to get rows
		rows, err := getDecisionTableRows(ctx, proxy, tableVersion)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get rows for version %d: %s", versionToRead, err), nil))
		}
		d.Set("rows", rows)

		log.Printf("Read business rules decision table %s version %d", tableId, versionToRead)
		return cc.CheckState(d)
	})
}

// addRowsToVersion adds all rows to a specific decision table version
func addRowsToVersion(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, version int, terraformRows []interface{}) error {
	// Get the table version to extract column mapping
	tableVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, version)
	if err != nil {
		return fmt.Errorf("failed to get table version for column mapping: %s", err)
	}

	// Get column IDs in order for positional mapping
	inputColumnIds, outputColumnIds := extractColumnOrder(tableVersion.Columns)

	// Convert and add each row individually using positional mapping
	for i, row := range terraformRows {
		rowMap := row.(map[string]interface{})

		// Convert Terraform row to SDK format using positional mapping
		sdkRow, err := convertTerraformRowToSDK(rowMap, inputColumnIds, outputColumnIds)
		if err != nil {
			return fmt.Errorf("failed to convert row %d: %s", i+1, err)
		}

		// Add the row to the version
		_, err = proxy.createDecisionTableRow(ctx, tableId, version, &sdkRow)
		if err != nil {
			return fmt.Errorf("failed to add row %d: %s", i+1, err)
		}

		log.Printf("Successfully added row %d to decision table %s version %d", i+1, tableId, version)
	}

	return nil
}

// publishDecisionTableVersion publishes a decision table version
func publishDecisionTableVersion(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, version int) error {
	_, err := proxy.publishDecisionTableVersion(ctx, tableId, version)
	if err != nil {
		return fmt.Errorf("failed to publish version %d: %s", version, err)
	}

	log.Printf("Successfully published decision table %s version %d", tableId, version)
	return nil
}

// getDecisionTableRows retrieves all rows from a specific decision table version
func getDecisionTableRows(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableVersion *platformclientv2.Decisiontableversion) ([]interface{}, error) {
	// Extract tableId and version from tableVersion
	tableId := *tableVersion.Id
	version := *tableVersion.Version

	var allRows []platformclientv2.Decisiontablerow
	const pageSize = 100
	pageNum := 1

	for {
		rowListing, _, err := proxy.getDecisionTableRows(ctx, tableId, version, fmt.Sprintf("%d", pageNum), fmt.Sprintf("%d", pageSize))
		if err != nil {
			return nil, fmt.Errorf("failed to get rows for version %d page %d: %s", version, pageNum, err)
		}

		if rowListing == nil || rowListing.Entities == nil || len(*rowListing.Entities) == 0 {
			break
		}

		allRows = append(allRows, *rowListing.Entities...)

		// Check if there are more pages
		if rowListing.PageCount == nil || pageNum >= *rowListing.PageCount {
			break
		}
		pageNum++
	}

	// Get column IDs in order for positional mapping
	inputColumnIds, outputColumnIds := extractColumnOrder(tableVersion.Columns)

	// Convert SDK rows to Terraform format using positional mapping
	terraformRows := make([]interface{}, len(allRows))
	for i, row := range allRows {
		// For now, use a simple conversion that includes all columns
		terraformRows[i] = convertSDKRowToTerraform(row, inputColumnIds, outputColumnIds)
	}

	return terraformRows, nil
}

// updateBusinessRulesDecisionTable updates a Genesys Cloud business rules decision table
func updateBusinessRulesDecisionTable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	tableId := d.Id()
	log.Printf("Updating Business Rules Decision Table: %s", tableId)

	// Check if name or description has changed
	if d.HasChange("name") || d.HasChange("description") {
		log.Printf("Updating name/description for decision table %s", tableId)

		// Create update request
		updateRequest := buildUpdateRequest(d)

		// Update the decision table
		_, resp, err := proxy.updateBusinessRulesDecisionTable(ctx, tableId, updateRequest)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update business rules decision table %s error: %s", tableId, err), resp)
		}
		log.Printf("Successfully updated name/description for decision table %s", tableId)
	}

	// Check if rows have changed
	if d.HasChange("rows") {
		log.Printf("Rows have changed for decision table %s", tableId)

		// Get old and new row data
		oldRows, newRows := d.GetChange("rows")
		oldRowsList := oldRows.([]interface{})
		newRowsList := newRows.([]interface{})

		// Create new version and update rows
		err := updateDecisionTableRows(ctx, proxy, tableId, oldRowsList, newRowsList)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update rows for decision table %s: %s", tableId, err), nil)
		}
		log.Printf("Successfully updated rows for decision table %s", tableId)
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

// updateDecisionTableRows handles updating rows in a decision table by creating a new version
func updateDecisionTableRows(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, oldRows []interface{}, newRows []interface{}) error {
	log.Printf("Starting row update for decision table %s", tableId)

	// Validate inputs
	if tableId == "" {
		return fmt.Errorf("tableId cannot be empty")
	}
	if newRows == nil {
		return fmt.Errorf("newRows cannot be nil")
	}
	if len(newRows) == 0 {
		return fmt.Errorf("at least one row is required")
	}

	// Step 1: Create new version
	log.Printf("Creating new version for decision table %s", tableId)
	newVersion, _, err := proxy.createDecisionTableVersion(ctx, tableId)
	if err != nil {
		return fmt.Errorf("failed to create new version: %s", err)
	}

	if newVersion.Version == nil {
		return fmt.Errorf("new version number is nil")
	}
	newVersionNumber := *newVersion.Version
	log.Printf("Created new version %d for decision table %s", newVersionNumber, tableId)

	// Track if we need to clean up the version on failure
	versionCreated := true
	defer func() {
		if versionCreated {
			// If we created a version but something failed, clean it up
			log.Printf("Cleaning up version %d due to failure", newVersionNumber)
			if cleanupErr := cleanupVersion(ctx, proxy, tableId, newVersionNumber); cleanupErr != nil {
				log.Printf("Warning: Failed to cleanup version %d: %s", newVersionNumber, cleanupErr)
			}
		}
	}()

	// Step 2: Poll until version reaches draft status
	log.Printf("Waiting for version %d to reach draft status", newVersionNumber)
	err = waitForVersionDraftStatus(ctx, proxy, tableId, newVersionNumber)
	if err != nil {
		return fmt.Errorf("failed to wait for version to reach draft status: %s", err)
	}
	log.Printf("Version %d is now in draft status", newVersionNumber)

	// Step 3: Compare old vs new rows and determine changes
	changes := compareRows(oldRows, newRows)
	log.Printf("Detected changes: %d adds, %d updates, %d deletes", len(changes.adds), len(changes.updates), len(changes.deletes))

	// Step 4: Apply changes to the draft version
	err = applyRowChanges(ctx, proxy, tableId, newVersionNumber, changes)
	if err != nil {
		return fmt.Errorf("failed to apply row changes: %s", err)
	}

	// Step 5: Publish the updated version
	log.Printf("Publishing version %d", newVersionNumber)
	err = publishDecisionTableVersion(ctx, proxy, tableId, newVersionNumber)
	if err != nil {
		return fmt.Errorf("failed to publish version %d: %s", newVersionNumber, err)
	}

	// Success - mark version as no longer needing cleanup
	versionCreated = false
	log.Printf("Successfully updated rows for decision table %s version %d", tableId, newVersionNumber)
	return nil
}

// cleanupVersion deletes a version if it exists and is in draft status
func cleanupVersion(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, version int) error {
	log.Printf("Attempting to cleanup version %d for table %s", version, tableId)

	// First check if the version exists and is in draft status
	versionData, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, version)
	if err != nil {
		log.Printf("Could not get version %d status (may already be deleted): %s", version, err)
		return nil // Don't fail cleanup if version doesn't exist
	}

	// Only delete if it's in draft status
	if versionData.Status != nil && *versionData.Status == "Draft" {
		log.Printf("Deleting draft version %d", version)
		_, err := proxy.deleteDecisionTableVersion(ctx, tableId, version)
		if err != nil {
			log.Printf("Failed to delete version %d: %s", version, err)
			return err
		}
		log.Printf("Successfully deleted version %d", version)
	} else {
		log.Printf("Version %d is not in draft status (%s), skipping deletion", version, *versionData.Status)
	}

	return nil
}

// RowChange represents changes to be made to rows
type RowChange struct {
	adds    []map[string]interface{} // New rows to add
	updates []map[string]interface{} // Existing rows to update
	deletes []string                 // Row IDs to delete
}

// waitForVersionDraftStatus polls until a version reaches draft status
func waitForVersionDraftStatus(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, version int) error {
	const maxRetries = 30 // 15 minutes with 30-second intervals
	const retryInterval = 30 * time.Second

	for i := 0; i < maxRetries; i++ {
		versionData, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, version)
		if err != nil {
			return fmt.Errorf("failed to get version status: %s", err)
		}

		if versionData.Status != nil {
			status := *versionData.Status
			log.Printf("Version %d status: %s", version, status)

			if status == "Draft" {
				return nil
			}

			if status == "Failed" || status == "Error" {
				return fmt.Errorf("version %d failed with status: %s", version, status)
			}
		}

		// Wait before next check
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("timeout waiting for version %d to reach draft status", version)
}

// compareRows compares old and new rows to determine what changes need to be made
func compareRows(oldRows []interface{}, newRows []interface{}) RowChange {
	changes := RowChange{
		adds:    []map[string]interface{}{},
		updates: []map[string]interface{}{},
		deletes: []string{},
	}

	// Create maps for easier lookup
	oldRowsMap := make(map[string]map[string]interface{})
	for i, row := range oldRows {
		rowMap := row.(map[string]interface{})
		if rowId, ok := rowMap["row_id"].(string); ok && rowId != "" {
			log.Printf("DEBUG: Old row %d: ID=%s, data=%+v", i, rowId, rowMap)
			oldRowsMap[rowId] = rowMap
		}
	}

	newRowsMap := make(map[string]map[string]interface{})
	for i, row := range newRows {
		rowMap := row.(map[string]interface{})
		if rowId, ok := rowMap["row_id"].(string); ok && rowId != "" {
			log.Printf("DEBUG: New row %d: ID=%s, data=%+v", i, rowId, rowMap)
			newRowsMap[rowId] = rowMap
		} else {
			// New row without ID (will be added)
			log.Printf("DEBUG: New row %d: No ID, will be added: %+v", i, rowMap)
			changes.adds = append(changes.adds, rowMap)
		}
	}

	// Find updates and deletes
	for rowId, oldRow := range oldRowsMap {
		if newRow, exists := newRowsMap[rowId]; exists {
			// Row exists in both - check if it changed
			if !rowsEqual(oldRow, newRow) {
				log.Printf("DEBUG: Row %s detected as changed", rowId)
				log.Printf("DEBUG: Old row: %+v", oldRow)
				log.Printf("DEBUG: New row: %+v", newRow)
				changes.updates = append(changes.updates, newRow)
			} else {
				log.Printf("DEBUG: Row %s unchanged, skipping update", rowId)
			}
		} else {
			// Row was deleted
			changes.deletes = append(changes.deletes, rowId)
		}
	}

	return changes
}

// rowsEqual compares two row maps to see if they're equal
func rowsEqual(row1, row2 map[string]interface{}) bool {
	// Compare inputs - these are arrays in positional mapping
	inputs1, ok1 := row1["inputs"].([]interface{})
	inputs2, ok2 := row2["inputs"].([]interface{})
	if !ok1 || !ok2 || !arraysEqual(inputs1, inputs2) {
		log.Printf("DEBUG: Inputs differ - row1: %+v, row2: %+v", inputs1, inputs2)
		return false
	}

	// Compare outputs - these are arrays in positional mapping
	outputs1, ok1 := row1["outputs"].([]interface{})
	outputs2, ok2 := row2["outputs"].([]interface{})
	if !ok1 || !ok2 || !arraysEqual(outputs1, outputs2) {
		log.Printf("DEBUG: Outputs differ - row1: %+v, row2: %+v", outputs1, outputs2)
		return false
	}

	return true
}

// arraysEqual compares two arrays for equality
func arraysEqual(arr1, arr2 []interface{}) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i, value1 := range arr1 {
		value2 := arr2[i]
		if !valuesEqual(value1, value2) {
			return false
		}
	}

	return true
}

// mapsEqual compares two maps for equality
func mapsEqual(map1, map2 map[string]interface{}) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, value1 := range map1 {
		value2, exists := map2[key]
		if !exists || !valuesEqual(value1, value2) {
			return false
		}
	}

	return true
}

// valuesEqual compares two values for equality
func valuesEqual(val1, val2 interface{}) bool {
	// Handle different types
	switch v1 := val1.(type) {
	case map[string]interface{}:
		v2, ok := val2.(map[string]interface{})
		if !ok {
			return false
		}
		return mapsEqual(v1, v2)
	case []interface{}:
		v2, ok := val2.([]interface{})
		if !ok {
			return false
		}
		return arraysEqual(v1, v2)
	default:
		return val1 == val2
	}
}

// applyRowChanges applies the detected changes to the draft version
func applyRowChanges(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, version int, changes RowChange) error {
	// Get the table version to extract column mapping
	tableVersion, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, version)
	if err != nil {
		return fmt.Errorf("failed to get table version for column mapping: %s", err)
	}

	// Get column IDs in order for positional mapping
	inputColumnIds, outputColumnIds := extractColumnOrder(tableVersion.Columns)

	// Track successfully added rows for potential rollback
	var addedRows []string

	// Delete rows first
	for _, rowId := range changes.deletes {
		log.Printf("Deleting row %s", rowId)
		_, err := proxy.deleteDecisionTableRow(ctx, tableId, version, rowId)
		if err != nil {
			return fmt.Errorf("failed to delete row %s: %s", rowId, err)
		}
		log.Printf("Successfully deleted row %s", rowId)
	}

	// Update existing rows
	for _, row := range changes.updates {
		rowId := row["row_id"].(string)
		log.Printf("Updating row %s", rowId)

		// Convert to SDK format using positional mapping (same as creation)
		sdkRow, err := convertTerraformRowToSDK(row, inputColumnIds, outputColumnIds)
		if err != nil {
			return fmt.Errorf("failed to convert row for update: %s", err)
		}

		// Convert SDK row to update request format
		updateRequest := convertSDKRowToUpdateRequest(sdkRow)

		// Update the row
		updatedRow, _, err := proxy.updateDecisionTableRow(ctx, tableId, version, rowId, updateRequest)
		if err != nil {
			return fmt.Errorf("failed to update row %s: %s", rowId, err)
		}

		// Log the returned row data for debugging
		if updatedRow != nil {
			rowIdStr := "unknown"
			rowIndexStr := "unknown"
			if updatedRow.Id != nil {
				rowIdStr = *updatedRow.Id
			}
			if updatedRow.RowIndex != nil {
				rowIndexStr = fmt.Sprintf("%d", *updatedRow.RowIndex)
			}
			log.Printf("Successfully updated row %s: returned row_id=%s, row_index=%s",
				rowId, rowIdStr, rowIndexStr)
		} else {
			log.Printf("Successfully updated row %s (no row data returned)", rowId)
		}
	}

	// Add new rows using positional mapping
	for i, row := range changes.adds {
		log.Printf("Adding new row %d/%d", i+1, len(changes.adds))
		sdkRow, err := convertTerraformRowToSDK(row, inputColumnIds, outputColumnIds)
		if err != nil {
			return fmt.Errorf("failed to convert row %d: %s", i+1, err)
		}
		_, err = proxy.createDecisionTableRow(ctx, tableId, version, &sdkRow)
		if err != nil {
			// If adding a row fails, we can't easily rollback individual rows
			// The version cleanup will handle the overall rollback
			return fmt.Errorf("failed to add new row %d/%d: %s", i+1, len(changes.adds), err)
		}

		// Track successfully added rows (if we had row IDs, we'd store them here)
		addedRows = append(addedRows, fmt.Sprintf("row_%d", i+1))
		log.Printf("Successfully added row %d/%d", i+1, len(changes.adds))
	}

	log.Printf("Successfully applied all row changes: %d deletes, %d updates, %d adds", len(changes.deletes), len(changes.updates), len(addedRows))
	return nil
}
