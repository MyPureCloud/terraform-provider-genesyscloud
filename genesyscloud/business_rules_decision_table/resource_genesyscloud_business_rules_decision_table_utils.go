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

// buildDefaultsTo builds SDK defaults_to from Terraform schema
func buildDefaultsTo(defaultsToList []interface{}) *platformclientv2.Decisiontablecolumndefaultrowvalue {
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

// flattenDefaultsTo flattens SDK defaults_to to Terraform format
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

// validateLiteralInput validates that literal input has required fields
func validateLiteralInput(literal map[string]interface{}) (string, string, error) {
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
			return nil, "", fmt.Errorf("value '%s' is not a valid %s", value, "integer")
		}
	case "number":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return &floatVal, "Number", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid %s", value, "number")
		}
	case "boolean":
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return &boolVal, "Boolean", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid %s", value, "boolean")
		}
	case "date":
		if parsedDate, err := time.Parse(resourcedata.DateParseFormat, value); err == nil {
			return &parsedDate, "Date", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid %s", value, "date")
		}
	case "datetime":
		if parsedDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", value); err == nil {
			return &parsedDateTime, "Datetime", nil
		} else {
			return nil, "", fmt.Errorf("value '%s' is not a valid %s", value, "datetime")
		}
	case "special":
		return &value, "Special", nil
	default:
		return nil, "", fmt.Errorf("unknown literal type: %s", valueType)
	}
}

// validationError creates a standardized validation error message
func validationError(rowNum, itemNum int, itemType, message string) error {
	return fmt.Errorf("row %d %s %d: %s", rowNum, itemType, itemNum, message)
}

// validationErrorWithDetails creates a validation error with additional details
func validationErrorWithDetails(rowNum, itemNum int, itemType, message string, details ...interface{}) error {
	baseMsg := fmt.Sprintf("row %d %s %d: %s", rowNum, itemType, itemNum, message)
	if len(details) > 0 {
		baseMsg += fmt.Sprintf(" %v", details)
	}
	return fmt.Errorf("%s", baseMsg)
}

// processRowsWithValidation processes a list of rows with validation
func processRowsWithValidation(rows []interface{}, processRow func(int, map[string]interface{}) error) error {
	for i, row := range rows {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			return fmt.Errorf("row %d is not a valid map", i+1)
		}
		if err := processRow(i, rowMap); err != nil {
			return err
		}
	}
	return nil
}

// processItemsWithValidation processes a list of items with validation
func processItemsWithValidation(items []interface{}, rowNum int, itemType string, processItem func(int, map[string]interface{}) error) error {
	for j, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return validationError(rowNum, j+1, itemType, "is not a valid map")
		}
		if err := processItem(j, itemMap); err != nil {
			return err
		}
	}
	return nil
}

// processItemsWithError processes a list of items with error return
func processItemsWithError(items []interface{}, itemType string, processItem func(int, map[string]interface{}) error) error {
	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s %d is not a valid map", itemType, i+1)
		}
		if err := processItem(i, itemMap); err != nil {
			return err
		}
	}
	return nil
}

// processItemsPositionally processes items with positional mapping
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

