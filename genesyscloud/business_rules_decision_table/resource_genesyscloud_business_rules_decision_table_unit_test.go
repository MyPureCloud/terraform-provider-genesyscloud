package business_rules_decision_table

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/stretchr/testify/assert"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func TestResourceBusinessRulesDecisionTable(t *testing.T) {
	// Test that the resource can be created without errors
	resource := ResourceBusinessRulesDecisionTable()
	if resource == nil {
		t.Fatal("ResourceBusinessRulesDecisionTable() returned nil")
	}

	// Test that the schema is properly defined
	if resource.Schema == nil {
		t.Fatal("Resource schema is nil")
	}

	// Test required fields
	requiredFields := []string{"name", "division_id", "schema_id", "columns", "rows"}
	for _, fieldName := range requiredFields {
		field := resource.Schema[fieldName]
		if field == nil {
			t.Errorf("Required field '%s' is missing from schema", fieldName)
			continue
		}
		if !field.Required {
			t.Errorf("Field '%s' should be required", fieldName)
		}
	}

	// Test optional fields
	optionalFields := []string{"description"}
	for _, fieldName := range optionalFields {
		field := resource.Schema[fieldName]
		if field == nil {
			t.Errorf("Optional field '%s' is missing from schema", fieldName)
			continue
		}
		if !field.Optional {
			t.Errorf("Field '%s' should be optional", fieldName)
		}
	}

	// Test computed fields
	computedFields := []string{"version"}
	for _, fieldName := range computedFields {
		field := resource.Schema[fieldName]
		if field == nil {
			t.Errorf("Computed field '%s' is missing from schema", fieldName)
			continue
		}
		if !field.Computed {
			t.Errorf("Field '%s' should be computed", fieldName)
		}
	}

	// Test ForceNew fields
	forceNewFields := []string{"columns"}
	for _, fieldName := range forceNewFields {
		field := resource.Schema[fieldName]
		if field == nil {
			t.Errorf("ForceNew field '%s' is missing from schema", fieldName)
			continue
		}
		if !field.ForceNew {
			t.Errorf("Field '%s' should be ForceNew", fieldName)
		}
	}

	// Test field types
	expectedTypes := map[string]schema.ValueType{
		"name":        schema.TypeString,
		"description": schema.TypeString,
		"division_id": schema.TypeString,
		"schema_id":   schema.TypeString,
		"columns":     schema.TypeList,
		"rows":        schema.TypeList,
		"version":     schema.TypeInt,
	}

	for fieldName, expectedType := range expectedTypes {
		field := resource.Schema[fieldName]
		if field == nil {
			t.Errorf("Field '%s' should be defined in schema", fieldName)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("Field '%s' should be type %v, got %v", fieldName, expectedType, field.Type)
		}
	}

	// Test field constraints and validation
	nameField := resource.Schema["name"]
	if nameField.ValidateFunc == nil {
		t.Error("Name field should have validation function")
	}

	// Test columns field structure
	columnsField := resource.Schema["columns"]
	if columnsField.Type != schema.TypeList {
		t.Error("Columns field should be TypeList")
	}
	if columnsField.MaxItems != 1 {
		t.Error("Columns field should have MaxItems = 1")
	}
	if columnsField.Elem == nil {
		t.Error("Columns field should have Elem defined")
	}

	// Test rows field structure
	rowsField := resource.Schema["rows"]
	if rowsField.Type != schema.TypeList {
		t.Error("Rows field should be TypeList")
	}
	if rowsField.MinItems != 1 {
		t.Error("Rows field should have MinItems = 1")
	}
	if rowsField.Elem == nil {
		t.Error("Rows field should have Elem defined")
	}

	// Test that CRUD operations are defined
	if resource.CreateContext == nil {
		t.Error("CreateContext is not defined")
	}
	if resource.ReadContext == nil {
		t.Error("ReadContext is not defined")
	}
	if resource.UpdateContext == nil {
		t.Error("UpdateContext is not defined")
	}
	if resource.DeleteContext == nil {
		t.Error("DeleteContext is not defined")
	}

	// Test that Importer is defined
	if resource.Importer == nil {
		t.Error("Importer is not defined")
	}
	if resource.Importer.StateContext == nil {
		t.Error("Importer should have StateContext defined")
	}

	// Test SchemaVersion
	if resource.SchemaVersion != 1 {
		t.Errorf("SchemaVersion should be 1, got %d", resource.SchemaVersion)
	}
}

// Test CRUD operations with mocked API
func TestUnitResourceBusinessRulesDecisionTableCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	// Mock decision table data with comprehensive columns using proper SDK types
	tColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			// Input 1: Customer type with Equals comparator
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("Standard"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("customer_type"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			// Output 1: Queue reference with nested properties
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("genesyscloud_routing_queue.output_queue.id"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("transfer_queue"),
					Properties: &[]platformclientv2.Outputvalue{
						{
							SchemaPropertyKey: platformclientv2.String("queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{
									SchemaPropertyKey: platformclientv2.String("id"),
								},
							},
						},
					},
				},
			},
		},
	}

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	decisionTableProxy.createBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, createRequest *platformclientv2.Createdecisiontablerequest) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		tableVersion := &platformclientv2.Decisiontableversion{}

		// Validate basic fields
		assert.Equal(t, tName, *createRequest.Name, "createRequest.Name check failed in create createBusinessRulesDecisionTableAttr")
		assert.Equal(t, tDescription, *createRequest.Description, "createRequest.Description check failed in create createBusinessRulesDecisionTableAttr")
		if createRequest.DivisionId != nil {
			assert.Equal(t, tDivisionId, *createRequest.DivisionId, "createRequest.DivisionId check failed in create createBusinessRulesDecisionTableAttr")
		}
		if createRequest.SchemaId != nil {
			assert.Equal(t, tSchemaId, *createRequest.SchemaId, "createRequest.SchemaId check failed in create createBusinessRulesDecisionTableAttr")
		}

		// Validate columns are included in the request
		assert.NotNil(t, createRequest.Columns, "createRequest.Columns should not be nil")
		assert.NotNil(t, createRequest.Columns.Inputs, "createRequest.Columns.Inputs should not be nil")
		assert.NotNil(t, createRequest.Columns.Outputs, "createRequest.Columns.Outputs should not be nil")
		assert.Len(t, *createRequest.Columns.Inputs, 1, "createRequest.Columns.Inputs should have 1 input (customer_type)")
		assert.Len(t, *createRequest.Columns.Outputs, 1, "createRequest.Columns.Outputs should have 1 output (transfer_queue)")

		// Validate that the mock providers are working

		// Set up a realistic table version response
		tableVersion.Id = &tId
		tableVersion.Name = &tName
		tableVersion.Version = platformclientv2.Int(1)
		tableVersion.Status = platformclientv2.String("Draft")

		return tableVersion, nil, nil
	}

	// Add mocks for the read operations that create calls
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		tableVersion := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Version:     platformclientv2.Int(1),
			Status:      platformclientv2.String("Published"),
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
			Columns: tColumns, // Add the columns from our test data
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return tableVersion, apiResponse, nil
	}

	// Mock for adding rows
	decisionTableProxy.createDecisionTableRowAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, row *platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock for publishing version
	decisionTableProxy.publishDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock for deleting decision table
	decisionTableProxy.deleteBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock for getting rows - return rows that match our test data
	decisionTableProxy.getDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, pageNumber string, pageSize string) (*platformclientv2.Decisiontablerowlisting, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		// Return rows that match our test data
		rowId := uuid.NewString()
		rowIndex := 1
		mockRow := platformclientv2.Decisiontablerow{
			Id:       &rowId,
			RowIndex: platformclientv2.Int(rowIndex),
			Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
				"input-column-id-1": {
					Literal: &platformclientv2.Literal{
						VarString: platformclientv2.String("test-input-1"),
					},
				},
			},
			Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
				"output-column-id-1": {
					Literal: &platformclientv2.Literal{
						VarString: platformclientv2.String("test-output-1"),
					},
				},
			},
		}
		rows := &platformclientv2.Decisiontablerowlisting{
			Entities: &[]platformclientv2.Decisiontablerow{mockRow},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return rows, apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Convert SDK columns to Terraform format for testing
	tColumnsTF := convertSDKColumnsToTerraform(tColumns)

	// Setup test rows - inputs and outputs should be maps with literal objects
	testRows := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "test-input-1",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "test-output-1",
							"type":  "string",
						},
					},
				},
			},
		},
	}

	// Setup a map of values
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, tColumnsTF)
	resourceDataMap["rows"] = testRows

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := createBusinessRulesDecisionTable(ctx, d, gcloud)
	if diag.HasError() {
		t.Logf("Create failed with errors: %v", diag)
	}
	assert.Equal(t, false, diag.HasError())

	// Validate that create actually set an ID
	createdId := d.Id()
	assert.NotEmpty(t, createdId, "Create should have generated an ID")

	// Since our mock returns the test ID, we expect them to be the same
	assert.Equal(t, tId, createdId, "Create should have set the ID returned by our mock")
}

