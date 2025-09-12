package business_rules_decision_table

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// Queue lookup interface for dependency injection during testing
type QueueLookupProvider interface {
	GetQueueByID(ctx context.Context, queueID string) (*platformclientv2.Queue, error)
}

// Schema lookup interface for determining column types
type SchemaLookupProvider interface {
	GetSchemaByID(ctx context.Context, schemaID string) (*platformclientv2.Dataschema, error)
}

// Default queue lookup implementation
type DefaultQueueLookupProvider struct {
	clientConfig *platformclientv2.Configuration
}

func NewDefaultQueueLookupProvider(clientConfig *platformclientv2.Configuration) *DefaultQueueLookupProvider {
	return &DefaultQueueLookupProvider{
		clientConfig: clientConfig,
	}
}

func (p *DefaultQueueLookupProvider) GetQueueByID(ctx context.Context, queueID string) (*platformclientv2.Queue, error) {
	routingApi := platformclientv2.NewRoutingApiWithConfig(p.clientConfig)
	queue, _, err := routingApi.GetRoutingQueue(queueID, nil)
	return queue, err
}

// Default schema lookup implementation
type DefaultSchemaLookupProvider struct {
	clientConfig *platformclientv2.Configuration
}

func NewDefaultSchemaLookupProvider(clientConfig *platformclientv2.Configuration) *DefaultSchemaLookupProvider {
	return &DefaultSchemaLookupProvider{
		clientConfig: clientConfig,
	}
}

func (p *DefaultSchemaLookupProvider) GetSchemaByID(ctx context.Context, schemaID string) (*platformclientv2.Dataschema, error) {
	businessRulesApi := platformclientv2.NewBusinessRulesApiWithConfig(p.clientConfig)
	schema, _, err := businessRulesApi.GetBusinessrulesSchema(schemaID)
	return schema, err
}

// buildSdkInputColumns builds the SDK input columns from the Terraform schema
func buildSdkInputColumns(inputColumns []interface{}) *[]platformclientv2.Decisiontableinputcolumnrequest {
	if len(inputColumns) == 0 {
		return nil
	}

	sdkInputColumns := make([]platformclientv2.Decisiontableinputcolumnrequest, 0)
	for _, inputColumn := range inputColumns {
		inputColumnMap := inputColumn.(map[string]interface{})
		sdkInputColumn := platformclientv2.Decisiontableinputcolumnrequest{}

		if defaultsToList, ok := inputColumnMap["defaults_to"].([]interface{}); ok && len(defaultsToList) > 0 {
			defaultsToMap := defaultsToList[0].(map[string]interface{})

			// Check for special values first
			special, specialOk := defaultsToMap["special"].(string)
			value, valueOk := defaultsToMap["value"].(string)

			if specialOk && special != "" {
				sdkInputColumn.DefaultsTo = &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Special: &special,
				}
			} else if valueOk && value != "" {
				// Only set Value, leave Special as nil
				sdkInputColumn.DefaultsTo = &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: &value,
				}
			}
		}

		if expressions, ok := inputColumnMap["expression"].([]interface{}); ok {
			sdkInputColumn.Expression = buildSdkExpression(expressions[0].(map[string]interface{}))
		}

		sdkInputColumns = append(sdkInputColumns, sdkInputColumn)
	}

	return &sdkInputColumns
}

