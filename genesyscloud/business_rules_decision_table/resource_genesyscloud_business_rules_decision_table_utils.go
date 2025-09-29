package business_rules_decision_table

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

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

// buildUpdateRequest builds the SDK update request from the Terraform schema
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

// flattenColumns flattens the SDK columns response to Terraform format
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

// flattenInputColumns flattens the SDK input columns to Terraform format
func flattenInputColumns(sdkInputColumns []platformclientv2.Decisiontableinputcolumn) []interface{} {
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
			expression := flattenExpression(sdkInput.Expression)
			input["expression"] = []interface{}{expression}
		}

		inputs = append(inputs, input)
	}
	return inputs
}

// flattenOutputColumns flattens the SDK output columns to Terraform format
func flattenOutputColumns(sdkOutputColumns []platformclientv2.Decisiontableoutputcolumn) []interface{} {
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
			value := flattenValue(sdkOutput.Value)
			output["value"] = []interface{}{value}
		}

		outputs = append(outputs, output)
	}
	return outputs
}

// flattenExpression flattens the SDK expression to Terraform format
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

// flattenValue flattens the SDK value to Terraform format
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

// flattenContractual flattens the SDK contractual to Terraform format
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

// flattenProperties flattens the SDK properties to Terraform format
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

// extractLiteralFromList extracts the literal map from a Terraform list (MaxItems: 1)
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

// convertLiteralToSDK converts a Terraform literal to SDK format
func convertLiteralToSDK(literal map[string]interface{}) (*platformclientv2.Literal, error) {
	log.Printf("DEBUG: Input literal map: %+v", literal)

	// If literal block is empty (no fields), omit this literal (use column default)
	if len(literal) == 0 {
		log.Printf("DEBUG: Empty literal block, omitting literal (using column default)")
		return nil, nil
	}

	sdkLiteral := &platformclientv2.Literal{}

	value, valueOk := literal["value"].(string)
	valueType, typeOk := literal["type"].(string)

	// If both value and type are missing or empty, omit this literal (use column default)
	if (!valueOk || value == "") && (!typeOk || valueType == "") {
		log.Printf("DEBUG: Both value and type are missing or empty, omitting literal (using column default)")
		return nil, nil
	}

	// If both value and type are empty strings, omit this literal (use column default)
	if value == "" && valueType == "" {
		log.Printf("DEBUG: Both value and type are empty strings, omitting literal (using column default)")
		return nil, nil
	}

	// If only one is provided, that's an error
	if (!valueOk || value == "") && (typeOk && valueType != "") {
		return nil, fmt.Errorf("value is required when type is specified")
	}
	if (valueOk && value != "") && (!typeOk || valueType == "") {
		return nil, fmt.Errorf("type is required when value is specified")
	}

	// If value is not empty but type is empty, that's an error
	if value != "" && valueType == "" {
		return nil, fmt.Errorf("type cannot be empty when value is specified")
	}

	log.Printf("DEBUG: Converting literal - value: %s, type: %s", value, valueType)

	switch valueType {
	case "string":
		sdkLiteral.SetField("VarString", &value)
		log.Printf("DEBUG: Set VarString to: %s", value)
	case "integer":
		if intVal, err := strconv.Atoi(value); err == nil {
			sdkLiteral.SetField("Integer", &intVal)
			log.Printf("DEBUG: Set Integer to: %d", intVal)
		} else {
			return nil, fmt.Errorf("value '%s' is not a valid integer", value)
		}
	case "number":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			sdkLiteral.SetField("Number", &floatVal)
			log.Printf("DEBUG: Set Number to: %f", floatVal)
		} else {
			return nil, fmt.Errorf("value '%s' is not a valid number", value)
		}
	case "boolean":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			sdkLiteral.SetField("Boolean", &boolVal)
			log.Printf("DEBUG: Set Boolean to: %t", boolVal)
		} else {
			return nil, fmt.Errorf("value '%s' is not a valid boolean", value)
		}
	case "date":
		if parsedDate, err := time.Parse(resourcedata.DateParseFormat, value); err == nil {
			sdkLiteral.SetField("Date", &parsedDate)
			log.Printf("DEBUG: Set Date to: %s", parsedDate.Format(resourcedata.DateParseFormat))
		} else {
			return nil, fmt.Errorf("value '%s' is not a valid date", value)
		}
	case "datetime":
		if parsedDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", value); err == nil {
			sdkLiteral.SetField("Datetime", &parsedDateTime)
			log.Printf("DEBUG: Set Datetime to: %s", parsedDateTime.Format("2006-01-02T15:04:05.000Z"))
		} else {
			return nil, fmt.Errorf("value '%s' is not a valid datetime", value)
		}
	case "special":
		sdkLiteral.SetField("Special", &value)
		log.Printf("DEBUG: Set Special to: %s", value)
	default:
		return nil, fmt.Errorf("unknown literal type: %s", valueType)
	}

	log.Printf("DEBUG: SetFieldNames after conversion: %+v", sdkLiteral.SetFieldNames)
	return sdkLiteral, nil
}