// buildSdkInputColumns builds the SDK input columns from the Terraform schema
func buildSdkInputColumns(inputColumns []interface{}) *[]platformclientv2.Decisiontableinputcolumnrequest {
	if len(inputColumns) == 0 {
		return nil
	}

	sdkInputColumns := make([]platformclientv2.Decisiontableinputcolumnrequest, 0, len(inputColumns))
	for _, inputColumn := range inputColumns {
		inputColumnMap := inputColumn.(map[string]interface{})
		sdkInputColumn := platformclientv2.Decisiontableinputcolumnrequest{}

		if defaultsToList, ok := inputColumnMap["defaults_to"].([]interface{}); ok {
			sdkInputColumn.DefaultsTo = buildDefaultsTo(defaultsToList)
		}

		if expressionList, ok := inputColumnMap["expression"].([]interface{}); ok && len(expressionList) > 0 {
			if expression, ok := expressionList[0].(map[string]interface{}); ok {
				sdkInputColumn.Expression = buildSdkExpression(expression)
			}
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

	sdkOutputColumns := make([]platformclientv2.Decisiontableoutputcolumnrequest, 0, len(outputColumns))
	for _, outputColumn := range outputColumns {
		outputColumnMap := outputColumn.(map[string]interface{})
		sdkOutputColumn := platformclientv2.Decisiontableoutputcolumnrequest{}

		if defaultsToList, ok := outputColumnMap["defaults_to"].([]interface{}); ok {
			sdkOutputColumn.DefaultsTo = buildDefaultsTo(defaultsToList)
		}

		if valueList, ok := outputColumnMap["value"].([]interface{}); ok && len(valueList) > 0 {
			if value, ok := valueList[0].(map[string]interface{}); ok {
				sdkOutputColumn.Value = buildSdkValue(value)
			}
		}

		sdkOutputColumns = append(sdkOutputColumns, sdkOutputColumn)
	}

	return &sdkOutputColumns
}

// buildSdkExpression builds the SDK expression from the Terraform schema
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

// buildSdkValue builds the SDK value from the Terraform schema
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

// buildSdkContractual builds the SDK contractual from the Terraform schema
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

// buildSdkProperties builds the SDK properties from the Terraform schema
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

// flattenOutputColumns flattens the SDK output columns to Terraform format
func flattenOutputColumns(sdkOutputColumns []platformclientv2.Decisiontableoutputcolumn) []interface{} {
	outputs := make([]interface{}, 0)
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

	// Validate input and extract values
	value, valueType, err := validateLiteralInput(literal)
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
		// Use 'g' format to avoid zero-padding while preserving precision
		literal["value"] = strconv.FormatFloat(*sdkLiteral.Number, 'g', -1, 64)
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
	return processRowsWithValidation(rows, func(i int, rowMap map[string]interface{}) error {
		// Validate inputs
		if inputs, ok := rowMap["inputs"].([]interface{}); ok {
			if err := processItemsWithValidation(inputs, i+1, "input", func(j int, inputMap map[string]interface{}) error {
				return validateInputSchemaKey(inputMap, inputKeys, i+1, j+1)
			}); err != nil {
				return err
			}
		}

		// Validate outputs
		if outputs, ok := rowMap["outputs"].([]interface{}); ok {
			if err := processItemsWithValidation(outputs, i+1, "output", func(j int, outputMap map[string]interface{}) error {
				return validateOutputSchemaKey(outputMap, outputKeys, i+1, j+1)
			}); err != nil {
				return err
			}
		}

		return nil
	})
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
		return validationError(rowNum, inputNum, "input", "schema_property_key is required")
	}

	comparator, _ := inputMap["comparator"].(string)

	// Check if schema property key exists
	availableComparators, exists := inputKeys[schemaPropertyKey]
	if !exists {
		availableKeys := make([]string, 0, len(inputKeys))
		for key := range inputKeys {
			availableKeys = append(availableKeys, key)
		}
		return validationErrorWithDetails(rowNum, inputNum, "input",
			fmt.Sprintf("schema_property_key '%s' not found in input columns. Available keys: %v", schemaPropertyKey, availableKeys))
	}

	// Check if comparator is valid for this schema property key
	if len(availableComparators) > 1 {
		// Multiple comparators available, user must specify one
		if comparator == "" {
			return validationErrorWithDetails(rowNum, inputNum, "input",
				fmt.Sprintf("comparator is required for schema_property_key '%s' (available: %v)", schemaPropertyKey, availableComparators))
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
			return validationErrorWithDetails(rowNum, inputNum, "input",
				fmt.Sprintf("invalid comparator '%s' for schema_property_key '%s' (available: %v)", comparator, schemaPropertyKey, availableComparators))
		}
	} else if len(availableComparators) == 1 && availableComparators[0] != "" {
		// Only one comparator available, validate it matches
		if comparator != "" && comparator != availableComparators[0] {
			return validationErrorWithDetails(rowNum, inputNum, "input",
				fmt.Sprintf("invalid comparator '%s' for schema_property_key '%s' (expected: '%s')", comparator, schemaPropertyKey, availableComparators[0]))
		}
	}

	return nil
}

// validateOutputSchemaKey validates a single output schema property key
func validateOutputSchemaKey(outputMap map[string]interface{}, outputKeys map[string][]string, rowNum, outputNum int) error {
	schemaPropertyKey, ok := outputMap["schema_property_key"].(string)
	if !ok || schemaPropertyKey == "" {
		return validationError(rowNum, outputNum, "output", "schema_property_key is required")
	}

	// Check if schema property key exists
	_, exists := outputKeys[schemaPropertyKey]
	if !exists {
		availableKeys := make([]string, 0, len(outputKeys))
		for key := range outputKeys {
			availableKeys = append(availableKeys, key)
		}
		return validationErrorWithDetails(rowNum, outputNum, "output",
			fmt.Sprintf("schema_property_key '%s' not found in output columns. Available keys: %v", schemaPropertyKey, availableKeys))
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
		if err := processItemsWithError(inputs, "input column", func(i int, inputMap map[string]interface{}) error {
			sdkInput := platformclientv2.Decisiontableinputcolumn{
				Id: platformclientv2.String(fmt.Sprintf("input-column-%d", i+1)),
			}

			// Convert expression
			if expressionList, ok := inputMap["expression"].([]interface{}); ok && len(expressionList) > 0 {
				if exprMap, ok := expressionList[0].(map[string]interface{}); ok {
					sdkExpr := &platformclientv2.Decisiontableinputcolumnexpression{}

					// Convert contractual
					if contractualList, ok := exprMap["contractual"].([]interface{}); ok && len(contractualList) > 0 {
						if contractualMap, ok := contractualList[0].(map[string]interface{}); ok {
							if val, ok := contractualMap["schema_property_key"].(string); ok && val != "" {
								contractualObj := &platformclientv2.Contractual{
									SchemaPropertyKey: &val,
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
			return nil
		}); err != nil {
			return nil, err
		}
		sdkColumns.Inputs = &sdkInputs
	}

	// Convert output columns
	if outputs, ok := columnsMap["outputs"].([]interface{}); ok {
		sdkOutputs := make([]platformclientv2.Decisiontableoutputcolumn, 0, len(outputs))
		if err := processItemsWithError(outputs, "output column", func(i int, outputMap map[string]interface{}) error {
			sdkOutput := platformclientv2.Decisiontableoutputcolumn{
				Id: platformclientv2.String(fmt.Sprintf("output-column-%d", i+1)),
			}

			// Convert value
			if valueList, ok := outputMap["value"].([]interface{}); ok && len(valueList) > 0 {
				if valueMap, ok := valueList[0].(map[string]interface{}); ok {
					sdkValue := &platformclientv2.Outputvalue{}

					if val, ok := valueMap["schema_property_key"].(string); ok && val != "" {
						sdkValue.SchemaPropertyKey = &val
					}

					// Handle nested properties if present
					if properties, ok := valueMap["properties"].([]interface{}); ok {
						sdkProperties, err := convertTerraformPropertiesToSDK(properties)
						if err != nil {
							return fmt.Errorf("failed to convert properties for output column %d: %s", i+1, err)
						}
						sdkValue.Properties = sdkProperties
					}

					sdkOutput.Value = sdkValue
				}
			}

			sdkOutputs = append(sdkOutputs, sdkOutput)
			return nil
		}); err != nil {
			return nil, err
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
	if err := processItemsWithError(properties, "property", func(i int, propMap map[string]interface{}) error {
		sdkProp := platformclientv2.Outputvalue{}

		if val, ok := propMap["schema_property_key"].(string); ok && val != "" {
			sdkProp.SchemaPropertyKey = &val
		}

		// Handle nested properties recursively
		if nestedProps, ok := propMap["properties"].([]interface{}); ok {
			nestedSdkProps, err := convertTerraformPropertiesToSDK(nestedProps)
			if err != nil {
				return fmt.Errorf("failed to convert nested properties: %s", err)
			}
			sdkProp.Properties = nestedSdkProps
		}

		sdkProperties = append(sdkProperties, sdkProp)
		return nil
	}); err != nil {
		return nil, err
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

	// Convert outputs using positional mapping
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