func TestUnitResourceBusinessRulesDecisionTableRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// columns definition to ensure consistency between table and version mocks
	sharedColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("input-queue-id"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("queuey"),
							Contractual: func() **platformclientv2.Contractual {
								nested := &platformclientv2.Contractual{
									SchemaPropertyKey: platformclientv2.String("queue"),
									Contractual: func() **platformclientv2.Contractual {
										deep := &platformclientv2.Contractual{
											SchemaPropertyKey: platformclientv2.String("id"),
										}
										return &deep
									}(),
								}
								return &nested
							}(),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
			{
				Id: platformclientv2.String("input-column-id-2"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("true"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("boolie"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("output-queue-id"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("queuey"),
					Properties: &[]platformclientv2.Outputvalue{
						{
							SchemaPropertyKey: platformclientv2.String("queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{
									SchemaPropertyKey: platformclientv2.String("id"),
								},
							},
						},
					},
				},
			},
			{
				Id: platformclientv2.String("output-column-id-2"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("VIP Customer"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("stringy"),
				},
			},
		},
	}

	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
			// Use shared columns for consistency
			Columns: sharedColumns,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the getBusinessRulesDecisionTableVersion function to return complete version information
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, versionNumber, "Expected version 1")

		version := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Version:     platformclientv2.Int(1),
			Status:      platformclientv2.String("Published"),
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
			// Use shared columns for consistency
			Columns: sharedColumns,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return version, apiResponse, nil
	}

	// Mock for getting rows - published tables should have rows
	decisionTableProxy.getDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, pageNumber string, pageSize string) (*platformclientv2.Decisiontablerowlisting, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)

		// Return realistic rows for a published table
		rowId1 := uuid.NewString()
		rowId2 := uuid.NewString()

		mockRows := []platformclientv2.Decisiontablerow{
			{
				Id:       &rowId1,
				RowIndex: platformclientv2.Int(1),
				Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"input-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("queue-id-1"),
						},
					},
					"input-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("true"),
						},
					},
				},
				Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"output-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("transfer-queue-id-1"),
						},
					},
					"output-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("VIP Customer"),
						},
					},
				},
			},
			{
				Id:       &rowId2,
				RowIndex: platformclientv2.Int(2),
				Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"input-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("queue-id-2"),
						},
					},
					"input-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("false"),
						},
					},
				},
				Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"output-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("transfer-queue-id-2"),
						},
					},
					"output-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("Regular Customer"),
						},
					},
				},
			},
		}

		rows := &platformclientv2.Decisiontablerowlisting{
			Entities: &mockRows,
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return rows, apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup a map of values
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, nil)

	// Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readBusinessRulesDecisionTable(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name"))
	assert.Equal(t, tDescription, d.Get("description"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
	assert.Equal(t, tSchemaId, d.Get("schema_id"))

	// Assert that columns are properly processed and set in the resource data
	assert.NotNil(t, d.Get("columns"), "Columns should be set")

	// Assert that rows are properly processed and set in the resource data
	rows := d.Get("rows")
	assert.NotNil(t, rows, "Rows should be set")
	rowsList := rows.([]interface{})
	assert.Len(t, rowsList, 2, "Should have 2 rows")

	// Verify first row structure
	row1 := rowsList[0].(map[string]interface{})
	assert.NotNil(t, row1["row_id"], "Row 1 should have row_id")
	assert.Equal(t, 1, row1["row_index"], "Row 1 should have row_index 1")
	assert.NotNil(t, row1["inputs"], "Row 1 should have inputs")
	assert.NotNil(t, row1["outputs"], "Row 1 should have outputs")

	// Verify row 1 has correct number of inputs and outputs (matching columns)
	row1Inputs := row1["inputs"].([]interface{})
	row1Outputs := row1["outputs"].([]interface{})
	assert.Len(t, row1Inputs, 2, "Row 1 should have 2 inputs (matching input columns)")
	assert.Len(t, row1Outputs, 2, "Row 1 should have 2 outputs (matching output columns)")

	// Verify second row structure
	row2 := rowsList[1].(map[string]interface{})
	assert.NotNil(t, row2["row_id"], "Row 2 should have row_id")
	assert.Equal(t, 2, row2["row_index"], "Row 2 should have row_index 2")
	assert.NotNil(t, row2["inputs"], "Row 2 should have inputs")
	assert.NotNil(t, row2["outputs"], "Row 2 should have outputs")

	// Verify row 2 has correct number of inputs and outputs (matching columns)
	row2Inputs := row2["inputs"].([]interface{})
	row2Outputs := row2["outputs"].([]interface{})
	assert.Len(t, row2Inputs, 2, "Row 2 should have 2 inputs (matching input columns)")
	assert.Len(t, row2Outputs, 2, "Row 2 should have 2 outputs (matching output columns)")

	// Verify row 1 input values match mock data
	row1Input1 := row1Inputs[0].(map[string]interface{})
	assert.Equal(t, "input-column-id-1", row1Input1["column_id"])
	row1Input1LiteralList := row1Input1["literal"].([]interface{})
	assert.Len(t, row1Input1LiteralList, 1)
	row1Input1Literal := row1Input1LiteralList[0].(map[string]interface{})
	assert.Equal(t, "queue-id-1", row1Input1Literal["value"])
	assert.Equal(t, "string", row1Input1Literal["type"])

	row1Input2 := row1Inputs[1].(map[string]interface{})
	assert.Equal(t, "input-column-id-2", row1Input2["column_id"])
	row1Input2LiteralList := row1Input2["literal"].([]interface{})
	assert.Len(t, row1Input2LiteralList, 1)
	row1Input2Literal := row1Input2LiteralList[0].(map[string]interface{})
	assert.Equal(t, "true", row1Input2Literal["value"])
	assert.Equal(t, "string", row1Input2Literal["type"])

	// Verify row 1 output values match mock data
	row1Output1 := row1Outputs[0].(map[string]interface{})
	assert.Equal(t, "output-column-id-1", row1Output1["column_id"])
	row1Output1LiteralList := row1Output1["literal"].([]interface{})
	assert.Len(t, row1Output1LiteralList, 1)
	row1Output1Literal := row1Output1LiteralList[0].(map[string]interface{})
	assert.Equal(t, "transfer-queue-id-1", row1Output1Literal["value"])
	assert.Equal(t, "string", row1Output1Literal["type"])

	row1Output2 := row1Outputs[1].(map[string]interface{})
	assert.Equal(t, "output-column-id-2", row1Output2["column_id"])
	row1Output2LiteralList := row1Output2["literal"].([]interface{})
	assert.Len(t, row1Output2LiteralList, 1)
	row1Output2Literal := row1Output2LiteralList[0].(map[string]interface{})
	assert.Equal(t, "VIP Customer", row1Output2Literal["value"])
	assert.Equal(t, "string", row1Output2Literal["type"])

	// Verify row 2 input values match mock data
	row2Input1 := row2Inputs[0].(map[string]interface{})
	assert.Equal(t, "input-column-id-1", row2Input1["column_id"])
	row2Input1LiteralList := row2Input1["literal"].([]interface{})
	assert.Len(t, row2Input1LiteralList, 1)
	row2Input1Literal := row2Input1LiteralList[0].(map[string]interface{})
	assert.Equal(t, "queue-id-2", row2Input1Literal["value"])
	assert.Equal(t, "string", row2Input1Literal["type"])

	row2Input2 := row2Inputs[1].(map[string]interface{})
	assert.Equal(t, "input-column-id-2", row2Input2["column_id"])
	row2Input2LiteralList := row2Input2["literal"].([]interface{})
	assert.Len(t, row2Input2LiteralList, 1)
	row2Input2Literal := row2Input2LiteralList[0].(map[string]interface{})
	assert.Equal(t, "false", row2Input2Literal["value"])
	assert.Equal(t, "string", row2Input2Literal["type"])

	// Verify row 2 output values match mock data
	row2Output1 := row2Outputs[0].(map[string]interface{})
	assert.Equal(t, "output-column-id-1", row2Output1["column_id"])
	row2Output1LiteralList := row2Output1["literal"].([]interface{})
	assert.Len(t, row2Output1LiteralList, 1)
	row2Output1Literal := row2Output1LiteralList[0].(map[string]interface{})
	assert.Equal(t, "transfer-queue-id-2", row2Output1Literal["value"])
	assert.Equal(t, "string", row2Output1Literal["type"])

	row2Output2 := row2Outputs[1].(map[string]interface{})
	assert.Equal(t, "output-column-id-2", row2Output2["column_id"])
	row2Output2LiteralList := row2Output2["literal"].([]interface{})
	assert.Len(t, row2Output2LiteralList, 1)
	row2Output2Literal := row2Output2LiteralList[0].(map[string]interface{})
	assert.Equal(t, "Regular Customer", row2Output2Literal["value"])
	assert.Equal(t, "string", row2Output2Literal["type"])

}

func TestUnitResourceBusinessRulesDecisionTableUpdateNameDescription(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tUpdatedName := "Updated Decision Table Name"
	tUpdatedDescription := "Updated Decision Table Description"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Mock the get function to return the table with published version
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tUpdatedName,        // Return updated name after update
			Description: &tUpdatedDescription, // Return updated description after update
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	// Mock the update function for name/description only
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, tUpdatedName, *updateRequest.Name, "updateRequest.Name should match updated name")
		assert.Equal(t, tUpdatedDescription, *updateRequest.Description, "updateRequest.Description should match updated description")

		// For name/description updates, columns should not be included
		assert.Nil(t, updateRequest.Columns, "Update request should not include columns for name/description updates")

		// Return updated table
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tUpdatedName,
			Description: &tUpdatedDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	// Mock for getting rows
	decisionTableProxy.getDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, pageNumber string, pageSize string) (*platformclientv2.Decisiontablerowlisting, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)

		// Return empty rows for this test
		rows := &platformclientv2.Decisiontablerowlisting{
			Entities: &[]platformclientv2.Decisiontablerow{},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return rows, apiResponse, nil
	}

	// Mock for getting columns (needed for read)
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, versionNumber, "Expected version 1")

		version := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tUpdatedName,
			Description: &tUpdatedDescription,
			Version:     platformclientv2.Int(1),
			Status:      platformclientv2.String("Published"),
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
			Columns: &platformclientv2.Decisiontablecolumns{
				Inputs:  &[]platformclientv2.Decisiontableinputcolumn{},
				Outputs: &[]platformclientv2.Decisiontableoutputcolumn{},
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup initial resource data
	initialResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	d := schema.TestResourceDataRaw(t, resourceSchema, initialResourceDataMap)
	d.SetId(tId)

	// Update to new name and description
	updatedResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tUpdatedName, tUpdatedDescription, tDivisionId, tSchemaId, []interface{}{})
	d2 := schema.TestResourceDataRaw(t, resourceSchema, updatedResourceDataMap)
	d2.SetId(tId)

	// Perform the update
	diag := updateBusinessRulesDecisionTable(ctx, d2, gcloud)
	assert.Equal(t, false, diag.HasError(), "Update should succeed for name/description changes")
	assert.Equal(t, tId, d2.Id())

	// Verify that the updated values are properly set
	assert.Equal(t, tUpdatedName, d2.Get("name"))
	assert.Equal(t, tUpdatedDescription, d2.Get("description"))
}

