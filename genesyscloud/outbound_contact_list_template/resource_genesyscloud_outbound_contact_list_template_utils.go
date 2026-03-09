package outbound_contact_list_template

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

type outboundContactListTemplateRawResponse struct {
	PhoneColumns []outboundContactListTemplateRawPhoneColumn `json:"phoneColumns"`
	EmailColumns []outboundContactListTemplateRawEmailColumn `json:"emailColumns"`
}

type outboundContactListTemplateRawPhoneColumn struct {
	ColumnName             *string `json:"columnName"`
	VarType                *string `json:"type"`
	CallableTimeColumn     *string `json:"callableTimeColumn"`
	CallableTimeColumnName *string `json:"callableTimeColumnName"`
}

type outboundContactListTemplateRawEmailColumn struct {
	ColumnName                *string `json:"columnName"`
	VarType                   *string `json:"type"`
	ContactableTimeColumn     *string `json:"contactableTimeColumn"`
	ContactableTimeColumnName *string `json:"contactableTimeColumnName"`
}

func buildOutboundContactListTemplateTimeColumnIndexFromSet(
	set *schema.Set,
	nameKey string,
	legacyKey string,
) map[string]string {
	if set == nil || set.Len() == 0 {
		return nil
	}

	idx := make(map[string]string, set.Len())
	for _, v := range set.List() {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		columnName, _ := m["column_name"].(string)
		varType, _ := m["type"].(string)
		if columnName == "" || varType == "" {
			continue
		}
		key := strings.ToLower(columnName) + "|" + strings.ToLower(varType)

		if val, ok := m[nameKey].(string); ok && val != "" {
			idx[key] = val
			continue
		}
		if val, ok := m[legacyKey].(string); ok && val != "" {
			idx[key] = val
		}
	}

	if len(idx) == 0 {
		return nil
	}
	return idx
}

func parseOutboundContactListTemplateRaw(respBody []byte) (phoneIdx map[string]string, emailIdx map[string]string) {
	if len(respBody) == 0 {
		return nil, nil
	}
	var raw outboundContactListTemplateRawResponse
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, nil
	}

	pIdx := make(map[string]string, len(raw.PhoneColumns))
	for _, c := range raw.PhoneColumns {
		if c.ColumnName == nil || c.VarType == nil {
			continue
		}
		key := strings.ToLower(*c.ColumnName) + "|" + strings.ToLower(*c.VarType)
		if c.CallableTimeColumnName != nil && *c.CallableTimeColumnName != "" {
			pIdx[key] = *c.CallableTimeColumnName
			continue
		}
		if c.CallableTimeColumn != nil && *c.CallableTimeColumn != "" {
			pIdx[key] = *c.CallableTimeColumn
		}
	}

	eIdx := make(map[string]string, len(raw.EmailColumns))
	for _, c := range raw.EmailColumns {
		if c.ColumnName == nil || c.VarType == nil {
			continue
		}
		key := strings.ToLower(*c.ColumnName) + "|" + strings.ToLower(*c.VarType)
		if c.ContactableTimeColumnName != nil && *c.ContactableTimeColumnName != "" {
			eIdx[key] = *c.ContactableTimeColumnName
			continue
		}
		if c.ContactableTimeColumn != nil && *c.ContactableTimeColumn != "" {
			eIdx[key] = *c.ContactableTimeColumn
		}
	}

	return pIdx, eIdx
}