// buildSdkOutputColumns builds the SDK output columns from the Terraform schema
func buildSdkOutputColumns(outputColumns []interface{}) *[]platformclientv2.Decisiontableoutputcolumnrequest {
	if len(outputColumns) == 0 {
		return nil
	}

	sdkOutputColumns := make([]platformclientv2.Decisiontableoutputcolumnrequest, 0)
	for _, outputColumn := range outputColumns {
		outputColumnMap := outputColumn.(map[string]interface{})
		sdkOutputColumn := platformclientv2.Decisiontableoutputcolumnrequest{}

		if defaultsToList, ok := outputColumnMap["defaults_to"].([]interface{}); ok && len(defaultsToList) > 0 {
			defaultsToMap := defaultsToList[0].(map[string]interface{})

			if special, ok := defaultsToMap["special"].(string); ok && special != "" {
				sdkOutputColumn.DefaultsTo = &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Special: &special,
				}
			} else if value, ok := defaultsToMap["value"].(string); ok {
				sdkOutputColumn.DefaultsTo = &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: &value,
				}
			} else {
				// Try to convert to string if it's a different type
				if defaultsToMap["value"] != nil {
					if valueStr, ok := defaultsToMap["value"].(string); ok {
						sdkOutputColumn.DefaultsTo = &platformclientv2.Decisiontablecolumndefaultrowvalue{
							Value: &valueStr,
						}
					}
				}
			}
		}

		if values, ok := outputColumnMap["value"].([]interface{}); ok {
			sdkOutputColumn.Value = buildSdkValue(values[0].(map[string]interface{}))
		}

		sdkOutputColumns = append(sdkOutputColumns, sdkOutputColumn)
	}

	return &sdkOutputColumns
}

// buildSdkExpression builds the SDK expression from the Terraform schema
func buildSdkExpression(expression map[string]interface{}) *platformclientv2.Decisiontableinputcolumnexpression {
	sdkExpression := platformclientv2.Decisiontableinputcolumnexpression{}

	if contractual, ok := expression["contractual"].([]interface{}); ok && len(contractual) > 0 {
		sdkExpression.Contractual = buildSdkContractual(contractual[0].(map[string]interface{}))
	}

	if comparator, ok := expression["comparator"].(string); ok {
		sdkExpression.Comparator = &comparator
	}

	return &sdkExpression
}

// buildSdkValue builds the SDK value from the Terraform schema
func buildSdkValue(value map[string]interface{}) *platformclientv2.Outputvalue {
	sdkValue := platformclientv2.Outputvalue{}

	if schemaPropertyKey, ok := value["schema_property_key"].(string); ok {
		sdkValue.SchemaPropertyKey = &schemaPropertyKey
	}

	if properties, ok := value["properties"].([]interface{}); ok && len(properties) > 0 {
		sdkValue.Properties = buildSdkProperties(properties)
	}

	return &sdkValue
}

// buildSdkContractual builds the SDK contractual from the Terraform schema
func buildSdkContractual(contractual map[string]interface{}) **platformclientv2.Contractual {
	sdkContractual := platformclientv2.Contractual{}

	if schemaPropertyKey, ok := contractual["schema_property_key"].(string); ok {
		sdkContractual.SchemaPropertyKey = &schemaPropertyKey
	}

	if nestedContractual, ok := contractual["contractual"].([]interface{}); ok && len(nestedContractual) > 0 {
		nested := buildSdkContractual(nestedContractual[0].(map[string]interface{}))
		sdkContractual.Contractual = nested
	}

	result := &sdkContractual
	return &result
}

// buildSdkProperties builds the SDK properties from the Terraform schema
func buildSdkProperties(properties []interface{}) *[]platformclientv2.Outputvalue {
	if len(properties) == 0 {
		return nil
	}

	sdkProperties := make([]platformclientv2.Outputvalue, 0)
	for _, property := range properties {
		propertyMap := property.(map[string]interface{})
		sdkProperty := platformclientv2.Outputvalue{}

		if schemaPropertyKey, ok := propertyMap["schema_property_key"].(string); ok {
			sdkProperty.SchemaPropertyKey = &schemaPropertyKey
		}

		if nestedProperties, ok := propertyMap["properties"].([]interface{}); ok {
			sdkProperty.Properties = buildSdkProperties(nestedProperties)
		}

		sdkProperties = append(sdkProperties, sdkProperty)
	}

	return &sdkProperties
}

// buildSdkColumns builds the SDK columns from the Terraform schema
func buildSdkColumns(columns map[string]interface{}) *platformclientv2.Createdecisiontablecolumnsrequest {
	sdkColumns := &platformclientv2.Createdecisiontablecolumnsrequest{}

	if inputs, ok := columns["inputs"].([]interface{}); ok {
		sdkColumns.Inputs = buildSdkInputColumns(inputs)
	}

	if outputs, ok := columns["outputs"].([]interface{}); ok {
		sdkColumns.Outputs = buildSdkOutputColumns(outputs)
	}

	return sdkColumns
}