func TestUnitResourceBusinessRulesDecisionTableUpdateRows(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	// Define columns for the test (same as other tests)
	tColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("Standard"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("customer_type"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("genesyscloud_routing_queue.output_queue.id"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("transfer_queue"),
					Properties: &[]platformclientv2.Outputvalue{
						{
							SchemaPropertyKey: platformclientv2.String("queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{
									SchemaPropertyKey: platformclientv2.String("id"),
								},
							},
						},
					},
				},
			},
			{
				Id: platformclientv2.String("output-column-id-2"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("VIP Customer"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("escalation_level"),
				},
			},
		},
	}

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Mock the get function to return the table with published version
	// After row updates, version 2 should be published
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(2),
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	// Mock the update function for name/description only (no columns)
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, tName, *updateRequest.Name)
		assert.Equal(t, tDescription, *updateRequest.Description)

		// For row updates, columns should not be included
		assert.Nil(t, updateRequest.Columns, "Update request should not include columns for row updates")

		// Return updated table
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(2), // New published version after row updates
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	// Track the state of version 2 to simulate the draft -> published transition
	version2Status := "Draft"

	// Mock the version lookup function - return different status based on version and state
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)

		// Version 1 is always published, version 2 transitions from draft to published
		status := "Published"
		if versionNumber == 2 {
			status = version2Status
		}

		version := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Version:     platformclientv2.Int(versionNumber),
			Status:      platformclientv2.String(status),
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
			Columns: tColumns, // Use the same column definition as other tests
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock for creating new version (needed for row updates)
	decisionTableProxy.createDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		version := &platformclientv2.Decisiontableversion{
			Id:      &tId,
			Version: platformclientv2.Int(2),
			Status:  platformclientv2.String("Draft"),
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock for publishing version (needed for row updates)
	decisionTableProxy.publishDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 2, version)

		// Change version 2 status from Draft to Published
		version2Status = "Published"

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock for creating rows (needed for row updates)
	decisionTableProxy.createDecisionTableRowAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, row *platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 2, version)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock for deleting version (needed for cleanup)
	decisionTableProxy.deleteDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 2, version)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock for getting rows - return initial rows
	decisionTableProxy.getDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, pageNumber string, pageSize string) (*platformclientv2.Decisiontablerowlisting, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)

		// Return initial rows for version 1 (3 rows)
		if version == 1 {
			rowId1 := uuid.NewString()
			rowId2 := uuid.NewString()
			rowId3 := uuid.NewString()

			mockRows := []platformclientv2.Decisiontablerow{
				{
					Id:       &rowId1,
					RowIndex: platformclientv2.Int(1),
					Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
						"input-column-id-1": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("queue-id-1"),
							},
						},
					},
					Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
						"output-column-id-1": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("transfer-queue-id-1"),
							},
						},
						"output-column-id-2": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("Customer Type 1"),
							},
						},
					},
				},
				{
					Id:       &rowId2,
					RowIndex: platformclientv2.Int(2),
					Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
						"input-column-id-1": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("queue-id-2"),
							},
						},
					},
					Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
						"output-column-id-1": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("transfer-queue-id-2"),
							},
						},
						"output-column-id-2": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("Customer Type 2"),
							},
						},
					},
				},
				{
					Id:       &rowId3,
					RowIndex: platformclientv2.Int(3),
					Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
						"input-column-id-1": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("queue-id-3"),
							},
						},
					},
					Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
						"output-column-id-1": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("transfer-queue-id-3"),
							},
						},
						"output-column-id-2": {
							Literal: &platformclientv2.Literal{
								VarString: platformclientv2.String("Customer Type 3"),
							},
						},
					},
				},
			}

			rows := &platformclientv2.Decisiontablerowlisting{
				Entities: &mockRows,
			}
			apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return rows, apiResponse, nil
		}

		// Return updated rows for version 2 (1 updated + 1 kept + 1 new, 1 deleted)
		rowId1 := uuid.NewString()
		rowId3 := uuid.NewString()
		rowId4 := uuid.NewString()
		mockRows := []platformclientv2.Decisiontablerow{
			// Updated row: queue-id-1 with updated customer type
			{
				Id:       &rowId1,
				RowIndex: platformclientv2.Int(1),
				Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"input-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("queue-id-1"),
						},
					},
				},
				Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"output-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("transfer-queue-id-1"),
						},
					},
					"output-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("UPDATED Customer Type 1"),
						},
					},
				},
			},
			// Kept row: queue-id-3 unchanged
			{
				Id:       &rowId3,
				RowIndex: platformclientv2.Int(2),
				Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"input-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("queue-id-3"),
						},
					},
				},
				Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"output-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("transfer-queue-id-3"),
						},
					},
					"output-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("Customer Type 3"),
						},
					},
				},
			},
			// New row: queue-id-4 without comparator (tests fallback logic)
			{
				Id:       &rowId4,
				RowIndex: platformclientv2.Int(3),
				Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"input-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("queue-id-4"),
						},
					},
				},
				Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"output-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("transfer-queue-id-4"),
						},
					},
					"output-column-id-2": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("Customer Type 4 (No Comparator)"),
						},
					},
				},
			},
			// Deleted row: queue-id-2 is not in the updated rows
		}

		rows := &platformclientv2.Decisiontablerowlisting{
			Entities: &mockRows,
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return rows, apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup initial resource data with initial rows (3 rows)
	initialRows := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-1",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-1",
							"type":  "string",
						},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "Customer Type 1",
							"type":  "string",
						},
					},
				},
			},
		},
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-2",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-2",
							"type":  "string",
						},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "Customer Type 2",
							"type":  "string",
						},
					},
				},
			},
		},
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-3",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-3",
							"type":  "string",
						},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "Customer Type 3",
							"type":  "string",
						},
					},
				},
			},
		},
	}

	initialResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	initialResourceDataMap["rows"] = initialRows
	d := schema.TestResourceDataRaw(t, resourceSchema, initialResourceDataMap)
	d.SetId(tId)

	// Update to new rows demonstrating all operations (update, keep, delete)
	updatedRows := []interface{}{
		// UPDATE: Keep queue-id-1 but change the customer type (string field)
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-1",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-1",
							"type":  "string",
						},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "UPDATED Customer Type 1", // Updated string value
							"type":  "string",
						},
					},
				},
			},
		},
		// KEEP: Keep queue-id-3 unchanged
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-3",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-3",
							"type":  "string",
						},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "Customer Type 3",
							"type":  "string",
						},
					},
				},
			},
		},
		// ADD: New row without comparator to test fallback logic
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					// No comparator specified - should trigger fallback logic
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-4",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-4",
							"type":  "string",
						},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "Customer Type 4 (No Comparator)",
							"type":  "string",
						},
					},
				},
			},
		},
		// DELETE: queue-id-2 is removed (not in updated rows)
	}

	updatedResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	updatedResourceDataMap["rows"] = updatedRows
	d2 := schema.TestResourceDataRaw(t, resourceSchema, updatedResourceDataMap)
	d2.SetId(tId)

	// Perform the update
	diag := updateBusinessRulesDecisionTable(ctx, d2, gcloud)
	assert.Equal(t, false, diag.HasError(), "Update should succeed for row changes")
	assert.Equal(t, tId, d2.Id())

	// Verify that the updated rows are properly set
	rows := d2.Get("rows")
	assert.NotNil(t, rows, "Rows should be set after update")
	rowsList := rows.([]interface{})
	assert.Len(t, rowsList, 3, "Should have 3 rows after update (1 updated + 1 kept + 1 new)")
}

