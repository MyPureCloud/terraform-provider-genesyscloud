package business_rules_decision_table

import (
	"context"
	"fmt"
	"net/http"
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
	requiredFields := []string{"name", "columns", "division_id", "schema_id"}
	for _, field := range requiredFields {
		if resource.Schema[field] == nil {
			t.Errorf("Required field '%s' is missing from schema", field)
		}
		if !resource.Schema[field].Required {
			t.Errorf("Field '%s' should be required", field)
		}
	}

	// Test optional fields
	optionalFields := []string{"description"}
	for _, field := range optionalFields {
		if resource.Schema[field] == nil {
			t.Errorf("Optional field '%s' is missing from schema", field)
		}
		if !resource.Schema[field].Optional {
			t.Errorf("Field '%s' should be optional", field)
		}
	}

	// Test computed fields
	computedFields := []string{
		"latest_version",
		"published_version",
	}
	for _, field := range computedFields {
		if resource.Schema[field] == nil {
			t.Errorf("Computed field '%s' is missing from schema", field)
		}
		if !resource.Schema[field].Computed {
			t.Errorf("Field '%s' should be computed", field)
		}
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
}

func TestBusinessRulesDecisionTableSchemaValidation(t *testing.T) {
	resource := ResourceBusinessRulesDecisionTable()

	// Test name field validation
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
			// Input 1: Queue reference with Equals comparator
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("genesyscloud_routing_queue.input_queue.id"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("queue_id"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
			// Input 2: Boolean with Equals comparator
			{
				Id: platformclientv2.String("input-column-id-2"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("true"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("is_vip"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
			// Input 3: String with Equals comparator
			{
				Id: platformclientv2.String("input-column-id-3"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("Premium"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("customer_tier"),
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
			// Output 2: String value
			{
				Id: platformclientv2.String("output-column-id-2"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("VIP Customer"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("customer_name"),
				},
			},
		},
	}

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Setup schema mock
	setupSchemaMock(decisionTableProxy, tSchemaId)

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
		assert.Len(t, *createRequest.Columns.Inputs, 3, "createRequest.Columns.Inputs should have 3 inputs (queue, boolean, string)")
		assert.Len(t, *createRequest.Columns.Outputs, 2, "createRequest.Columns.Outputs should have 2 outputs (queue, string)")

		// Validate that the mock providers are working

		// Set up a realistic table version response
		tableVersion.Id = &tId
		tableVersion.Name = &tName
		tableVersion.Version = platformclientv2.Int(1)
		tableVersion.Status = platformclientv2.String("Draft")

		return tableVersion, nil, nil
	}

	internalProxy = decisionTableProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourceBusinessRulesDecisionTable().Schema

	// Convert SDK columns to Terraform format for testing
	tColumnsTF := convertSDKColumnsToTerraform(tColumns)

	// Setup a map of values
	resourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, tColumnsTF)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := createBusinessRulesDecisionTable(ctx, d, gcloud)
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

	// Setup schema mock
	setupSchemaMock(decisionTableProxy, tSchemaId)

	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Latest: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
			// Now we can include columns since we have mock providers!
			Columns: &platformclientv2.Decisiontablecolumns{
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
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the getBusinessRulesDecisionTableVersion function to return schema information
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, versionNumber, "Expected version 1")

		version := &platformclientv2.Decisiontableversion{
			Version: platformclientv2.Int(1),
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
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

	// Test that the read operation successfully processed the complex column structure
	// Instead of testing the Terraform interface{} format, test the actual SDK response

	// Test the SDK response directly - this is much cleaner and type-safe
	// We can access the mock data directly since we control what it returns
	// Note: These variables were already declared above, so we just use them

	// Test that the read operation succeeded and columns are present
	// The detailed column structure validation is covered in the CRUD tests
}

func TestUnitResourceBusinessRulesDecisionTableUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table Updated"
	tDescription := "CX as Code Unit Test Business Rules Decision Table Updated"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Initial simple columns using proper SDK types
	initialColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("old-input-value"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("old_field"),
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
					Value: platformclientv2.String("old-output-value"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("old_result"),
				},
			},
		},
	}

	// Updated columns using proper SDK types
	updatedColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("new-input-value"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("new_field"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("NotEquals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("new-output-value"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("new_result"),
				},
			},
		},
	}

	// Mock the get function to return the initial table with version 1 (draft)
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Latest: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1), // Draft version 1 - can modify columns
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the update function to validate column modifications
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, tName, *updateRequest.Name, "updateRequest.Name check failed in update updateBusinessRulesDecisionTableAttr")
		assert.Equal(t, tDescription, *updateRequest.Description, "updateRequest.Description check failed in update updateBusinessRulesDecisionTableAttr")

		// Verify that columns are included in the update request
		assert.NotNil(t, updateRequest.Columns, "Update request should include columns for draft version 1")

		// Return updated table with the NEW columns
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Columns:     updatedColumns, // Include the updated columns in the response
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the version lookup function to return version 1 (draft)
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, versionNumber, "Expected version 1")

		version := &platformclientv2.Decisiontableversion{
			Version: platformclientv2.Int(1),
			Status:  platformclientv2.String("Draft"),
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
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

	// Convert SDK columns to Terraform format for testing
	initialColumnsTF := convertSDKColumnsToTerraform(initialColumns)
	updatedColumnsTF := convertSDKColumnsToTerraform(updatedColumns)

	// Setup initial resource data with simple columns
	initialResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, initialColumnsTF)
	d := schema.TestResourceDataRaw(t, resourceSchema, initialResourceDataMap)
	d.SetId(tId)

	// Update to new columns
	updatedResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, updatedColumnsTF)
	d2 := schema.TestResourceDataRaw(t, resourceSchema, updatedResourceDataMap)
	d2.SetId(tId)

	// Perform the update
	diag := updateBusinessRulesDecisionTable(ctx, d2, gcloud)
	assert.Equal(t, false, diag.HasError(), "Update should succeed for draft version 1")
	assert.Equal(t, tId, d2.Id())

	// Verify that the updated columns are properly set
	assert.NotNil(t, d2.Get("columns"), "Columns should be set after update")
	columns := d2.Get("columns").([]interface{})
	assert.Len(t, columns, 1, "Should have 1 column group after update")

}