// buildSdkUpdateColumns builds the SDK update columns from the Terraform schema
func buildSdkUpdateColumns(columns map[string]interface{}) *platformclientv2.Updatedecisiontablecolumnsrequest {
	sdkColumns := &platformclientv2.Updatedecisiontablecolumnsrequest{}

	if inputs, ok := columns["inputs"].([]interface{}); ok {
		sdkColumns.Inputs = buildSdkInputColumns(inputs)
	}

	if outputs, ok := columns["outputs"].([]interface{}); ok {
		sdkColumns.Outputs = buildSdkOutputColumns(outputs)
	}

	return sdkColumns
}

// buildTerraformColumns builds the Terraform columns from the SDK response
func buildTerraformColumns(sdkColumns *platformclientv2.Decisiontablecolumns, queueLookup QueueLookupProvider, schemaLookup SchemaLookupProvider, schemaID string, ctx context.Context) map[string]interface{} {
	if sdkColumns == nil {
		return make(map[string]interface{})
	}

	columns := make(map[string]interface{})

	// Get the schema to determine column types
	var schema *platformclientv2.Dataschema
	if schemaLookup != nil && schemaID != "" {
		var err error
		schema, err = schemaLookup.GetSchemaByID(ctx, schemaID)
		if err != nil {
			log.Printf("Warning: Could not look up schema %s for column type detection: %v", schemaID, err)
		}
	}

	if sdkColumns.Inputs != nil {
		inputs := buildTerraformInputColumns(*sdkColumns.Inputs, queueLookup, schema, ctx)
		columns["inputs"] = inputs
	}

	if sdkColumns.Outputs != nil {
		outputs := buildTerraformOutputColumns(*sdkColumns.Outputs, queueLookup, schema, ctx)
		columns["outputs"] = outputs
	}

	return columns
}

// buildTerraformInputColumns builds the Terraform input columns from the SDK response
func buildTerraformInputColumns(sdkInputColumns []platformclientv2.Decisiontableinputcolumn, queueLookup QueueLookupProvider, schema *platformclientv2.Dataschema, ctx context.Context) []interface{} {
	inputs := make([]interface{}, 0)
	for _, sdkInput := range sdkInputColumns {
		input := make(map[string]interface{})

		if sdkInput.Id != nil {
			input["id"] = *sdkInput.Id
		}

		// Handle both Special and Value fields for defaults_to
		if sdkInput.DefaultsTo != nil {
			defaultsTo := make(map[string]interface{})
			if sdkInput.DefaultsTo.Special != nil {
				defaultsTo["special"] = *sdkInput.DefaultsTo.Special
			} else if sdkInput.DefaultsTo.Value != nil {
				// Preserve the original queue ID without conversion to maintain state consistency
				defaultsTo["value"] = *sdkInput.DefaultsTo.Value
			}
			input["defaults_to"] = []interface{}{defaultsTo}
		}

		if sdkInput.Expression != nil {
			expression := buildTerraformExpression(sdkInput.Expression)
			input["expression"] = []interface{}{expression}
		}

		inputs = append(inputs, input)
	}
	return inputs
}