func buildSdkOutboundContactListTemplateContactPhoneNumberColumnSlice(contactPhoneNumberColumn *schema.Set) *[]platformclientv2.Contactphonenumbercolumn {
	if contactPhoneNumberColumn == nil {
		return nil
	}
	sdkContactPhoneNumberColumnSlice := make([]platformclientv2.Contactphonenumbercolumn, 0)
	contactPhoneNumberColumnList := contactPhoneNumberColumn.List()
	for _, configPhoneColumn := range contactPhoneNumberColumnList {
		var sdkContactPhoneNumberColumn platformclientv2.Contactphonenumbercolumn
		contactPhoneNumberColumnMap := configPhoneColumn.(map[string]interface{})
		if columnName := contactPhoneNumberColumnMap["column_name"].(string); columnName != "" {
			sdkContactPhoneNumberColumn.ColumnName = &columnName
		}
		if varType := contactPhoneNumberColumnMap["type"].(string); varType != "" {
			sdkContactPhoneNumberColumn.VarType = &varType
		}
		if callableTimeColumnName, ok := contactPhoneNumberColumnMap["callable_time_column_name"].(string); ok && callableTimeColumnName != "" {
			sdkContactPhoneNumberColumn.CallableTimeColumn = &callableTimeColumnName
		} else if callableTimeColumn := contactPhoneNumberColumnMap["callable_time_column"].(string); callableTimeColumn != "" {
			sdkContactPhoneNumberColumn.CallableTimeColumn = &callableTimeColumn
		}

		sdkContactPhoneNumberColumnSlice = append(sdkContactPhoneNumberColumnSlice, sdkContactPhoneNumberColumn)
	}
	return &sdkContactPhoneNumberColumnSlice
}

func flattenSdkOutboundContactListTemplateContactPhoneNumberColumnSlice(
	contactPhoneNumberColumns []platformclientv2.Contactphonenumbercolumn,
	callableTimeColumnNameIndex map[string]string,
	fallbackTimeColumnIndex map[string]string,
) *schema.Set {
	if len(contactPhoneNumberColumns) == 0 {
		return nil
	}

	contactPhoneNumberColumnSet := schema.NewSet(hashOutboundContactListTemplatePhoneColumn, []interface{}{})
	for _, contactPhoneNumberColumn := range contactPhoneNumberColumns {
		contactPhoneNumberColumnMap := make(map[string]interface{})

		var key string
		if contactPhoneNumberColumn.ColumnName != nil {
			contactPhoneNumberColumnMap["column_name"] = *contactPhoneNumberColumn.ColumnName
			if contactPhoneNumberColumn.VarType != nil {
				key = strings.ToLower(*contactPhoneNumberColumn.ColumnName) + "|" + strings.ToLower(*contactPhoneNumberColumn.VarType)
			}
		}
		if contactPhoneNumberColumn.VarType != nil {
			contactPhoneNumberColumnMap["type"] = *contactPhoneNumberColumn.VarType
		}

		var tz string
		if contactPhoneNumberColumn.CallableTimeColumn != nil && *contactPhoneNumberColumn.CallableTimeColumn != "" {
			tz = *contactPhoneNumberColumn.CallableTimeColumn
		} else if callableTimeColumnNameIndex != nil && key != "" {
			if v, ok := callableTimeColumnNameIndex[key]; ok && v != "" {
				tz = v
			}
		}
		if tz == "" && fallbackTimeColumnIndex != nil && key != "" {
			if v, ok := fallbackTimeColumnIndex[key]; ok && v != "" {
				tz = v
			}
		}
		if tz != "" {
			// Keep legacy + new fields in sync while users migrate.
			contactPhoneNumberColumnMap["callable_time_column_name"] = tz
			contactPhoneNumberColumnMap["callable_time_column"] = tz
		}

		contactPhoneNumberColumnSet.Add(contactPhoneNumberColumnMap)
	}

	return contactPhoneNumberColumnSet
}