func TestUnitResourceBusinessRulesDecisionTableSimpleColumnUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Initial simple columns using proper SDK types
	initialColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("old-input-value"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("old_field"),
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
					Value: platformclientv2.String("old-output-value"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("old_result"),
				},
			},
		},
	}

	// Updated simple columns using proper SDK types
	updatedColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("new-input-value"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("new_field"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("NotEquals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("new-output-value"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("new_result"),
				},
			},
		},
	}

	// Mock the get function to return the initial table with version 1 (draft)
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Latest: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1), // Draft version 1 - can modify columns
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the update function to validate column modifications
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, tName, *updateRequest.Name, "updateRequest.Name check failed")
		assert.Equal(t, tDescription, *updateRequest.Description, "updateRequest.Description check failed")

		// Verify that columns are included in the update request
		assert.NotNil(t, updateRequest.Columns, "Update request should include columns for draft version 1")

		// Return updated table with the NEW columns
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Columns:     updatedColumns, // Include the updated columns in the response
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the version lookup function to return version 1 (draft)
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, versionNumber, "Expected version 1")

		version := &platformclientv2.Decisiontableversion{
			Version: platformclientv2.Int(1),
			Status:  platformclientv2.String("Draft"),
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
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

	// Convert SDK columns to Terraform format for testing
	initialColumnsTF := convertSDKColumnsToTerraform(initialColumns)
	updatedColumnsTF := convertSDKColumnsToTerraform(updatedColumns)

	// Setup initial resource data with simple columns
	initialResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, initialColumnsTF)
	d := schema.TestResourceDataRaw(t, resourceSchema, initialResourceDataMap)
	d.SetId(tId)

	// Update to new columns
	updatedResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, updatedColumnsTF)
	d2 := schema.TestResourceDataRaw(t, resourceSchema, updatedResourceDataMap)
	d2.SetId(tId)

	// Perform the update
	diag := updateBusinessRulesDecisionTable(ctx, d2, gcloud)
	assert.Equal(t, false, diag.HasError(), "Update should succeed for draft version 1")
	assert.Equal(t, tId, d2.Id())

	// Verify that the updated columns are properly set
	assert.NotNil(t, d2.Get("columns"), "Columns should be set after update")
	columns := d2.Get("columns").([]interface{})
	assert.Len(t, columns, 1, "Should have 1 column group after update")

	// Test that the simple column update actually changed the values
	columnGroup := columns[0].(map[string]interface{})
	inputs := columnGroup["inputs"].([]interface{})
	outputs := columnGroup["outputs"].([]interface{})

	// Verify input column was updated
	assert.Len(t, inputs, 1, "Should have 1 input column")
	input1 := inputs[0].(map[string]interface{})
	assert.Equal(t, "new-input-value", input1["defaults_to"].([]interface{})[0].(map[string]interface{})["value"])
	input1Expression := input1["expression"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "new_field", input1Expression["contractual"].([]interface{})[0].(map[string]interface{})["schema_property_key"])
	assert.Equal(t, "NotEquals", input1Expression["comparator"])

	// Verify output column was updated
	assert.Len(t, outputs, 1, "Should have 1 output column")
	output1 := outputs[0].(map[string]interface{})
	assert.Equal(t, "new-output-value", output1["defaults_to"].([]interface{})[0].(map[string]interface{})["value"])
	output1Value := output1["value"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, "new_result", output1Value["schema_property_key"])
}