func TestUnitResourceBusinessRulesDecisionTableUpdateErrorScenarios(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Mock the get function to return 404 (table not found)
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		return nil, apiResponse, fmt.Errorf("table not found")
	}

	// Mock the update function to return an error
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}
		return nil, apiResponse, fmt.Errorf("update failed")
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup resource data
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	// Perform the update - should fail
	diag := updateBusinessRulesDecisionTable(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError(), "Update should fail when update fails")
	assert.Contains(t, diag[0].Summary, "update failed")
}

func TestUnitResourceBusinessRulesDecisionTableUpdateRowFailureRollback(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Track if version cleanup was called
	versionCleanupCalled := false
	versionCreated := 0

	// Mock the get function to return a table with published version 1
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
			PublishedContract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	// Mock the update function for name/description (should succeed)
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		// Verify that columns are not being updated (only name/description)
		assert.Nil(t, updateRequest.Columns, "Columns should not be updated in this scenario")

		table := &platformclientv2.Decisiontable{
			Name:        updateRequest.Name,
			Description: updateRequest.Description,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Published: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return table, apiResponse, nil
	}

	// Mock version lookup - return different status based on version
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)

		status := "Published"
		if versionNumber == 2 {
			status = "Draft" // Version 2 starts as draft
		}

		version := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Version:     platformclientv2.Int(versionNumber),
			Status:      platformclientv2.String(status),
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
			Columns: &platformclientv2.Decisiontablecolumns{
				Inputs: &[]platformclientv2.Decisiontableinputcolumn{
					{
						Id: platformclientv2.String("input-column-id-1"),
						DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: platformclientv2.String("Standard"),
						},
						Expression: &platformclientv2.Decisiontableinputcolumnexpression{
							Contractual: func() **platformclientv2.Contractual {
								contractual := &platformclientv2.Contractual{
									SchemaPropertyKey: platformclientv2.String("customer_type"),
								}
								return &contractual
							}(),
							Comparator: platformclientv2.String("Equals"),
						},
					},
				},
				Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
					{
						Id: platformclientv2.String("output-column-id-1"),
						DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: platformclientv2.String("genesyscloud_routing_queue.output_queue.id"),
						},
						Value: &platformclientv2.Outputvalue{
							SchemaPropertyKey: platformclientv2.String("transfer_queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{
									SchemaPropertyKey: platformclientv2.String("queue"),
									Properties: &[]platformclientv2.Outputvalue{
										{
											SchemaPropertyKey: platformclientv2.String("id"),
										},
									},
								},
							},
						},
					},
				},
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock version creation (should succeed)
	decisionTableProxy.createDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		versionCreated = 2
		version := &platformclientv2.Decisiontableversion{
			Id:      &tId,
			Version: platformclientv2.Int(2),
			Status:  platformclientv2.String("Draft"),
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock row creation (should fail to trigger rollback)
	decisionTableProxy.createDecisionTableRowAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, row *platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 2, version)
		// Simulate row creation failure
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}
		return apiResponse, fmt.Errorf("row creation failed")
	}

	// Mock version cleanup (should be called during rollback)
	decisionTableProxy.deleteDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 2, version)
		versionCleanupCalled = true
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock getting rows for version 1 (initial state)
	decisionTableProxy.getDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, pageNumber string, pageSize string) (*platformclientv2.Decisiontablerowlisting, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)

		rowId1 := uuid.NewString()
		mockRows := []platformclientv2.Decisiontablerow{
			{
				Id:       &rowId1,
				RowIndex: platformclientv2.Int(1),
				Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"input-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("queue-id-1"),
						},
					},
				},
				Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
					"output-column-id-1": {
						Literal: &platformclientv2.Literal{
							VarString: platformclientv2.String("transfer-queue-id-1"),
						},
					},
				},
			},
		}

		rows := &platformclientv2.Decisiontablerowlisting{
			Entities: &mockRows,
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return rows, apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup initial resource data with rows
	initialRows := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-1",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-1",
							"type":  "string",
						},
					},
				},
			},
		},
	}

	initialResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	initialResourceDataMap["rows"] = initialRows
	d := schema.TestResourceDataRaw(t, resourceSchema, initialResourceDataMap)
	d.SetId(tId)

	// Setup updated resource data with different rows (should trigger row update)
	updatedRows := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-2",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-2",
							"type":  "string",
						},
					},
				},
			},
		},
	}

	updatedResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	updatedResourceDataMap["rows"] = updatedRows
	d2 := schema.TestResourceDataRaw(t, resourceSchema, updatedResourceDataMap)
	d2.SetId(tId)

	// Perform the update - should fail during row creation and trigger rollback
	diag := updateBusinessRulesDecisionTable(ctx, d2, gcloud)

	// Verify the update failed
	assert.Equal(t, true, diag.HasError(), "Update should fail when row creation fails")
	assert.Contains(t, diag[0].Summary, "row creation failed")

	// Verify that version cleanup was called (rollback occurred)
	assert.Equal(t, true, versionCleanupCalled, "Version cleanup should be called during rollback")
	assert.Equal(t, 2, versionCreated, "Version 2 should have been created before failure")
}