func buildSdkOutboundContactListContactEmailAddressColumnSlice(contactEmailAddressColumn *schema.Set) *[]platformclientv2.Emailcolumn {
	if contactEmailAddressColumn == nil {
		return nil
	}
	sdkContactEmailAddressColumnSlice := make([]platformclientv2.Emailcolumn, 0)
	contactEmailAddressColumnList := contactEmailAddressColumn.List()
	for _, configEmailColumn := range contactEmailAddressColumnList {
		var sdkContactEmailAddressColumn platformclientv2.Emailcolumn
		contactEmailAddressColumnMap := configEmailColumn.(map[string]interface{})
		if columnName := contactEmailAddressColumnMap["column_name"].(string); columnName != "" {
			sdkContactEmailAddressColumn.ColumnName = &columnName
		}
		if varType := contactEmailAddressColumnMap["type"].(string); varType != "" {
			sdkContactEmailAddressColumn.VarType = &varType
		}
		if contactableTimeColumnName, ok := contactEmailAddressColumnMap["contactable_time_column_name"].(string); ok && contactableTimeColumnName != "" {
			sdkContactEmailAddressColumn.ContactableTimeColumn = &contactableTimeColumnName
		} else if contactableTimeColumn := contactEmailAddressColumnMap["contactable_time_column"].(string); contactableTimeColumn != "" {
			sdkContactEmailAddressColumn.ContactableTimeColumn = &contactableTimeColumn
		}

		sdkContactEmailAddressColumnSlice = append(sdkContactEmailAddressColumnSlice, sdkContactEmailAddressColumn)
	}
	return &sdkContactEmailAddressColumnSlice
}

func flattenSdkOutboundContactListTemplateContactEmailAddressColumnSlice(
	contactEmailAddressColumns []platformclientv2.Emailcolumn,
	contactableTimeColumnNameIndex map[string]string,
	fallbackTimeColumnIndex map[string]string,
) *schema.Set {
	if len(contactEmailAddressColumns) == 0 {
		return nil
	}

	contactEmailAddressColumnSet := schema.NewSet(hashOutboundContactListTemplateEmailColumn, []interface{}{})
	for _, contactEmailAddressColumn := range contactEmailAddressColumns {
		contactEmailAddressColumnMap := make(map[string]interface{})

		var key string
		if contactEmailAddressColumn.ColumnName != nil {
			contactEmailAddressColumnMap["column_name"] = *contactEmailAddressColumn.ColumnName
			if contactEmailAddressColumn.VarType != nil {
				key = strings.ToLower(*contactEmailAddressColumn.ColumnName) + "|" + strings.ToLower(*contactEmailAddressColumn.VarType)
			}
		}
		if contactEmailAddressColumn.VarType != nil {
			contactEmailAddressColumnMap["type"] = *contactEmailAddressColumn.VarType
		}

		var tz string
		if contactEmailAddressColumn.ContactableTimeColumn != nil && *contactEmailAddressColumn.ContactableTimeColumn != "" {
			tz = *contactEmailAddressColumn.ContactableTimeColumn
		} else if contactableTimeColumnNameIndex != nil && key != "" {
			if v, ok := contactableTimeColumnNameIndex[key]; ok && v != "" {
				tz = v
			}
		}
		if tz == "" && fallbackTimeColumnIndex != nil && key != "" {
			if v, ok := fallbackTimeColumnIndex[key]; ok && v != "" {
				tz = v
			}
		}
		if tz != "" {
			// Keep legacy + new fields in sync while users migrate.
			contactEmailAddressColumnMap["contactable_time_column_name"] = tz
			contactEmailAddressColumnMap["contactable_time_column"] = tz
		}

		contactEmailAddressColumnSet.Add(contactEmailAddressColumnMap)
	}

	return contactEmailAddressColumnSet
}