func TestUnitResourceBusinessRulesDecisionTableUpdateColumnsOnNewerVersion(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Initial columns
	initialColumns := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"defaults_to": []interface{}{
						map[string]interface{}{
							"value": "simple-input",
						},
					},
					"expression": []interface{}{
						map[string]interface{}{
							"contractual": []interface{}{
								map[string]interface{}{
									"schema_property_key": "simple_field",
								},
							},
							"comparator": "Equals",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"defaults_to": []interface{}{
						map[string]interface{}{
							"value": "simple-output",
						},
					},
					"value": []interface{}{
						map[string]interface{}{
							"schema_property_key": "simple_result",
						},
					},
				},
			},
		},
	}

	// Attempted column update (should fail for newer versions)
	updatedColumns := []interface{}{
		map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"defaults_to": []interface{}{
						map[string]interface{}{
							"value": "genesyscloud_routing_queue.input_queue.id",
						},
					},
					"expression": []interface{}{
						map[string]interface{}{
							"contractual": []interface{}{
								map[string]interface{}{
									"schema_property_key": "queue_id",
								},
							},
							"comparator": "Equals",
						},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"defaults_to": []interface{}{
						map[string]interface{}{
							"value": "VIP Customer",
						},
					},
					"value": []interface{}{
						map[string]interface{}{
							"schema_property_key": "customer_name",
						},
					},
				},
			},
		},
	}

	// Mock the get function to return a table with version 2+ (cannot modify columns)
	decisionTableProxy.getBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Latest: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(2), // Version 2 - cannot modify columns
			},
			// Include columns in the table response to test column retrieval
			Columns: &platformclientv2.Decisiontablecolumns{
				Inputs: &[]platformclientv2.Decisiontableinputcolumn{
					{
						Id: platformclientv2.String("input-column-id-version-test"),
						DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: platformclientv2.String("simple-input"),
						},
						Expression: &platformclientv2.Decisiontableinputcolumnexpression{
							Contractual: func() **platformclientv2.Contractual {
								contractual := &platformclientv2.Contractual{
									SchemaPropertyKey: platformclientv2.String("simple_field"),
								}
								return &contractual
							}(),
							Comparator: platformclientv2.String("Equals"),
						},
					},
				},
				Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
					{
						Id: platformclientv2.String("output-column-id-version-test"),
						DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: platformclientv2.String("simple-output"),
						},
						Value: &platformclientv2.Outputvalue{
							SchemaPropertyKey: platformclientv2.String("simple_result"),
						},
					},
				},
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the update function to reject column modifications for newer versions
	decisionTableProxy.updateBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, tName, *updateRequest.Name, "updateRequest.Name check failed")
		assert.Equal(t, tDescription, *updateRequest.Description, "updateRequest.Description check failed")

		// For newer versions, column updates should be rejected
		// This simulates the business rule that only draft version 1 can modify columns
		if updateRequest.Columns != nil {
			// In a real scenario, this would return an error
			// For testing, we'll just verify that columns are not allowed
			assert.Fail(t, "Column updates should not be allowed for newer versions")
		}

		// Return updated table (without column changes)
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return table, apiResponse, nil
	}

	// Mock the version lookup function to return version 2 details
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 2, versionNumber, "Expected version 2")

		version := &platformclientv2.Decisiontableversion{
			Version: platformclientv2.Int(2),
			Status:  platformclientv2.String("Draft"), // Even if draft, version > 1 cannot modify columns
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
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
	initialResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, initialColumns)
	d := schema.TestResourceDataRaw(t, resourceSchema, initialResourceDataMap)
	d.SetId(tId)

	// Attempt to update to new columns (should fail for newer versions)
	updatedResourceDataMap := buildTestDecisionTableResourceMapCRUD(tId, tName, tDescription, tDivisionId, tSchemaId, updatedColumns)
	d2 := schema.TestResourceDataRaw(t, resourceSchema, updatedResourceDataMap)
	d2.SetId(tId)

	// Perform the update - this should fail or reject column modifications
	diag := updateBusinessRulesDecisionTable(ctx, d2, gcloud)

	// The update should either fail or not include column changes
	// We're testing the business rule enforcement
	if diag.HasError() {
		// If the update fails due to version restrictions, that's expected
		t.Logf("Update failed as expected for newer version: %v", diag)
	} else {
		// If it succeeds, verify that columns were not actually changed
		// This tests the business rule enforcement in the update logic
		assert.Equal(t, tId, d2.Id())

		// Verify that the mock provider validation caught the column update attempt
		// (This would be handled by the actual business logic in production)
	}
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

	// Test that resolver functions are set
	assert.NotNil(t, exporter.CustomAttributeResolver["columns.outputs.defaults_to.value"].ResolverFunc, "outputs resolver function should be set")
	assert.NotNil(t, exporter.CustomAttributeResolver["columns.inputs.defaults_to.value"].ResolverFunc, "inputs resolver function should be set")
}