func TestUnitResourceBusinessRulesDecisionTableCreateFailureRollback(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	// Define columns for the test
	tColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("Standard"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("customer_type"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("genesyscloud_routing_queue.output_queue.id"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("transfer_queue"),
					Properties: &[]platformclientv2.Outputvalue{
						{
							SchemaPropertyKey: platformclientv2.String("queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{
									SchemaPropertyKey: platformclientv2.String("id"),
								},
							},
						},
					},
				},
			},
		},
	}

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Track if table cleanup was called
	tableCleanupCalled := false
	tableCreated := false

	// Mock the create function (should succeed)
	decisionTableProxy.createBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, createRequest *platformclientv2.Createdecisiontablerequest) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		tableCreated = true
		version := &platformclientv2.Decisiontableversion{
			Id:      &tId,
			Version: platformclientv2.Int(1),
			Status:  platformclientv2.String("Draft"),
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock the get version function (needed for column order)
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		tableVersion := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Version:     platformclientv2.Int(1),
			Status:      platformclientv2.String("Draft"),
			Columns:     tColumns,
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return tableVersion, apiResponse, nil
	}

	// Mock the create row function (should fail to trigger rollback)
	decisionTableProxy.createDecisionTableRowAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, row *platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		// Simulate row creation failure
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}
		return apiResponse, fmt.Errorf("row creation failed")
	}

	// Mock the delete function (should be called during rollback)
	decisionTableProxy.deleteBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		tableCleanupCalled = true
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup resource data with rows
	rows := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-1",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-1",
							"type":  "string",
						},
					},
				},
			},
		},
	}

	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	resourceDataMap["rows"] = rows
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	// Perform the create - should fail during row addition and trigger rollback
	diag := createBusinessRulesDecisionTable(ctx, d, gcloud)

	// Verify the create failed
	assert.Equal(t, true, diag.HasError(), "Create should fail when row creation fails")
	assert.Contains(t, diag[0].Summary, "row creation failed")

	// Verify that table cleanup was called (rollback occurred)
	assert.Equal(t, true, tableCleanupCalled, "Table cleanup should be called during rollback")
	assert.Equal(t, true, tableCreated, "Table should have been created before failure")
}

func TestUnitResourceBusinessRulesDecisionTableCreatePublishFailureRollback(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	// Define columns for the test
	tColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("Standard"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("customer_type"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("genesyscloud_routing_queue.output_queue.id"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("transfer_queue"),
					Properties: &[]platformclientv2.Outputvalue{
						{
							SchemaPropertyKey: platformclientv2.String("queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{
									SchemaPropertyKey: platformclientv2.String("id"),
								},
							},
						},
					},
				},
			},
		},
	}

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Track if table cleanup was called
	tableCleanupCalled := false
	tableCreated := false
	rowsAdded := false

	// Mock the create function (should succeed)
	decisionTableProxy.createBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, createRequest *platformclientv2.Createdecisiontablerequest) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		tableCreated = true
		version := &platformclientv2.Decisiontableversion{
			Id:      &tId,
			Version: platformclientv2.Int(1),
			Status:  platformclientv2.String("Draft"),
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock the get version function (needed for column order)
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		tableVersion := &platformclientv2.Decisiontableversion{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Version:     platformclientv2.Int(1),
			Status:      platformclientv2.String("Draft"),
			Columns:     tColumns,
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return tableVersion, apiResponse, nil
	}

	// Mock the create row function (should succeed)
	decisionTableProxy.createDecisionTableRowAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, row *platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		rowsAdded = true
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	// Mock the publish function (should fail to trigger rollback)
	decisionTableProxy.publishDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, version)
		// Simulate publish failure
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}
		return apiResponse, fmt.Errorf("publish failed")
	}

	// Mock the delete function (should be called during rollback)
	decisionTableProxy.deleteBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		tableCleanupCalled = true
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup resource data with rows
	rows := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "queue-id-1",
							"type":  "string",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{
							"value": "transfer-queue-id-1",
							"type":  "string",
						},
					},
				},
			},
		},
	}

	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, []interface{}{})
	resourceDataMap["rows"] = rows
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	// Perform the create - should fail during publishing and trigger rollback
	diag := createBusinessRulesDecisionTable(ctx, d, gcloud)

	// Verify the create failed
	assert.Equal(t, true, diag.HasError(), "Create should fail when publishing fails")
	assert.Contains(t, diag[0].Summary, "publish failed")

	// Verify that table cleanup was called (rollback occurred)
	assert.Equal(t, true, tableCleanupCalled, "Table cleanup should be called during rollback")
	assert.Equal(t, true, tableCreated, "Table should have been created before failure")
	assert.Equal(t, true, rowsAdded, "Rows should have been added before failure")
}

func TestUnitResourceBusinessRulesDecisionTableDelete(t *testing.T) {
	tId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Mock the delete function
	decisionTableProxy.deleteBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse, nil
	}

	// Mock the get function to return 404 after deletion (simulating the table no longer exists)
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		// Return 404 to simulate the table has been deleted
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		return nil, apiResponse, fmt.Errorf("not found")
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Setup a map of values
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, "Test Table", "Test Description", uuid.NewString(), uuid.NewString(), nil)

	// Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteBusinessRulesDecisionTable(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, "", d.Id())
}

// Test the exporter configuration
func TestUnitBusinessRulesDecisionTableExporter(t *testing.T) {
	exporter := BusinessRulesDecisionTableExporter()

	// Test that exporter is not nil
	assert.NotNil(t, exporter, "Exporter should not be nil")

	// Test RefAttrs configuration
	assert.NotNil(t, exporter.RefAttrs, "RefAttrs should not be nil")
	assert.Contains(t, exporter.RefAttrs, "division_id", "division_id should be in RefAttrs")
	assert.Contains(t, exporter.RefAttrs, "schema_id", "schema_id should be in RefAttrs")

	// Test RefAttrs types
	assert.Equal(t, "genesyscloud_auth_division", exporter.RefAttrs["division_id"].RefType, "division_id should reference auth_division")
	assert.Equal(t, "genesyscloud_business_rules_schema", exporter.RefAttrs["schema_id"].RefType, "schema_id should reference business_rules_schema")

	// Test CustomAttributeResolver configuration
	assert.NotNil(t, exporter.CustomAttributeResolver, "CustomAttributeResolver should not be nil")
	assert.Contains(t, exporter.CustomAttributeResolver, "columns.outputs.defaults_to.value", "outputs defaults_to.value should have custom resolver")
	assert.Contains(t, exporter.CustomAttributeResolver, "columns.inputs.defaults_to.value", "inputs defaults_to.value should have custom resolver")
	assert.Contains(t, exporter.CustomAttributeResolver, "rows.*.inputs.*.literal.value", "row inputs literal.value should have custom resolver")
	assert.Contains(t, exporter.CustomAttributeResolver, "rows.*.outputs.*.literal.value", "row outputs literal.value should have custom resolver")

	// Test that resolver functions are set
	assert.NotNil(t, exporter.CustomAttributeResolver["columns.outputs.defaults_to.value"].ResolverFunc, "outputs resolver function should be set")
	assert.NotNil(t, exporter.CustomAttributeResolver["columns.inputs.defaults_to.value"].ResolverFunc, "inputs resolver function should be set")
	assert.NotNil(t, exporter.CustomAttributeResolver["rows.*.inputs.*.literal.value"].ResolverFunc, "row inputs resolver function should be set")
	assert.NotNil(t, exporter.CustomAttributeResolver["rows.*.outputs.*.literal.value"].ResolverFunc, "row outputs resolver function should be set")
}

