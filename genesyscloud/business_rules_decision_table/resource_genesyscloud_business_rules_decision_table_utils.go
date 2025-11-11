package business_rules_decision_table

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v172/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// DateTimeParseFormat is the format used for parsing datetime values
const DateTimeParseFormat = "2006-01-02T15:04:05.000Z"

// buildDefaultsTo builds SDK defaults_to from provider schema
func buildDefaultsToFromProvider(defaultsToList []interface{}) *platformclientv2.Decisiontablecolumndefaultrowvalue {
	if len(defaultsToList) == 0 {
		return nil
	}

	defaultsToMap := defaultsToList[0].(map[string]interface{})
	special, specialOk := defaultsToMap["special"].(string)
	value, valueOk := defaultsToMap["value"].(string)

	if specialOk && special != "" {
		return &platformclientv2.Decisiontablecolumndefaultrowvalue{
			Special: &special,
		}
	}

	if valueOk && value != "" {
		return &platformclientv2.Decisiontablecolumndefaultrowvalue{
			Value: &value,
		}
	}

	// Handle type conversion for output columns
	if defaultsToMap["value"] != nil {
		if valueStr, ok := defaultsToMap["value"].(string); ok {
			return &platformclientv2.Decisiontablecolumndefaultrowvalue{
				Value: &valueStr,
			}
		}
	}

	return nil
}

// flattenDefaultsTo flattens SDK defaults_to to provider format
func flattenDefaultsTo(sdkDefaultsTo *platformclientv2.Decisiontablecolumndefaultrowvalue) []interface{} {
	if sdkDefaultsTo == nil {
		return nil
	}

	defaultsTo := make(map[string]interface{})
	if sdkDefaultsTo.Special != nil {
		defaultsTo["special"] = *sdkDefaultsTo.Special
	} else if sdkDefaultsTo.Value != nil {
		defaultsTo["value"] = *sdkDefaultsTo.Value
	}

	return []interface{}{defaultsTo}
}

// validateLiteral validates that a literal block has required fields
func validateLiteral(literal map[string]interface{}) (string, string, error) {
	value, valueOk := literal["value"].(string)
	valueType, typeOk := literal["type"].(string)

	// If both value and type are missing or empty, omit this literal (use column default)
	if (!valueOk || value == "") && (!typeOk || valueType == "") {
		log.Printf("DEBUG: Both value and type are missing or empty, omitting literal (using column default)")
		return "", "", nil
	}

	// If both value and type are empty strings, omit this literal (use column default)
	if value == "" && valueType == "" {
		log.Printf("DEBUG: Both value and type are empty strings, omitting literal (using column default)")
		return "", "", nil
	}

	// If only one is provided, that's an error
	if (!valueOk || value == "") && (typeOk && valueType != "") {
		return "", "", fmt.Errorf("value is required when type is specified")
	}
	if (valueOk && value != "") && (!typeOk || valueType == "") {
		return "", "", fmt.Errorf("type is required when value is specified")
	}

	// If value is not empty but type is empty, that's an error
	if value != "" && valueType == "" {
		return "", "", fmt.Errorf("type cannot be empty when value is specified")
	}

	return value, valueType, nil
}

// convertLiteralValue converts a string value to the appropriate type and returns the correct pointer
func convertLiteralValue(value, valueType string) (interface{}, string, error) {
	switch valueType {
	case "string":
		return &value, "VarString", nil
	case "integer":
		if intVal, err := strconv.Atoi(value); err == nil {
			return &intVal, "Integer", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid integer", value)
		}
	case "number":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return &floatVal, "Number", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid number", value)
		}
	case "boolean":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return &boolVal, "Boolean", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid boolean", value)
		}
	case "date":
		if parsedDate, err := time.Parse(resourcedata.DateParseFormat, value); err == nil {
			return &parsedDate, "Date", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid date", value)
		}
	case "datetime":
		if parsedDateTime, err := time.Parse(DateTimeParseFormat, value); err == nil {
			return &parsedDateTime, "Datetime", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid datetime", value)
		}
	case "special":
		return &value, "Special", nil
	default:
		return nil, "", fmt.Errorf("unknown literal type: %s", valueType)
	}
}