// convertLiteralToTerraform converts an SDK literal to Terraform format
func convertLiteralToTerraform(sdkLiteral *platformclientv2.Literal) map[string]interface{} {
	literal := make(map[string]interface{})

	if sdkLiteral.VarString != nil {
		literal["value"] = *sdkLiteral.VarString
		literal["type"] = "string"
	} else if sdkLiteral.Integer != nil {
		literal["value"] = strconv.Itoa(*sdkLiteral.Integer)
		literal["type"] = "integer"
	} else if sdkLiteral.Number != nil {
		// Format number to preserve the original string representation
		// Use 'f' format with 1 decimal place to ensure consistency with "999.0" format
		literal["value"] = strconv.FormatFloat(*sdkLiteral.Number, 'f', 1, 64)
		literal["type"] = "number"
	} else if sdkLiteral.Date != nil {
		literal["value"] = sdkLiteral.Date.Format(resourcedata.DateParseFormat)
		literal["type"] = "date"
	} else if sdkLiteral.Datetime != nil {
		literal["value"] = sdkLiteral.Datetime.Format("2006-01-02T15:04:05.000Z")
		literal["type"] = "datetime"
	} else if sdkLiteral.Boolean != nil {
		literal["value"] = strconv.FormatBool(*sdkLiteral.Boolean)
		literal["type"] = "boolean"
	} else if sdkLiteral.Special != nil {
		literal["value"] = *sdkLiteral.Special
		literal["type"] = "special"
	}

	return literal
}

// convertSDKRowToTerraformSimple converts an SDK row to Terraform format using positional mapping
// This function ensures all columns are included, with empty literals for missing values
func convertSDKRowToTerraformSimple(sdkRow platformclientv2.Decisiontablerow, inputColumnIds []string, outputColumnIds []string) map[string]interface{} {
	terraformRow := map[string]interface{}{
		"row_id":    sdkRow.Id,
		"row_index": sdkRow.RowIndex,
	}

	// Convert inputs using positional mapping
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
				literalValue := convertLiteralToTerraform(paramValue.Literal)
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

		terraformRow["inputs"] = inputs
	}

	// Convert outputs using positional mapping
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
				literalValue := convertLiteralToTerraform(paramValue.Literal)
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

		terraformRow["outputs"] = outputs
	}

	return terraformRow
}