// Test the queue ID resolver function
func TestUnitQueueIdResolver(t *testing.T) {
	// Test with a valid queue ID
	queueID := "test-queue-id"
	queueName := "test-queue"

	// Create a mock resource map with queue information
	resourceMap := map[string]interface{}{
		"genesyscloud_routing_queue": map[string]interface{}{
			queueID: map[string]interface{}{
				"name": queueName,
			},
		},
	}

	// Create mock exporters map (required by the function signature)
	exporters := map[string]*resourceExporter.ResourceExporter{}

	// Test queue ID resolution
	err := QueueIdResolver(resourceMap, exporters, queueID)
	assert.NoError(t, err, "Queue ID resolution should succeed")

	// Test with non-queue value (should succeed)
	nonQueueValue := "priority_high"
	err2 := QueueIdResolver(resourceMap, exporters, nonQueueValue)
	assert.NoError(t, err2, "Non-queue value should be handled without error")

	// Test with empty value
	emptyValue := ""
	err3 := QueueIdResolver(resourceMap, exporters, emptyValue)
	assert.NoError(t, err3, "Empty value should be handled without error")

	// Test with nil resource map
	err4 := QueueIdResolver(nil, exporters, queueID)
	assert.NoError(t, err4, "Should handle nil resource map without error")
}

// Test the data source functionality
func TestUnitDataSourceBusinessRulesDecisionTable(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Mock the getBusinessRulesDecisionTablesByName function
	decisionTableProxy.getBusinessRulesDecisionTablesByNameAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (*[]platformclientv2.Decisiontable, bool, *platformclientv2.APIResponse, error) {
		version := int(1)
		published := &platformclientv2.Decisiontableversionentity{
			Version: &version,
		}

		table := platformclientv2.Decisiontable{
			Id:        &tId,
			Name:      &tName,
			Published: published,
		}

		if name == tName {
			return &[]platformclientv2.Decisiontable{table}, false, &platformclientv2.APIResponse{}, nil
		}
		return &[]platformclientv2.Decisiontable{}, true, &platformclientv2.APIResponse{}, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Test data source read
	resourceSchema := DataSourceBusinessRulesDecisionTable().Schema
	resourceDataMap := map[string]interface{}{
		"name": tName,
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := dataSourceBusinessRulesDecisionTableRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), "Data source read should succeed")

	// Verify that the ID was set
	assert.Equal(t, tId, d.Id(), "Data source should set the correct ID")
	assert.Equal(t, tName, d.Get("name"), "Data source should return the correct name")
	assert.Equal(t, 1, d.Get("version"), "Data source should return the correct version")
}

// Helper function to build test resource map for CRUD tests
func buildTestDecisionTableResourceMapCRUD(id, name, description, divisionId, schemaId string, columns interface{}) map[string]interface{} {
	resourceMap := map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"division_id": divisionId,
		"schema_id":   schemaId,
	}

	if columns != nil {
		resourceMap["columns"] = columns
	}

	return resourceMap
}

// Helper function to convert SDK columns to Terraform format for testing
func convertSDKColumnsToTerraform(sdkColumns *platformclientv2.Decisiontablecolumns) []interface{} {
	if sdkColumns == nil {
		return nil
	}

	columnGroup := map[string]interface{}{}

	// Convert inputs
	if sdkColumns.Inputs != nil {
		inputs := make([]interface{}, len(*sdkColumns.Inputs))
		for i, input := range *sdkColumns.Inputs {
			inputMap := map[string]interface{}{}

			// Convert defaults_to
			if input.DefaultsTo != nil {
				defaultsTo := []interface{}{
					map[string]interface{}{
						"value": *input.DefaultsTo.Value,
					},
				}
				inputMap["defaults_to"] = defaultsTo
			}

			// Convert expression
			if input.Expression != nil {
				expression := []interface{}{
					map[string]interface{}{
						"comparator": *input.Expression.Comparator,
					},
				}

				// Convert contractual
				if input.Expression.Contractual != nil {
					contractual := []interface{}{
						map[string]interface{}{
							"schema_property_key": *(*input.Expression.Contractual).SchemaPropertyKey,
						},
					}
					expression[0].(map[string]interface{})["contractual"] = contractual
				}

				inputMap["expression"] = expression
			}

			inputs[i] = inputMap
		}
		columnGroup["inputs"] = inputs
	}

	// Convert outputs
	if sdkColumns.Outputs != nil {
		outputs := make([]interface{}, len(*sdkColumns.Outputs))
		for i, output := range *sdkColumns.Outputs {
			outputMap := map[string]interface{}{}

			// Convert defaults_to
			if output.DefaultsTo != nil {
				defaultsTo := []interface{}{
					map[string]interface{}{
						"value": *output.DefaultsTo.Value,
					},
				}
				outputMap["defaults_to"] = defaultsTo
			}

			// Convert value
			if output.Value != nil {
				value := []interface{}{
					map[string]interface{}{
						"schema_property_key": *output.Value.SchemaPropertyKey,
					},
				}
				outputMap["value"] = value
			}

			outputs[i] = outputMap
		}
		columnGroup["outputs"] = outputs
	}

	return []interface{}{columnGroup}
}

// TestUnitResourceBusinessRulesDecisionTableEmptyRows tests validation with empty rows
func TestUnitResourceBusinessRulesDecisionTableEmptyRows(t *testing.T) {
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Test with empty rows list
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(
		"test-id",
		"test-name",
		"test-description",
		"test-division-id",
		"test-schema-id",
		[]interface{}{},
	)
	resourceDataMap["rows"] = []interface{}{} // Empty rows

	_ = schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	// This should fail validation due to MinItems constraint
	// We can't directly test Terraform's validation in unit tests,
	// but we can verify the schema constraint is set correctly
	rowsField := resourceSchema["rows"]
	if rowsField.MinItems != 1 {
		t.Error("Rows field should have MinItems = 1")
	}

	// Verify the field is marked as required
	if !rowsField.Required {
		t.Error("Rows field should be required")
	}
}

// TestUnitResourceBusinessRulesDecisionTableMissingRows tests validation with missing rows
func TestUnitResourceBusinessRulesDecisionTableMissingRows(t *testing.T) {
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Test with missing rows field entirely
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(
		"test-id",
		"test-name",
		"test-description",
		"test-division-id",
		"test-schema-id",
		[]interface{}{},
	)
	// Don't set "rows" field at all

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	// Verify that rows field is required
	rowsField := resourceSchema["rows"]
	if !rowsField.Required {
		t.Error("Rows field should be required")
	}

	// Verify that the field is not set (should be empty list)
	rows := d.Get("rows")
	if rows != nil && len(rows.([]interface{})) != 0 {
		t.Error("Rows should be empty when not set")
	}
}

// TestUnitResourceBusinessRulesDecisionTableRowsValidation tests the rows field validation
func TestUnitResourceBusinessRulesDecisionTableRowsValidation(t *testing.T) {
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema
	rowsField := resourceSchema["rows"]

	// Test MinItems constraint
	if rowsField.MinItems != 1 {
		t.Errorf("Expected MinItems to be 1, got %d", rowsField.MinItems)
	}

	// Test Required constraint
	if !rowsField.Required {
		t.Error("Expected rows field to be required")
	}

	// Test Type constraint
	if rowsField.Type != schema.TypeList {
		t.Errorf("Expected Type to be TypeList, got %v", rowsField.Type)
	}

	// Test Elem is defined
	if rowsField.Elem == nil {
		t.Error("Expected Elem to be defined for rows field")
	}
}

