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
	if description != "" {
		log.Printf("Creating business rules decision table with name: %s, description: %s", tableName, description)
	} else {
		log.Printf("Creating business rules decision table with name: %s", tableName)
	}

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
		log.Printf("Flattening columns for decision table %s version %d", tableId, versionToRead)
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
		log.Printf("Getting rows for decision table %s version %d", tableId, versionToRead)
		rows, err := getDecisionTableRows(ctx, proxy, tableVersion)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get rows for version %d: %s", versionToRead, err), nil))
		}
		log.Printf("Successfully retrieved %d rows for decision table %s version %d", len(rows), tableId, versionToRead)
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

	// Get column IDs in order for column order mapping
	inputColumnIds, outputColumnIds := extractColumnOrder(tableVersion.Columns)

	// Convert and add each row individually using column order mapping
	for i, row := range terraformRows {
		rowMap := row.(map[string]interface{})

		// Convert Terraform row to SDK format using column order mapping
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

	// Get column IDs in order for column order mapping
	inputColumnIds, outputColumnIds := extractColumnOrder(tableVersion.Columns)

	// Convert SDK rows to Terraform format using column order mapping
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

// waitForVersionDraftStatus polls until a version reaches draft status
func waitForVersionDraftStatus(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string, version int) error {
	const maxRetries = 30 // 15 minutes with 30-second intervals
	const retryInterval = 30 * time.Second

	log.Printf("Starting to poll for version %d to reach draft status (max %d retries, %v intervals)", version, maxRetries, retryInterval)

	for i := 0; i < maxRetries; i++ {
		log.Printf("Polling attempt %d/%d for version %d status", i+1, maxRetries, version)

		versionData, _, err := proxy.getBusinessRulesDecisionTableVersion(ctx, tableId, version)
		if err != nil {
			log.Printf("Failed to get version %d status on attempt %d: %s", version, i+1, err)
			return fmt.Errorf("failed to get version status: %s", err)
		}

		if versionData.Status != nil {
			status := *versionData.Status
			log.Printf("Version %d status on attempt %d: %s", version, i+1, status)

			if status == "Draft" {
				log.Printf("Version %d successfully reached draft status after %d attempts", version, i+1)
				return nil
			}

			if status == "Failed" || status == "Error" {
				log.Printf("Version %d failed with status: %s after %d attempts", version, status, i+1)
				return fmt.Errorf("version %d failed with status: %s", version, status)
			}
		} else {
			log.Printf("Version %d status is nil on attempt %d", version, i+1)
		}

		// Wait before next check
		if i < maxRetries-1 { // Don't sleep on the last iteration
			log.Printf("Waiting %v before next poll attempt for version %d", retryInterval, version)
			time.Sleep(retryInterval)
		}
	}

	log.Printf("Timeout reached after %d attempts waiting for version %d to reach draft status", maxRetries, version)
	return fmt.Errorf("timeout waiting for version %d to reach draft status", version)
}