func buildSdkOutboundContactListTemplateColumnDataTypeSpecifications(columnDataTypeSpecifications []interface{}) *[]platformclientv2.Columndatatypespecification {
	if len(columnDataTypeSpecifications) < 1 {
		return nil
	}

	sdkColumnDataTypeSpecificationsSlice := make([]platformclientv2.Columndatatypespecification, 0)

	for _, spec := range columnDataTypeSpecifications {
		if specMap, ok := spec.(map[string]interface{}); ok {
			var sdkColumnDataTypeSpecification platformclientv2.Columndatatypespecification
			if columnNameStr, ok := specMap["column_name"].(string); ok {
				sdkColumnDataTypeSpecification.ColumnName = &columnNameStr
			}
			if columnDataTypeStr, ok := specMap["column_data_type"].(string); ok && columnDataTypeStr != "" {
				sdkColumnDataTypeSpecification.ColumnDataType = &columnDataTypeStr
			}
			if minInt, ok := specMap["min"].(int); ok {
				sdkColumnDataTypeSpecification.Min = &minInt
			}
			if maxInt, ok := specMap["max"].(int); ok {
				sdkColumnDataTypeSpecification.Max = &maxInt
			}
			if maxLengthInt, ok := specMap["max_length"].(int); ok {
				sdkColumnDataTypeSpecification.MaxLength = &maxLengthInt
			}
			sdkColumnDataTypeSpecificationsSlice = append(sdkColumnDataTypeSpecificationsSlice, sdkColumnDataTypeSpecification)
		}
	}

	return &sdkColumnDataTypeSpecificationsSlice
}

func flattenSdkOutboundContactListTemplateColumnDataTypeSpecifications(columnDataTypeSpecifications []platformclientv2.Columndatatypespecification) []interface{} {
	if len(columnDataTypeSpecifications) == 0 {
		return nil
	}

	columnDataTypeSpecificationsSlice := make([]interface{}, 0)

	for _, s := range columnDataTypeSpecifications {
		columnDataTypeSpecification := make(map[string]interface{})
		columnDataTypeSpecification["column_name"] = *s.ColumnName

		if s.ColumnDataType != nil {
			columnDataTypeSpecification["column_data_type"] = *s.ColumnDataType
		}
		if s.Min != nil {
			columnDataTypeSpecification["min"] = *s.Min
		}
		if s.Max != nil {
			columnDataTypeSpecification["max"] = *s.Max
		}
		if s.MaxLength != nil {
			columnDataTypeSpecification["max_length"] = *s.MaxLength
		}

		columnDataTypeSpecificationsSlice = append(columnDataTypeSpecificationsSlice, columnDataTypeSpecification)
	}

	return columnDataTypeSpecificationsSlice
}

// type OutboundContactListInstance struct{
// }

// func (*OutboundContactListInstance) ResourceOutboundContactList() *schema.Resource {
// 	ResourceOutboundContactList() *schema.Resource
// }

func GeneratePhoneColumnsBlock(columnName, columnType, callableTimeColumn string) string {
	return fmt.Sprintf(`
	phone_columns {
		column_name          = "%s"
		type                 = "%s"
		callable_time_column_name = %s
	}
`, columnName, columnType, callableTimeColumn)
}

func GenerateOutboundContactListTemplate(
	resourceLabel string,
	name string,
	previewModeColumnName string,
	previewModeAcceptedValues []string,
	columnNames []string,
	automaticTimeZoneMapping string,
	zipCodeColumnName string,
	attemptLimitId string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	name                         = "%s"
	preview_mode_column_name     = %s
	preview_mode_accepted_values = [%s]
	column_names                 = [%s]
	automatic_time_zone_mapping  = %s
	zip_code_column_name         = %s
	attempt_limit_id             = %s
	%s
}
`, ResourceType, resourceLabel, name, previewModeColumnName, strings.Join(previewModeAcceptedValues, ", "),
		strings.Join(columnNames, ", "), automaticTimeZoneMapping, zipCodeColumnName, attemptLimitId, strings.Join(nestedBlocks, "\n"))
}

func GeneratePhoneColumnsDataTypeSpecBlock(columnName, columnDataType, min, max, maxLength string) string {
	return fmt.Sprintf(`
	column_data_type_specifications {
		column_name      = %s
		column_data_type = %s
		min              = %s
		max              = %s
		max_length       = %s
	}
	`, columnName, columnDataType, min, max, maxLength)
}

func GenerateEmailColumnsBlock(columnName, columnType, contactableTimeColumn string) string {
	return fmt.Sprintf(`
	email_columns {
		column_name             = "%s"
		type                    = "%s"
		contactable_time_column_name = %s
	}
`, columnName, columnType, contactableTimeColumn)
}