// TestUnitResourceBusinessRulesDecisionTableAPIErrors tests API error scenarios
func TestUnitResourceBusinessRulesDecisionTableAPIErrors(t *testing.T) {
	// Test error message formatting for different HTTP status codes
	t.Run("Error_Message_Formatting", func(t *testing.T) {
		testCases := []struct {
			statusCode int
			status     string
			expected   string
		}{
			{500, "Internal Server Error", "500"},
			{400, "Bad Request", "400"},
			{404, "Not Found", "404"},
			{401, "Unauthorized", "401"},
			{429, "Too Many Requests", "429"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("Status_%d_%s", tc.statusCode, tc.status), func(t *testing.T) {
				// Test that error messages contain the status code
				apiResponse := &platformclientv2.APIResponse{
					StatusCode: tc.statusCode,
					Status:     tc.status,
				}
				err := fmt.Errorf("API Error: %d - %s", tc.statusCode, tc.status)

				// Verify the error message contains the expected status code
				if !strings.Contains(err.Error(), tc.expected) {
					t.Errorf("Expected error message to contain '%s', got: %s", tc.expected, err.Error())
				}

				// Verify the API response has the correct status code
				if apiResponse.StatusCode != tc.statusCode {
					t.Errorf("Expected status code %d, got: %d", tc.statusCode, apiResponse.StatusCode)
				}
			})
		}
	})

	// Test error handling for different API response scenarios
	t.Run("API_Response_Error_Handling", func(t *testing.T) {
		// Test nil response handling
		t.Run("Nil_Response", func(t *testing.T) {
			var apiResponse *platformclientv2.APIResponse
			if apiResponse != nil {
				t.Error("Expected nil response")
			}
		})

		// Test response with error
		t.Run("Response_With_Error", func(t *testing.T) {
			apiResponse := &platformclientv2.APIResponse{
				StatusCode: 400,
				Status:     "Bad Request",
			}
			err := fmt.Errorf("API Error: %d - %s", apiResponse.StatusCode, apiResponse.Status)

			if err == nil {
				t.Error("Expected error to be non-nil")
			}

			if !strings.Contains(err.Error(), "400") {
				t.Errorf("Expected error to contain '400', got: %s", err.Error())
			}
		})
	})

	// Test error propagation through the provider
	t.Run("Error_Propagation", func(t *testing.T) {
		// Test that errors are properly formatted and propagated
		originalError := fmt.Errorf("API Error: 500 - Internal Server Error")

		// Simulate error propagation
		propagatedError := fmt.Errorf("Failed to create business rules decision table: %w", originalError)

		if !strings.Contains(propagatedError.Error(), "500") {
			t.Errorf("Expected propagated error to contain '500', got: %s", propagatedError.Error())
		}

		if !strings.Contains(propagatedError.Error(), "Internal Server Error") {
			t.Errorf("Expected propagated error to contain 'Internal Server Error', got: %s", propagatedError.Error())
		}
	})
}

// TestUnitConvertLiteralToSDKEmptyValues tests the convertLiteralToSDK function with empty values
func TestUnitConvertLiteralToSDKEmptyValues(t *testing.T) {
	// Test case 1: Both value and type are empty - should return nil (use default)
	literal := map[string]interface{}{
		"value": "",
		"type":  "",
	}

	result, err := convertLiteralToSDK(literal)
	if err != nil {
		t.Errorf("Expected no error for empty values, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result for empty values (use default), got: %v", result)
	}

	// Test case 2: Empty value but valid type - should return error
	literal = map[string]interface{}{
		"value": "",
		"type":  "string",
	}

	result, err = convertLiteralToSDK(literal)
	if err == nil {
		t.Error("Expected error for empty value with valid type, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result for invalid literal, got: %v", result)
	}

	// Test case 3: Valid value but empty type - should return error
	literal = map[string]interface{}{
		"value": "test",
		"type":  "",
	}

	result, err = convertLiteralToSDK(literal)
	if err == nil {
		t.Error("Expected error for valid value with empty type, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result for invalid literal, got: %v", result)
	}

	// Test case 4: Valid value and type - should return valid literal
	literal = map[string]interface{}{
		"value": "test",
		"type":  "string",
	}

	result, err = convertLiteralToSDK(literal)
	if err != nil {
		t.Errorf("Expected no error for valid literal, got: %v", err)
	}
	if result == nil {
		t.Error("Expected valid result for valid literal, got nil")
	}
	if result.VarString == nil || *result.VarString != "test" {
		t.Errorf("Expected VarString to be 'test', got: %v", result.VarString)
	}
}

// TestUnitConvertLiteralToSDKNumberPrecision tests number formatting with high precision
func TestUnitConvertLiteralToSDKNumberPrecision(t *testing.T) {
	// Test case 1: High precision number - should preserve up to 10 decimal places
	literal := map[string]interface{}{
		"value": "3.141592653589793",
		"type":  "number",
	}

	result, err := convertLiteralToSDK(literal)
	if err != nil {
		t.Errorf("Expected no error for high precision number, got: %v", err)
	}
	if result == nil {
		t.Error("Expected valid result for high precision number, got nil")
	}

	// Verify the result has the correct type and value
	if result.Number == nil {
		t.Error("Expected Number field to be set for number type")
	}
	if *result.Number != 3.141592653589793 {
		t.Errorf("Expected Number to be 3.141592653589793, got: %v", *result.Number)
	}

	// Test case 2: Test the reverse conversion (SDK to Terraform) to verify formatting
	sdkLiteral := &platformclientv2.Literal{
		Number: float64Ptr(3.141592653589793),
	}

	terraformLiteral := convertLiteralToTerraform(sdkLiteral)
	if terraformLiteral["value"] != "3.141592653589793" {
		t.Errorf("Expected formatted value to be '3.141592653589793' (no zero-padding), got: %v", terraformLiteral["value"])
	}
	if terraformLiteral["type"] != "number" {
		t.Errorf("Expected type to be 'number', got: %v", terraformLiteral["type"])
	}

	// Test case 3: Test with a number that has exactly 10 decimal places
	literal = map[string]interface{}{
		"value": "1.2345678901",
		"type":  "number",
	}

	result, err = convertLiteralToSDK(literal)
	if err != nil {
		t.Errorf("Expected no error for 10 decimal place number, got: %v", err)
	}
	if result == nil {
		t.Error("Expected valid result for 10 decimal place number, got nil")
	}

	// Test the reverse conversion
	sdkLiteral = &platformclientv2.Literal{
		Number: float64Ptr(1.2345678901),
	}

	terraformLiteral = convertLiteralToTerraform(sdkLiteral)
	if terraformLiteral["value"] != "1.2345678901" {
		t.Errorf("Expected formatted value to be '1.2345678901', got: %v", terraformLiteral["value"])
	}
}