// buildTerraformOutputColumns builds the Terraform output columns from the SDK response
func buildTerraformOutputColumns(sdkOutputColumns []platformclientv2.Decisiontableoutputcolumn, queueLookup QueueLookupProvider, schema *platformclientv2.Dataschema, ctx context.Context) []interface{} {
	outputs := make([]interface{}, 0)
	for _, sdkOutput := range sdkOutputColumns {
		output := make(map[string]interface{})

		if sdkOutput.Id != nil {
			output["id"] = *sdkOutput.Id
		}

		// Handle both Special and Value fields for defaults_to
		if sdkOutput.DefaultsTo != nil {
			defaultsTo := make(map[string]interface{})
			if sdkOutput.DefaultsTo.Special != nil {
				defaultsTo["special"] = *sdkOutput.DefaultsTo.Special
			} else if sdkOutput.DefaultsTo.Value != nil {
				// For data source reading, preserve the original value without conversion
				// This ensures consistency with what was set during resource creation
				defaultsTo["value"] = *sdkOutput.DefaultsTo.Value
			}
			output["defaults_to"] = []interface{}{defaultsTo}
		}

		if sdkOutput.Value != nil {
			value := buildTerraformValue(sdkOutput.Value)
			output["value"] = []interface{}{value}
		}

		outputs = append(outputs, output)
	}
	return outputs
}

// buildTerraformExpression builds the Terraform expression from the SDK response
func buildTerraformExpression(sdkExpression *platformclientv2.Decisiontableinputcolumnexpression) map[string]interface{} {
	expression := make(map[string]interface{})

	if sdkExpression.Contractual != nil && *sdkExpression.Contractual != nil {
		contractual := buildTerraformContractual(*sdkExpression.Contractual)
		expression["contractual"] = []interface{}{contractual}
	}

	if sdkExpression.Comparator != nil {
		expression["comparator"] = *sdkExpression.Comparator
	}

	return expression
}

// buildTerraformValue builds the Terraform value from the SDK response
func buildTerraformValue(sdkValue *platformclientv2.Outputvalue) map[string]interface{} {
	value := make(map[string]interface{})

	if sdkValue.SchemaPropertyKey != nil {
		value["schema_property_key"] = *sdkValue.SchemaPropertyKey
	}

	if sdkValue.Properties != nil {
		properties := buildTerraformProperties(*sdkValue.Properties)
		value["properties"] = properties
	}

	return value
}

// buildTerraformContractual builds the Terraform contractual from the SDK response
func buildTerraformContractual(sdkContractual *platformclientv2.Contractual) map[string]interface{} {
	contractual := make(map[string]interface{})

	if sdkContractual.SchemaPropertyKey != nil {
		contractual["schema_property_key"] = *sdkContractual.SchemaPropertyKey
	}

	if sdkContractual.Contractual != nil && *sdkContractual.Contractual != nil {
		nestedContractual := buildTerraformContractual(*sdkContractual.Contractual)
		contractual["contractual"] = []interface{}{nestedContractual}
	}

	return contractual
}

// buildTerraformProperties builds the Terraform properties from the SDK response
func buildTerraformProperties(sdkProperties []platformclientv2.Outputvalue) []interface{} {
	properties := make([]interface{}, 0)
	for _, sdkProperty := range sdkProperties {
		property := make(map[string]interface{})

		if sdkProperty.SchemaPropertyKey != nil {
			property["schema_property_key"] = *sdkProperty.SchemaPropertyKey
		}

		if sdkProperty.Properties != nil {
			nestedProperties := buildTerraformProperties(*sdkProperty.Properties)
			property["properties"] = nestedProperties
		}

		properties = append(properties, property)
	}
	return properties
}

// buildCreateRequest builds a CreateDecisionTableRequest from Terraform resource data
func buildCreateRequest(d *schema.ResourceData) *platformclientv2.Createdecisiontablerequest {
	tableName := d.Get("name").(string)
	divisionId := d.Get("division_id").(string)
	schemaId := d.Get("schema_id").(string)
	columns := d.Get("columns").([]interface{})

	createRequest := &platformclientv2.Createdecisiontablerequest{
		Name:       platformclientv2.String(tableName),
		DivisionId: platformclientv2.String(divisionId),
		SchemaId:   platformclientv2.String(schemaId),
	}

	// Add description if specified (optional field)
	if description, ok := d.GetOk("description"); ok {
		createRequest.Description = platformclientv2.String(description.(string))
	}

	// Build columns (required field)
	if len(columns) > 0 {
		columnData := columns[0].(map[string]interface{})
		createRequest.Columns = buildSdkColumns(columnData)
	}

	return createRequest
}