// validateSchemaPropertyKeys validates that all schema property keys in rows exist in the column definitions
func validateSchemaPropertyKeys(columns *platformclientv2.Decisiontablecolumns, rows []interface{}) error {
	if columns == nil {
		return fmt.Errorf("columns are required for validation")
	}

	// Build maps of available schema property keys and their comparators
	inputKeys, outputKeys, err := buildSchemaKeyMaps(columns)
	if err != nil {
		return fmt.Errorf("failed to build schema key maps: %s", err)
	}

	// Validate each row
	for i, row := range rows {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			return fmt.Errorf("row %d is not a valid map", i+1)
		}

		// Validate inputs
		if inputs, ok := rowMap["inputs"].([]interface{}); ok {
			for j, input := range inputs {
				inputMap, ok := input.(map[string]interface{})
				if !ok {
					return fmt.Errorf("row %d input %d is not a valid map", i+1, j+1)
				}

				if err := validateInputSchemaKey(inputMap, inputKeys, i+1, j+1); err != nil {
					return err
				}
			}
		}

		// Validate outputs
		if outputs, ok := rowMap["outputs"].([]interface{}); ok {
			for j, output := range outputs {
				outputMap, ok := output.(map[string]interface{})
				if !ok {
					return fmt.Errorf("row %d output %d is not a valid map", i+1, j+1)
				}

				if err := validateOutputSchemaKey(outputMap, outputKeys, i+1, j+1); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// buildSchemaKeyMaps builds maps of available schema property keys and their comparators
func buildSchemaKeyMaps(columns *platformclientv2.Decisiontablecolumns) (map[string][]string, map[string][]string, error) {
	inputKeys := make(map[string][]string)
	outputKeys := make(map[string][]string)

	// Build input column map
	if columns.Inputs != nil {
		for _, input := range *columns.Inputs {
			if input.Expression == nil || input.Expression.Contractual == nil {
				continue
			}

			schemaPropertyKey := *(*input.Expression.Contractual).SchemaPropertyKey
			comparator := ""
			if input.Expression.Comparator != nil {
				comparator = *input.Expression.Comparator
			}

			if inputKeys[schemaPropertyKey] == nil {
				inputKeys[schemaPropertyKey] = []string{}
			}
			inputKeys[schemaPropertyKey] = append(inputKeys[schemaPropertyKey], comparator)
		}
	}

	// Build output column map
	if columns.Outputs != nil {
		for _, output := range *columns.Outputs {
			if output.Value == nil {
				continue
			}

			schemaPropertyKey := *output.Value.SchemaPropertyKey
			// Outputs don't have comparators, so we use empty string
			comparator := ""

			if outputKeys[schemaPropertyKey] == nil {
				outputKeys[schemaPropertyKey] = []string{}
			}
			outputKeys[schemaPropertyKey] = append(outputKeys[schemaPropertyKey], comparator)
		}
	}

	return inputKeys, outputKeys, nil
}

// validateInputSchemaKey validates a single input schema property key
func validateInputSchemaKey(inputMap map[string]interface{}, inputKeys map[string][]string, rowNum, inputNum int) error {
	schemaPropertyKey, ok := inputMap["schema_property_key"].(string)
	if !ok || schemaPropertyKey == "" {
		return fmt.Errorf("row %d input %d: schema_property_key is required", rowNum, inputNum)
	}

	comparator, _ := inputMap["comparator"].(string)

	// Check if schema property key exists
	availableComparators, exists := inputKeys[schemaPropertyKey]
	if !exists {
		availableKeys := make([]string, 0, len(inputKeys))
		for key := range inputKeys {
			availableKeys = append(availableKeys, key)
		}
		return fmt.Errorf("row %d input %d: schema_property_key '%s' not found in input columns. Available keys: %v",
			rowNum, inputNum, schemaPropertyKey, availableKeys)
	}

	// Check if comparator is valid for this schema property key
	if len(availableComparators) > 1 {
		// Multiple comparators available, user must specify one
		if comparator == "" {
			return fmt.Errorf("row %d input %d: comparator is required for schema_property_key '%s' (available: %v)",
				rowNum, inputNum, schemaPropertyKey, availableComparators)
		}

		// Check if the specified comparator is valid
		validComparator := false
		for _, validComp := range availableComparators {
			if validComp == comparator {
				validComparator = true
				break
			}
		}

		if !validComparator {
			return fmt.Errorf("row %d input %d: invalid comparator '%s' for schema_property_key '%s' (available: %v)",
				rowNum, inputNum, comparator, schemaPropertyKey, availableComparators)
		}
	} else if len(availableComparators) == 1 && availableComparators[0] != "" {
		// Only one comparator available, validate it matches
		if comparator != "" && comparator != availableComparators[0] {
			return fmt.Errorf("row %d input %d: invalid comparator '%s' for schema_property_key '%s' (expected: '%s')",
				rowNum, inputNum, comparator, schemaPropertyKey, availableComparators[0])
		}
	}

	return nil
}

// validateOutputSchemaKey validates a single output schema property key
func validateOutputSchemaKey(outputMap map[string]interface{}, outputKeys map[string][]string, rowNum, outputNum int) error {
	schemaPropertyKey, ok := outputMap["schema_property_key"].(string)
	if !ok || schemaPropertyKey == "" {
		return fmt.Errorf("row %d output %d: schema_property_key is required", rowNum, outputNum)
	}

	// Check if schema property key exists
	_, exists := outputKeys[schemaPropertyKey]
	if !exists {
		availableKeys := make([]string, 0, len(outputKeys))
		for key := range outputKeys {
			availableKeys = append(availableKeys, key)
		}
		return fmt.Errorf("row %d output %d: schema_property_key '%s' not found in output columns. Available keys: %v",
			rowNum, outputNum, schemaPropertyKey, availableKeys)
	}

	// Outputs don't have comparators, so we don't validate them
	return nil
}

// convertTerraformColumnsToSDK converts Terraform column configuration to SDK format for validation
func convertTerraformColumnsToSDK(columnsMap map[string]interface{}) (*platformclientv2.Decisiontablecolumns, error) {
	sdkColumns := &platformclientv2.Decisiontablecolumns{}

	// Convert input columns
	if inputs, ok := columnsMap["inputs"].([]interface{}); ok {
		sdkInputs := make([]platformclientv2.Decisiontableinputcolumn, 0, len(inputs))
		for i, input := range inputs {
			inputMap, ok := input.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("input column %d is not a valid map", i+1)
			}

			sdkInput := platformclientv2.Decisiontableinputcolumn{
				Id: platformclientv2.String(fmt.Sprintf("input-column-%d", i+1)),
			}

			// Convert expression
			if expression, ok := inputMap["expression"].([]interface{}); ok && len(expression) > 0 {
				if exprMap, ok := expression[0].(map[string]interface{}); ok {
					sdkExpr := &platformclientv2.Decisiontableinputcolumnexpression{}

					// Convert contractual
					if contractual, ok := exprMap["contractual"].([]interface{}); ok && len(contractual) > 0 {
						if contractualMap, ok := contractual[0].(map[string]interface{}); ok {
							if schemaPropertyKey, ok := contractualMap["schema_property_key"].(string); ok {
								contractualObj := &platformclientv2.Contractual{
									SchemaPropertyKey: &schemaPropertyKey,
								}
								sdkExpr.Contractual = &contractualObj
							}
						}
					}

					// Convert comparator
					if comparator, ok := exprMap["comparator"].(string); ok {
						sdkExpr.Comparator = &comparator
					}

					sdkInput.Expression = sdkExpr
				}
			}

			sdkInputs = append(sdkInputs, sdkInput)
		}
		sdkColumns.Inputs = &sdkInputs
	}

	// Convert output columns
	if outputs, ok := columnsMap["outputs"].([]interface{}); ok {
		sdkOutputs := make([]platformclientv2.Decisiontableoutputcolumn, 0, len(outputs))
		for i, output := range outputs {
			outputMap, ok := output.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("output column %d is not a valid map", i+1)
			}

			sdkOutput := platformclientv2.Decisiontableoutputcolumn{
				Id: platformclientv2.String(fmt.Sprintf("output-column-%d", i+1)),
			}

			// Convert value
			if value, ok := outputMap["value"].([]interface{}); ok && len(value) > 0 {
				if valueMap, ok := value[0].(map[string]interface{}); ok {
					sdkValue := &platformclientv2.Outputvalue{}

					if schemaPropertyKey, ok := valueMap["schema_property_key"].(string); ok {
						sdkValue.SchemaPropertyKey = &schemaPropertyKey
					}

					// Handle nested properties if present
					if properties, ok := valueMap["properties"].([]interface{}); ok {
						sdkProperties, err := convertTerraformPropertiesToSDK(properties)
						if err != nil {
							return nil, fmt.Errorf("failed to convert properties for output column %d: %s", i+1, err)
						}
						sdkValue.Properties = sdkProperties
					}

					sdkOutput.Value = sdkValue
				}
			}

			sdkOutputs = append(sdkOutputs, sdkOutput)
		}
		sdkColumns.Outputs = &sdkOutputs
	}

	return sdkColumns, nil
}

// convertTerraformPropertiesToSDK converts Terraform properties to SDK format recursively
func convertTerraformPropertiesToSDK(properties []interface{}) (*[]platformclientv2.Outputvalue, error) {
	if len(properties) == 0 {
		return nil, nil
	}

	sdkProperties := make([]platformclientv2.Outputvalue, 0, len(properties))
	for i, prop := range properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("property %d is not a valid map", i+1)
		}

		sdkProp := platformclientv2.Outputvalue{}

		if schemaPropertyKey, ok := propMap["schema_property_key"].(string); ok {
			sdkProp.SchemaPropertyKey = &schemaPropertyKey
		}

		// Handle nested properties recursively
		if nestedProps, ok := propMap["properties"].([]interface{}); ok && len(nestedProps) > 0 {
			nestedSdkProps, err := convertTerraformPropertiesToSDK(nestedProps)
			if err != nil {
				return nil, fmt.Errorf("failed to convert nested properties: %s", err)
			}
			sdkProp.Properties = nestedSdkProps
		}

		sdkProperties = append(sdkProperties, sdkProp)
	}

	return &sdkProperties, nil
}