// processItemsPositionally processes items with column order mapping
func processItemsPositionally(items []interface{}, maxCount int, processItem func(int, map[string]interface{}) error) error {
	for i, item := range items {
		if i >= maxCount {
			break
		}
		if itemMap, ok := item.(map[string]interface{}); ok {
			if err := processItem(i, itemMap); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildSdkInputColumns builds the SDK input columns from the provider schema
func buildSdkInputColumns(inputColumns []interface{}) (*[]platformclientv2.Decisiontableinputcolumnrequest, error) {
	if len(inputColumns) == 0 {
		return nil, nil
	}

	sdkInputColumns := make([]platformclientv2.Decisiontableinputcolumnrequest, 0, len(inputColumns))
	for _, inputColumn := range inputColumns {
		inputColumnMap, ok := inputColumn.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("inputColumn is not a map[string]interface{}")
		}
		sdkInputColumn := platformclientv2.Decisiontableinputcolumnrequest{}

		if defaultsToList, ok := inputColumnMap["defaults_to"].([]interface{}); ok {
			sdkInputColumn.DefaultsTo = buildDefaultsToFromProvider(defaultsToList)
		}

		if expressionList, ok := inputColumnMap["expression"].([]interface{}); ok && len(expressionList) > 0 {
			if expression, ok := expressionList[0].(map[string]interface{}); ok {
				sdkInputColumn.Expression = buildSdkExpression(expression)
			}
		}

		sdkInputColumns = append(sdkInputColumns, sdkInputColumn)
	}

	return &sdkInputColumns, nil
}

// buildSdkOutputColumns builds the SDK output columns from the provider schema
func buildSdkOutputColumns(outputColumns []interface{}) (*[]platformclientv2.Decisiontableoutputcolumnrequest, error) {
	if len(outputColumns) == 0 {
		return nil, nil
	}

	sdkOutputColumns := make([]platformclientv2.Decisiontableoutputcolumnrequest, 0, len(outputColumns))
	for _, outputColumn := range outputColumns {
		outputColumnMap, ok := outputColumn.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("outputColumn is not a map[string]interface{}")
		}
		sdkOutputColumn := platformclientv2.Decisiontableoutputcolumnrequest{}

		if defaultsToList, ok := outputColumnMap["defaults_to"].([]interface{}); ok {
			sdkOutputColumn.DefaultsTo = buildDefaultsToFromProvider(defaultsToList)
		}

		if valueList, ok := outputColumnMap["value"].([]interface{}); ok && len(valueList) > 0 {
			if value, ok := valueList[0].(map[string]interface{}); ok {
				sdkOutputColumn.Value = buildSdkValue(value)
			}
		}

		sdkOutputColumns = append(sdkOutputColumns, sdkOutputColumn)
	}

	return &sdkOutputColumns, nil
}

// buildSdkExpression builds the SDK expression from the provider schema
func buildSdkExpression(expression map[string]interface{}) *platformclientv2.Decisiontableinputcolumnexpression {
	sdkExpression := platformclientv2.Decisiontableinputcolumnexpression{}
	if contractualList, ok := expression["contractual"].([]interface{}); ok && len(contractualList) > 0 {
		if contractual, ok := contractualList[0].(map[string]interface{}); ok {
			sdkExpression.Contractual = buildSdkContractual(contractual)
		}
	}

	if comparator, ok := expression["comparator"].(string); ok {
		sdkExpression.Comparator = &comparator
	}

	return &sdkExpression
}

// buildSdkValue builds the SDK value from the provider schema
func buildSdkValue(value map[string]interface{}) *platformclientv2.Outputvalue {
	sdkValue := platformclientv2.Outputvalue{}

	if val, ok := value["schema_property_key"].(string); ok && val != "" {
		sdkValue.SchemaPropertyKey = &val
	}

	if properties, ok := value["properties"].([]interface{}); ok {
		sdkValue.Properties = buildSdkProperties(properties)
	}

	return &sdkValue
}

// buildSdkContractual builds the SDK contractual from the provider schema
func buildSdkContractual(contractual map[string]interface{}) **platformclientv2.Contractual {
	sdkContractual := platformclientv2.Contractual{}

	if val, ok := contractual["schema_property_key"].(string); ok && val != "" {
		sdkContractual.SchemaPropertyKey = &val
	}

	if nestedContractualList, ok := contractual["contractual"].([]interface{}); ok && len(nestedContractualList) > 0 {
		if nestedContractual, ok := nestedContractualList[0].(map[string]interface{}); ok {
			sdkContractual.Contractual = buildSdkContractual(nestedContractual)
		}
	}

	result := &sdkContractual
	return &result
}

// buildSdkProperties builds the SDK properties from the provider schema
func buildSdkProperties(properties []interface{}) *[]platformclientv2.Outputvalue {
	if len(properties) == 0 {
		return nil
	}

	sdkProperties := make([]platformclientv2.Outputvalue, 0)
	for _, property := range properties {
		propertyMap := property.(map[string]interface{})
		sdkProperty := platformclientv2.Outputvalue{}

		if val, ok := propertyMap["schema_property_key"].(string); ok && val != "" {
			sdkProperty.SchemaPropertyKey = &val
		}

		if nestedProperties, ok := propertyMap["properties"].([]interface{}); ok {
			sdkProperty.Properties = buildSdkProperties(nestedProperties)
		}

		sdkProperties = append(sdkProperties, sdkProperty)
	}

	return &sdkProperties
}

// buildSdkColumns builds the SDK columns from the provider schema
func buildSdkColumns(columns map[string]interface{}) (*platformclientv2.Createdecisiontablecolumnsrequest, error) {
	sdkColumns := &platformclientv2.Createdecisiontablecolumnsrequest{}

	if inputs, ok := columns["inputs"].([]interface{}); ok {
		inputColumns, err := buildSdkInputColumns(inputs)
		if err != nil {
			return nil, err
		}
		sdkColumns.Inputs = inputColumns
	}

	if outputs, ok := columns["outputs"].([]interface{}); ok {
		outputColumns, err := buildSdkOutputColumns(outputs)
		if err != nil {
			return nil, err
		}
		sdkColumns.Outputs = outputColumns
	}

	return sdkColumns, nil
}

// buildUpdateRequest builds the SDK update request from the provider schema
func buildUpdateRequest(d *schema.ResourceData) *platformclientv2.Updatedecisiontablerequest {
	updateRequest := &platformclientv2.Updatedecisiontablerequest{}

	if d.HasChange("name") {
		updateRequest.Name = platformclientv2.String(d.Get("name").(string))
	}

	if d.HasChange("description") {
		updateRequest.Description = platformclientv2.String(d.Get("description").(string))
	}

	return updateRequest
}

// convertSDKRowToUpdateRequest converts an SDK row to update request format
func convertSDKRowToUpdateRequest(sdkRow platformclientv2.Createdecisiontablerowrequest) *platformclientv2.Putdecisiontablerowrequest {
	updateRequest := &platformclientv2.Putdecisiontablerowrequest{}

	// Copy inputs if they exist
	if sdkRow.Inputs != nil {
		updateRequest.Inputs = sdkRow.Inputs
	}

	// Copy outputs if they exist
	if sdkRow.Outputs != nil {
		updateRequest.Outputs = sdkRow.Outputs
	}

	return updateRequest
}

// flattenColumns flattens the SDK columns response to provider format
func flattenColumns(sdkColumns *platformclientv2.Decisiontablecolumns) map[string]interface{} {
	if sdkColumns == nil {
		return make(map[string]interface{})
	}

	columns := make(map[string]interface{})

	if sdkColumns.Inputs != nil {
		inputs := flattenInputColumns(*sdkColumns.Inputs)
		columns["inputs"] = inputs
	}

	if sdkColumns.Outputs != nil {
		outputs := flattenOutputColumns(*sdkColumns.Outputs)
		columns["outputs"] = outputs
	}

	return columns
}

// flattenInputColumns flattens the SDK input columns to provider format
func flattenInputColumns(sdkInputColumns []platformclientv2.Decisiontableinputcolumn) []interface{} {
	inputs := make([]interface{}, 0, len(sdkInputColumns))
	for _, sdkInput := range sdkInputColumns {
		input := make(map[string]interface{})

		if sdkInput.Id != nil {
			input["id"] = *sdkInput.Id
		}

		// Handle both Special and Value fields for defaults_to
		if defaultsTo := flattenDefaultsTo(sdkInput.DefaultsTo); defaultsTo != nil {
			input["defaults_to"] = defaultsTo
		}

		if sdkInput.Expression != nil {
			expression := flattenExpression(sdkInput.Expression)
			input["expression"] = []interface{}{expression}
		}

		inputs = append(inputs, input)
	}
	return inputs
}

// flattenOutputColumns flattens the SDK output columns to provider format
func flattenOutputColumns(sdkOutputColumns []platformclientv2.Decisiontableoutputcolumn) []interface{} {
	outputs := make([]interface{}, 0, len(sdkOutputColumns))
	for _, sdkOutput := range sdkOutputColumns {
		output := make(map[string]interface{})

		if sdkOutput.Id != nil {
			output["id"] = *sdkOutput.Id
		}

		// Handle both Special and Value fields for defaults_to
		if defaultsTo := flattenDefaultsTo(sdkOutput.DefaultsTo); defaultsTo != nil {
			output["defaults_to"] = defaultsTo
		}

		if sdkOutput.Value != nil {
			value := flattenValue(sdkOutput.Value)
			output["value"] = []interface{}{value}
		}

		outputs = append(outputs, output)
	}
	return outputs
}

// flattenExpression flattens the SDK expression to provider format
func flattenExpression(sdkExpression *platformclientv2.Decisiontableinputcolumnexpression) map[string]interface{} {
	expression := make(map[string]interface{})

	if sdkExpression.Contractual != nil && *sdkExpression.Contractual != nil {
		contractual := flattenContractual(*sdkExpression.Contractual)
		expression["contractual"] = []interface{}{contractual}
	}

	if sdkExpression.Comparator != nil {
		expression["comparator"] = *sdkExpression.Comparator
	}

	return expression
}

// flattenValue flattens the SDK value to provider format
func flattenValue(sdkValue *platformclientv2.Outputvalue) map[string]interface{} {
	value := make(map[string]interface{})

	if sdkValue.SchemaPropertyKey != nil {
		value["schema_property_key"] = *sdkValue.SchemaPropertyKey
	}

	if sdkValue.Properties != nil {
		properties := flattenProperties(*sdkValue.Properties)
		value["properties"] = properties
	}

	return value
}

// flattenContractual flattens the SDK contractual to provider format
func flattenContractual(sdkContractual *platformclientv2.Contractual) map[string]interface{} {
	contractual := make(map[string]interface{})

	if sdkContractual.SchemaPropertyKey != nil {
		contractual["schema_property_key"] = *sdkContractual.SchemaPropertyKey
	}

	if sdkContractual.Contractual != nil && *sdkContractual.Contractual != nil {
		nestedContractual := flattenContractual(*sdkContractual.Contractual)
		contractual["contractual"] = []interface{}{nestedContractual}
	}

	return contractual
}

// flattenProperties flattens the SDK properties to provider format
func flattenProperties(sdkProperties []platformclientv2.Outputvalue) []interface{} {
	properties := make([]interface{}, 0)
	for _, sdkProperty := range sdkProperties {
		property := make(map[string]interface{})

		if sdkProperty.SchemaPropertyKey != nil {
			property["schema_property_key"] = *sdkProperty.SchemaPropertyKey
		}

		if sdkProperty.Properties != nil {
			nestedProperties := flattenProperties(*sdkProperty.Properties)
			property["properties"] = nestedProperties
		}

		properties = append(properties, property)
	}
	return properties
}

// buildCreateRequest builds a CreateDecisionTableRequest from provider resource data
func buildCreateRequest(d *schema.ResourceData) (*platformclientv2.Createdecisiontablerequest, error) {
	tableName := d.Get("name").(string)
	divisionId := d.Get("division_id").(string)
	schemaId := d.Get("schema_id").(string)
	columns := d.Get("columns").([]interface{})

	// Validate required fields
	if tableName == "" {
		return nil, fmt.Errorf("name is required")
	}
	if divisionId == "" {
		return nil, fmt.Errorf("division_id is required")
	}
	if schemaId == "" {
		return nil, fmt.Errorf("schema_id is required")
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("columns are required")
	}

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
	columnData := columns[0].(map[string]interface{})
	sdkColumns, err := buildSdkColumns(columnData)
	if err != nil {
		return nil, err
	}
	createRequest.Columns = sdkColumns

	return createRequest, nil
}

// extractColumnOrder extracts the order of input and output columns from SDK columns
func extractColumnOrder(sdkColumns *platformclientv2.Decisiontablecolumns) ([]string, []string) {
	var inputOrder []string
	var outputOrder []string

	if sdkColumns == nil {
		return inputOrder, outputOrder
	}

	// Extract input column IDs in order
	if sdkColumns.Inputs != nil {
		for _, input := range *sdkColumns.Inputs {
			if input.Id != nil {
				inputOrder = append(inputOrder, *input.Id)
			}
		}
	}

	// Extract output column IDs in order
	if sdkColumns.Outputs != nil {
		for _, output := range *sdkColumns.Outputs {
			if output.Id != nil {
				outputOrder = append(outputOrder, *output.Id)
			}
		}
	}

	return inputOrder, outputOrder
}

// extractLiteralFromList extracts the literal map from a provider list (MaxItems: 1)
func extractLiteralFromList(literalList interface{}) map[string]interface{} {
	if literalList == nil {
		return nil
	}

	if list, ok := literalList.([]interface{}); ok && len(list) > 0 {
		if literal, ok := list[0].(map[string]interface{}); ok {
			return literal
		}
	}

	return nil
}

// convertLiteralToSDK converts a provider literal to SDK format
func convertLiteralToSDK(literal map[string]interface{}) (*platformclientv2.Literal, error) {
	log.Printf("DEBUG: Input literal map: %+v", literal)

	// If literal block is empty (no fields), omit this literal (use column default)
	if len(literal) == 0 {
		log.Printf("DEBUG: Empty literal block, omitting literal (using column default)")
		return nil, nil
	}

	// Validate literal and extract values
	value, valueType, err := validateLiteral(literal)
	if err != nil {
		return nil, err
	}

	// If both value and type are empty, omit this literal (use column default)
	if value == "" && valueType == "" {
		return nil, nil
	}

	log.Printf("DEBUG: Converting literal - value: %s, type: %s", value, valueType)

	// Convert the value using the appropriate converter
	convertedValue, fieldName, err := convertLiteralValue(value, valueType)
	if err != nil {
		return nil, err
	}

	// Create SDK literal and set the field
	sdkLiteral := &platformclientv2.Literal{}
	sdkLiteral.SetField(fieldName, convertedValue)
	log.Printf("DEBUG: Set %s to: %v", fieldName, convertedValue)

	log.Printf("DEBUG: SetFieldNames after conversion: %+v", sdkLiteral.SetFieldNames)
	return sdkLiteral, nil
}

// converts an SDK literal to provider format
func convertSDKLiteralToProvider(sdkLiteral *platformclientv2.Literal) map[string]interface{} {
	literal := make(map[string]interface{})

	if sdkLiteral.VarString != nil {
		literal["value"] = *sdkLiteral.VarString
		literal["type"] = "string"
	} else if sdkLiteral.Integer != nil {
		literal["value"] = strconv.Itoa(*sdkLiteral.Integer)
		literal["type"] = "integer"
	} else if sdkLiteral.Number != nil {
		// Format number to preserve the original string representation
		// Use 'g' format to avoid zero-padding while preserving precision
		literal["value"] = strconv.FormatFloat(*sdkLiteral.Number, 'g', -1, 64)
		literal["type"] = "number"
	} else if sdkLiteral.Date != nil {
		literal["value"] = sdkLiteral.Date.Format(resourcedata.DateParseFormat)
		literal["type"] = "date"
	} else if sdkLiteral.Datetime != nil {
		literal["value"] = sdkLiteral.Datetime.Format(DateTimeParseFormat)
		literal["type"] = "datetime"
	} else if sdkLiteral.Boolean != nil {
		literal["value"] = strconv.FormatBool(*sdkLiteral.Boolean)
		literal["type"] = "boolean"
	} else if sdkLiteral.Special != nil {
		literal["value"] = *sdkLiteral.Special
		literal["type"] = "special"
	} else {
		// If no fields are set, return empty values to indicate use of column default
		literal["value"] = ""
		literal["type"] = ""
	}

	return literal
}

// convertSDKRowToProvider converts an SDK row to provider format
// This function ensures all columns are included, with empty literals for missing values
func convertSDKRowToProvider(sdkRow platformclientv2.Decisiontablerow, inputColumnIds []string, outputColumnIds []string) map[string]interface{} {
	providerRow := map[string]interface{}{
		"row_id":    sdkRow.Id,
		"row_index": sdkRow.RowIndex,
	}

	// Convert inputs using column order mapping
	if sdkRow.Inputs != nil {
		var inputs []interface{}

		// Create a map of columnId -> paramValue for easy lookup
		inputData := make(map[string]platformclientv2.Decisiontablerowparametervalue)
		for columnId, paramValue := range *sdkRow.Inputs {
			inputData[columnId] = paramValue
		}

		// Order inputs according to column order
		for _, columnId := range inputColumnIds {
			input := map[string]interface{}{
				"column_id": columnId,
			}

			if paramValue, exists := inputData[columnId]; exists && paramValue.Literal != nil {
				// Column has a literal value - convert it
				literalValue := convertSDKLiteralToProvider(paramValue.Literal)
				input["literal"] = []interface{}{literalValue}
			} else {
				// Column uses default value - export as empty string values
				input["literal"] = []interface{}{
					map[string]interface{}{
						"value": "",
						"type":  "",
					},
				}
			}

			inputs = append(inputs, input)
		}

		providerRow["inputs"] = inputs
	}

	// Convert outputs using column order mapping
	if sdkRow.Outputs != nil {
		var outputs []interface{}

		// Create a map of columnId -> paramValue for easy lookup
		outputData := make(map[string]platformclientv2.Decisiontablerowparametervalue)
		for columnId, paramValue := range *sdkRow.Outputs {
			outputData[columnId] = paramValue
		}

		// Order outputs according to column order
		for _, columnId := range outputColumnIds {
			output := map[string]interface{}{
				"column_id": columnId,
			}

			if paramValue, exists := outputData[columnId]; exists && paramValue.Literal != nil {
				// Column has a literal value - convert it
				literalValue := convertSDKLiteralToProvider(paramValue.Literal)
				output["literal"] = []interface{}{literalValue}
			} else {
				// Column uses default value - export as empty string values
				output["literal"] = []interface{}{
					map[string]interface{}{
						"value": "",
						"type":  "",
					},
				}
			}

			outputs = append(outputs, output)
		}

		providerRow["outputs"] = outputs
	}

	return providerRow
}

// converts row from provider to SDK format
func convertDecisionTableRowFromProviderToSDK(rowMap map[string]interface{}, inputColumnIds []string, outputColumnIds []string) (platformclientv2.Createdecisiontablerowrequest, error) {
	sdkRow := platformclientv2.Createdecisiontablerowrequest{}

	// Convert inputs using column order mapping
	if inputs, ok := rowMap["inputs"].([]interface{}); ok {
		sdkInputs := make(map[string]platformclientv2.Decisiontablerowparametervalue)
		hasExplicitInput := false

		if err := processItemsPositionally(inputs, len(inputColumnIds), func(i int, inputMap map[string]interface{}) error {
			columnId := inputColumnIds[i]

			// Extract literal if present
			if literal := extractLiteralFromList(inputMap["literal"]); literal != nil {
				sdkLiteral, err := convertLiteralToSDK(literal)
				if err != nil {
					return err
				}
				// Only include the input if we have a literal value
				if sdkLiteral != nil {
					paramValue := platformclientv2.Decisiontablerowparametervalue{
						Literal: sdkLiteral,
					}
					sdkInputs[columnId] = paramValue
					hasExplicitInput = true
				}
			}
			return nil
		}); err != nil {
			return platformclientv2.Createdecisiontablerowrequest{}, err
		}

		// Validate that at least one input has an explicit value
		if len(inputs) > 0 && !hasExplicitInput {
			return platformclientv2.Createdecisiontablerowrequest{}, fmt.Errorf("at least one input must have an explicit value (not just column defaults)")
		}

		if len(sdkInputs) > 0 {
			sdkRow.Inputs = &sdkInputs
		}
	}

	// Convert outputs using column order mapping
	if outputs, ok := rowMap["outputs"].([]interface{}); ok {
		sdkOutputs := make(map[string]platformclientv2.Decisiontablerowparametervalue)
		hasExplicitOutput := false

		if err := processItemsPositionally(outputs, len(outputColumnIds), func(i int, outputMap map[string]interface{}) error {
			columnId := outputColumnIds[i]

			// Extract literal if present
			if literal := extractLiteralFromList(outputMap["literal"]); literal != nil {
				sdkLiteral, err := convertLiteralToSDK(literal)
				if err != nil {
					return err
				}
				// Only include the output if we have a literal value
				if sdkLiteral != nil {
					paramValue := platformclientv2.Decisiontablerowparametervalue{
						Literal: sdkLiteral,
					}
					sdkOutputs[columnId] = paramValue
					hasExplicitOutput = true
				}
			}
			return nil
		}); err != nil {
			return platformclientv2.Createdecisiontablerowrequest{}, err
		}

		// Validate that at least one output has an explicit value
		if len(outputs) > 0 && !hasExplicitOutput {
			return platformclientv2.Createdecisiontablerowrequest{}, fmt.Errorf("at least one output must have an explicit value (not just column defaults)")
		}

		if len(sdkOutputs) > 0 {
			sdkRow.Outputs = &sdkOutputs
		}
	}

	return sdkRow, nil
}

// RowChange represents changes to be made to rows
type RowChange struct {
	adds    []map[string]interface{} // New rows to add
	updates []map[string]interface{} // Existing rows to update
	deletes []string                 // Row IDs to delete
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
	// Compare inputs - these are arrays in column order mapping
	inputs1, ok1 := row1["inputs"].([]interface{})
	inputs2, ok2 := row2["inputs"].([]interface{})
	if !ok1 || !ok2 || !arraysEqual(inputs1, inputs2) {
		log.Printf("DEBUG: Inputs differ - row1: %+v, row2: %+v", inputs1, inputs2)
		return false
	}

	// Compare outputs - these are arrays in column order mapping
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

	// Get column IDs in order for column order mapping
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

		// Convert to SDK format using column order mapping (same as creation)
		sdkRow, err := convertDecisionTableRowFromProviderToSDK(row, inputColumnIds, outputColumnIds)
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

	// Add new rows using column order mapping
	for i, row := range changes.adds {
		log.Printf("Adding new row %d/%d", i+1, len(changes.adds))
		sdkRow, err := convertDecisionTableRowFromProviderToSDK(row, inputColumnIds, outputColumnIds)
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