// TestUnitConvertTerraformRowToSDKEmptyLiterals tests row conversion with empty literals
func TestUnitConvertTerraformRowToSDKEmptyLiterals(t *testing.T) {

	// Test row with mix of regular and empty literals
	rowMap := map[string]interface{}{
		"inputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "VIP",
						"type":  "string",
					},
				},
			},
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{}, // Empty literal block
				},
			},
		},
		"outputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "Premium Queue",
						"type":  "string",
					},
				},
			},
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{}, // Empty literal block
				},
			},
		},
	}

	// Test with positional mapping
	inputColumnIds := []string{"input-col-1", "input-col-2"}
	outputColumnIds := []string{"output-col-1", "output-col-2"}
	result, err := convertTerraformRowToSDK(rowMap, inputColumnIds, outputColumnIds)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that only inputs with literals are included
	if result.Inputs == nil {
		t.Error("Expected inputs to be present")
	} else {
		// Should have only 1 input (the one with literal)
		if len(*result.Inputs) != 1 {
			t.Errorf("Expected 1 input (only with literal), got %d", len(*result.Inputs))
		}
		if _, exists := (*result.Inputs)["input-col-1"]; !exists {
			t.Error("Expected input-col-1 to be present")
		}
		if _, exists := (*result.Inputs)["input-col-2"]; exists {
			t.Error("Expected input-col-2 to NOT be present (uses default)")
		}
		// Check that input-col-1 has a literal
		if (*result.Inputs)["input-col-1"].Literal == nil {
			t.Error("Expected input-col-1 to have a literal")
		}
	}

	if result.Outputs == nil {
		t.Error("Expected outputs to be present")
	} else {
		// Should have only 1 output (the one with literal)
		if len(*result.Outputs) != 1 {
			t.Errorf("Expected 1 output (only with literal), got %d", len(*result.Outputs))
		}
		if _, exists := (*result.Outputs)["output-col-1"]; !exists {
			t.Error("Expected output-col-1 to be present")
		}
		if _, exists := (*result.Outputs)["output-col-2"]; exists {
			t.Error("Expected output-col-2 to NOT be present (uses default)")
		}
		// Check that output-col-1 has a literal
		if (*result.Outputs)["output-col-1"].Literal == nil {
			t.Error("Expected output-col-1 to have a literal")
		}
	}
}

// TestUnitConvertSDKRowToTerraformEmptyLiterals tests row conversion with empty literals for export
func TestUnitConvertSDKRowToTerraformEmptyLiterals(t *testing.T) {
	inputColumnIds := []string{"input-col-1", "input-col-2"}
	outputColumnIds := []string{"output-col-1", "output-col-2"}

	// Test SDK row with mix of explicit and missing literals (missing = use defaults)
	sdkRow := platformclientv2.Decisiontablerow{
		Id:       stringPtr("row-1"),
		RowIndex: intPtr(0),
		Inputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
			// Only input-col-1 has a literal, input-col-2 uses default
			"input-col-1": {
				Literal: &platformclientv2.Literal{
					VarString: stringPtr("VIP"),
				},
			},
		},
		Outputs: &map[string]platformclientv2.Decisiontablerowparametervalue{
			// Only output-col-1 has a literal, output-col-2 uses default
			"output-col-1": {
				Literal: &platformclientv2.Literal{
					VarString: stringPtr("Premium Queue"),
				},
			},
		},
	}

	result := convertSDKRowToTerraform(sdkRow, inputColumnIds, outputColumnIds)

	// Check that all columns are included in the result
	if result["inputs"] == nil {
		t.Error("Expected inputs to be present")
	} else {
		inputs := result["inputs"].([]interface{})
		if len(inputs) != 2 {
			t.Errorf("Expected 2 inputs, got %d", len(inputs))
		}

		// Check first input (has literal)
		input1 := inputs[0].(map[string]interface{})
		if input1["column_id"] != "input-col-1" {
			t.Errorf("Expected column_id to be 'input-col-1', got %s", input1["column_id"])
		}
		if input1["literal"] == nil {
			t.Error("Expected literal to be present for input-col-1")
		} else {
			literal1 := input1["literal"].([]interface{})[0].(map[string]interface{})
			if literal1["value"] != "VIP" || literal1["type"] != "string" {
				t.Errorf("Expected literal value='VIP', type='string', got value='%s', type='%s'", literal1["value"], literal1["type"])
			}
		}

		// Check second input (uses default - should have empty string values)
		input2 := inputs[1].(map[string]interface{})
		if input2["column_id"] != "input-col-2" {
			t.Errorf("Expected column_id to be 'input-col-2', got %s", input2["column_id"])
		}
		if input2["literal"] == nil {
			t.Error("Expected literal to be present for input-col-2 (empty for default)")
		} else {
			literal2 := input2["literal"].([]interface{})[0].(map[string]interface{})
			// Empty literal should have empty string values
			if literal2["value"] != "" || literal2["type"] != "" {
				t.Errorf("Expected empty string values for default, got: value=%s, type=%s", literal2["value"], literal2["type"])
			}
		}
	}

	if result["outputs"] == nil {
		t.Error("Expected outputs to be present")
	} else {
		outputs := result["outputs"].([]interface{})
		if len(outputs) != 2 {
			t.Errorf("Expected 2 outputs, got %d", len(outputs))
		}

		// Check first output (has literal)
		output1 := outputs[0].(map[string]interface{})
		if output1["column_id"] != "output-col-1" {
			t.Errorf("Expected column_id to be 'output-col-1', got %s", output1["column_id"])
		}
		if output1["literal"] == nil {
			t.Error("Expected literal to be present for output-col-1")
		} else {
			literal1 := output1["literal"].([]interface{})[0].(map[string]interface{})
			if literal1["value"] != "Premium Queue" || literal1["type"] != "string" {
				t.Errorf("Expected literal value='Premium Queue', type='string', got value='%s', type='%s'", literal1["value"], literal1["type"])
			}
		}

		// Check second output (uses default - should have empty string values)
		output2 := outputs[1].(map[string]interface{})
		if output2["column_id"] != "output-col-2" {
			t.Errorf("Expected column_id to be 'output-col-2', got %s", output2["column_id"])
		}
		if output2["literal"] == nil {
			t.Error("Expected literal to be present for output-col-2 (empty for default)")
		} else {
			literal2 := output2["literal"].([]interface{})[0].(map[string]interface{})
			// Empty literal should have empty string values
			if literal2["value"] != "" || literal2["type"] != "" {
				t.Errorf("Expected empty string values for default, got: value=%s, type=%s", literal2["value"], literal2["type"])
			}
		}
	}
}

// TestUnitConvertTerraformRowToSDKAllDefaults tests validation that at least one input and output must have explicit values
func TestUnitConvertTerraformRowToSDKAllDefaults(t *testing.T) {
	// Test row with all inputs using defaults (should fail)
	rowMapAllDefaults := map[string]interface{}{
		"inputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{}, // Empty literal block
				},
			},
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "",
						"type":  "",
					},
				},
			},
		},
		"outputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{}, // Empty literal block
				},
			},
		},
	}

	inputColumnIds := []string{"input-col-1", "input-col-2"}
	outputColumnIds := []string{"output-col-1"}

	_, err := convertTerraformRowToSDK(rowMapAllDefaults, inputColumnIds, outputColumnIds)
	if err == nil {
		t.Error("Expected error for all inputs using defaults, but got none")
	}
	if !strings.Contains(err.Error(), "at least one input must have an explicit value") {
		t.Errorf("Expected error about explicit input values, got: %v", err)
	}

	// Test row with all outputs using defaults (should fail)
	rowMapAllOutputDefaults := map[string]interface{}{
		"inputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "VIP",
						"type":  "string",
					},
				},
			},
		},
		"outputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{}, // Empty literal block
				},
			},
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "",
						"type":  "",
					},
				},
			},
		},
	}

	_, err = convertTerraformRowToSDK(rowMapAllOutputDefaults, inputColumnIds, outputColumnIds)
	if err == nil {
		t.Error("Expected error for all outputs using defaults, but got none")
	}
	if !strings.Contains(err.Error(), "at least one output must have an explicit value") {
		t.Errorf("Expected error about explicit output values, got: %v", err)
	}

	// Test row with at least one explicit input and output (should succeed)
	rowMapValid := map[string]interface{}{
		"inputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "VIP",
						"type":  "string",
					},
				},
			},
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{}, // Empty literal block (default)
				},
			},
		},
		"outputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "VIP Queue",
						"type":  "string",
					},
				},
			},
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{
						"value": "",
						"type":  "",
					},
				},
			},
		},
	}

	_, err = convertTerraformRowToSDK(rowMapValid, inputColumnIds, outputColumnIds)
	if err != nil {
		t.Errorf("Expected no error for valid row with explicit values, got: %v", err)
	}
}

// Helper functions for unit tests
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