// convertTerraformRowToSDKPositional converts a Terraform row to SDK format using positional mapping
func convertTerraformRowToSDKPositional(rowMap map[string]interface{}, inputColumnIds []string, outputColumnIds []string) (platformclientv2.Createdecisiontablerowrequest, error) {
	sdkRow := platformclientv2.Createdecisiontablerowrequest{}

	// Convert inputs using positional mapping
	if inputs, ok := rowMap["inputs"].([]interface{}); ok {
		sdkInputs := make(map[string]platformclientv2.Decisiontablerowparametervalue)
		hasExplicitInput := false

		for i, inputItem := range inputs {
			if i >= len(inputColumnIds) {
				break // Don't process more inputs than we have columns
			}

			if inputMap, ok := inputItem.(map[string]interface{}); ok {
				columnId := inputColumnIds[i]

				// Extract literal if present
				if literal := extractLiteralFromList(inputMap["literal"]); literal != nil {
					sdkLiteral, err := convertLiteralToSDK(literal)
					if err != nil {
						return platformclientv2.Createdecisiontablerowrequest{}, err
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
			}
		}

		// Validate that at least one input has an explicit value
		if len(inputs) > 0 && !hasExplicitInput {
			return platformclientv2.Createdecisiontablerowrequest{}, fmt.Errorf("at least one input must have an explicit value (not just column defaults)")
		}

		if len(sdkInputs) > 0 {
			sdkRow.Inputs = &sdkInputs
		}
	}

	// Convert outputs using positional mapping
	if outputs, ok := rowMap["outputs"].([]interface{}); ok {
		sdkOutputs := make(map[string]platformclientv2.Decisiontablerowparametervalue)
		hasExplicitOutput := false

		for i, outputItem := range outputs {
			if i >= len(outputColumnIds) {
				break // Don't process more outputs than we have columns
			}

			if outputMap, ok := outputItem.(map[string]interface{}); ok {
				columnId := outputColumnIds[i]

				// Extract literal if present
				if literal := extractLiteralFromList(outputMap["literal"]); literal != nil {
					sdkLiteral, err := convertLiteralToSDK(literal)
					if err != nil {
						return platformclientv2.Createdecisiontablerowrequest{}, err
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
			}
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