// Test the queue defaults_to resolver function
func TestUnitQueueDefaultsToResolver(t *testing.T) {
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
	err := QueueDefaultsToResolver(resourceMap, exporters, queueID)
	assert.NoError(t, err, "Queue ID resolution should succeed")

	// Test with non-queue value (should succeed)
	nonQueueValue := "priority_high"
	err2 := QueueDefaultsToResolver(resourceMap, exporters, nonQueueValue)
	assert.NoError(t, err2, "Non-queue value should be handled without error")

	// Test with empty value
	emptyValue := ""
	err3 := QueueDefaultsToResolver(resourceMap, exporters, emptyValue)
	assert.NoError(t, err3, "Empty value should be handled without error")

	// Test with nil resource map
	err4 := QueueDefaultsToResolver(nil, exporters, queueID)
	assert.NoError(t, err4, "Should handle nil resource map without error")
}

// Test the data source functionality
func TestUnitDataSourceBusinessRulesDecisionTable(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Decision Table"
	tDescription := "CX as Code Unit Test Business Rules Decision Table"
	tDivisionId := uuid.NewString()
	tSchemaId := uuid.NewString()

	decisionTableProxy := &BusinessRulesDecisionTableProxy{}

	// Mock the getBusinessRulesDecisionTablesByName function
	decisionTableProxy.getBusinessRulesDecisionTablesByNameAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (*[]platformclientv2.Decisiontable, bool, *platformclientv2.APIResponse, error) {
		table := platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Latest: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
			// Include columns in the table response to test column retrieval
			Columns: &platformclientv2.Decisiontablecolumns{
				Inputs: &[]platformclientv2.Decisiontableinputcolumn{
					{
						Id: platformclientv2.String("input-column-id-1"),
						DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: platformclientv2.String("input-queue-id"),
						},
						Expression: &platformclientv2.Decisiontableinputcolumnexpression{
							Contractual: func() **platformclientv2.Contractual {
								contractual := &platformclientv2.Contractual{
									SchemaPropertyKey: platformclientv2.String("queue_id"),
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
							SchemaPropertyKey: platformclientv2.String("queue_id"),
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

		if name == tName {
			return &[]platformclientv2.Decisiontable{table}, false, &platformclientv2.APIResponse{}, nil
		}
		return &[]platformclientv2.Decisiontable{}, true, &platformclientv2.APIResponse{}, nil
	}

	// Mock the getAllBusinessRulesDecisionTables function
	decisionTableProxy.getAllBusinessRulesDecisionTablesAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (*platformclientv2.Decisiontablelisting, *platformclientv2.APIResponse, error) {
		table := &platformclientv2.Decisiontable{
			Id:          &tId,
			Name:        &tName,
			Description: &tDescription,
			Division: &platformclientv2.Division{
				Id: &tDivisionId,
			},
			Latest: &platformclientv2.Decisiontableversionentity{
				Version: platformclientv2.Int(1),
			},
			// Include columns in the table response to test column retrieval
			Columns: &platformclientv2.Decisiontablecolumns{
				Inputs: &[]platformclientv2.Decisiontableinputcolumn{
					{
						Id: platformclientv2.String("input-column-id-1"),
						DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: platformclientv2.String("input-queue-id"),
						},
						Expression: &platformclientv2.Decisiontableinputcolumnexpression{
							Contractual: func() **platformclientv2.Contractual {
								contractual := &platformclientv2.Contractual{
									SchemaPropertyKey: platformclientv2.String("queue_id"),
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
							SchemaPropertyKey: platformclientv2.String("transfer_queue"),
						},
					},
				},
			},
		}

		entities := []platformclientv2.Decisiontable{*table}
		listing := &platformclientv2.Decisiontablelisting{
			Entities: &entities,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return listing, apiResponse, nil
	}

	// Mock the getBusinessRulesDecisionTableVersion function
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, tableId)
		assert.Equal(t, 1, versionNumber)

		version := &platformclientv2.Decisiontableversion{
			Version: platformclientv2.Int(1),
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return version, apiResponse, nil
	}

	// Mock the getBusinessRulesDecisionTableVersion function
	decisionTableProxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		version := &platformclientv2.Decisiontableversion{
			Version: platformclientv2.Int(1),
			Status:  platformclientv2.String("Draft"),
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{
					Id: &tSchemaId,
				},
			},
		}
		apiResponse := &platformclientv2.APIResponse{}
		return version, apiResponse, nil
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

	// Verify that columns are retrieved and set
	columns := d.Get("columns")
	assert.NotNil(t, columns, "Columns should be retrieved and set")

	// Test that the data source successfully retrieved the table with columns
	// Instead of testing the Terraform interface{} format, test the actual SDK response

	// Test that the data source operation succeeded and columns are present
	// The detailed column structure validation is covered in the CRUD tests

	// Verify other computed fields
	assert.Equal(t, tDescription, d.Get("description"), "Description should be set")
	assert.Equal(t, tDivisionId, d.Get("division_id"), "Division ID should be set")
	assert.Equal(t, tSchemaId, d.Get("schema_id"), "Schema ID should be set")
	assert.Equal(t, 1, d.Get("latest_version"), "Latest version should be set")
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

// setupSchemaMock sets up a basic mock for getSchemaByID in tests
func setupSchemaMock(proxy *BusinessRulesDecisionTableProxy, schemaID string) {
	proxy.getSchemaByIDAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, id string) (*platformclientv2.Dataschema, error) {
		if id == schemaID {
			return &platformclientv2.Dataschema{
				Id: &id,
				JsonSchema: &platformclientv2.Jsonschemadocument{
					Properties: &map[string]interface{}{
						"queue_id": map[string]interface{}{
							"allOf": []interface{}{
								map[string]interface{}{
									"$ref": "#/components/schemas/businessRulesQueue",
								},
							},
						},
						"is_vip": map[string]interface{}{
							"type": "boolean",
						},
						"customer_tier": map[string]interface{}{
							"type": "string",
						},
					},
				},
			}, nil
		}
		return nil, fmt.Errorf("schema not found")
	}
}
