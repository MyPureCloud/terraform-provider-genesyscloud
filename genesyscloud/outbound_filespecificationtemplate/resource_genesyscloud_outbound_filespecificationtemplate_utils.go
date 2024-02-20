package outbound_filespecificationtemplate

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func buildSdkOutboundFileSpecificationTemplate(d *schema.ResourceData) platformclientv2.Filespecificationtemplate {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	format := d.Get("format").(string)
	numberOfHeaderLinesSkipped := d.Get("number_of_header_lines_skipped").(int)
	numberOfTrailerLinesSkipped := d.Get("number_of_trailer_lines_skipped").(int)
	header := d.Get("header").(bool)
	delimiter := d.Get("delimiter").(string)
	delimiterValue := d.Get("delimiter_value").(string)

	sdkFileSpecificationTemplate := platformclientv2.Filespecificationtemplate{
		NumberOfHeadingLinesSkipped:  &numberOfHeaderLinesSkipped,
		NumberOfTrailingLinesSkipped: &numberOfTrailerLinesSkipped,
		Header:                       &header,
		ColumnInformation:            buildSdkOutboundFileSpecificationTemplateColumnInformationSlice(d.Get("column_information").([]interface{})),
		PreprocessingRules:           buildSdkOutboundFileSpecificationTemplatePreprocessingRulesSlice(d.Get("preprocessing_rule").([]interface{})),
	}

	if name != "" {
		sdkFileSpecificationTemplate.Name = &name
	}
	if description != "" {
		sdkFileSpecificationTemplate.Description = &description
	}
	if format != "" {
		sdkFileSpecificationTemplate.Format = &format
	}
	if delimiter != "" {
		sdkFileSpecificationTemplate.Delimiter = &delimiter
	}
	sdkFileSpecificationTemplate.DelimiterValue = &delimiterValue
	return sdkFileSpecificationTemplate
}

func buildSdkOutboundFileSpecificationTemplateColumnInformationSlice(columnInformation []interface{}) *[]platformclientv2.Column {
	if columnInformation == nil || len(columnInformation) < 1 {
		return nil
	}
	sdkColumnInformationSlice := make([]platformclientv2.Column, 0)
	for _, columnInfo := range columnInformation {
		if columnInfoMap, ok := columnInfo.(map[string]interface{}); ok {
			var sdkColumnInformation platformclientv2.Column

			if columnNameStr, ok := columnInfoMap["column_name"].(string); ok {
				sdkColumnInformation.ColumnName = &columnNameStr
			}
			if columnNumberInt, ok := columnInfoMap["column_number"].(int); ok {
				sdkColumnInformation.ColumnNumber = &columnNumberInt
			}
			if startPositionInt, ok := columnInfoMap["start_position"].(int); ok {
				sdkColumnInformation.StartPosition = &startPositionInt
			}
			if lengthInt, ok := columnInfoMap["length"].(int); ok {
				sdkColumnInformation.Length = &lengthInt
			}
			sdkColumnInformationSlice = append(sdkColumnInformationSlice, sdkColumnInformation)
		}
	}
	return &sdkColumnInformationSlice
}

func buildSdkOutboundFileSpecificationTemplatePreprocessingRulesSlice(preprocessingRules []interface{}) *[]platformclientv2.Preprocessingrule {
	if preprocessingRules == nil || len(preprocessingRules) < 1 {
		return nil
	}
	sdkPreprocessingRulesSlice := make([]platformclientv2.Preprocessingrule, 0)
	for _, preprocessingRule := range preprocessingRules {
		if preprocessingRuleMap, ok := preprocessingRule.(map[string]interface{}); ok {
			var sdkPreprocessingRule platformclientv2.Preprocessingrule

			if findStr, ok := preprocessingRuleMap["find"].(string); ok {
				sdkPreprocessingRule.Find = &findStr
			}
			if replaceWithStr, ok := preprocessingRuleMap["replace_with"].(string); ok {
				sdkPreprocessingRule.ReplaceWith = &replaceWithStr
			}
			if isGlobal, ok := preprocessingRuleMap["global"].(bool); ok {
				sdkPreprocessingRule.Global = platformclientv2.Bool(isGlobal)
			}
			if isIgnoreCase, ok := preprocessingRuleMap["ignore_case"].(bool); ok {
				sdkPreprocessingRule.IgnoreCase = platformclientv2.Bool(isIgnoreCase)
			}
			sdkPreprocessingRulesSlice = append(sdkPreprocessingRulesSlice, sdkPreprocessingRule)
		}
	}
	return &sdkPreprocessingRulesSlice
}

func flattenSdkOutboundFileSpecificationTemplateColumnInformationSlice(fileSpecificationTemplateColumnInformation []platformclientv2.Column) []interface{} {
	if len(fileSpecificationTemplateColumnInformation) == 0 {
		return nil
	}
	columnInformationList := make([]interface{}, 0)
	for _, columnInformation := range fileSpecificationTemplateColumnInformation {
		columnInformationMap := make(map[string]interface{})
		if columnInformation.ColumnName != nil {
			columnInformationMap["column_name"] = *columnInformation.ColumnName
		}
		if columnInformation.ColumnNumber != nil {
			columnInformationMap["column_number"] = *columnInformation.ColumnNumber
		}
		if columnInformation.StartPosition != nil {
			columnInformationMap["start_position"] = *columnInformation.StartPosition
		}
		if columnInformation.Length != nil {
			columnInformationMap["length"] = *columnInformation.Length
		}
		columnInformationList = append(columnInformationList, columnInformationMap)
	}
	return columnInformationList
}

func flattenSdkOutboundFileSpecificationTemplatePreprocessingRulesSlice(fileSpecificationTemplatePreprocessingRules []platformclientv2.Preprocessingrule) []interface{} {
	if len(fileSpecificationTemplatePreprocessingRules) == 0 {
		return nil
	}
	preprocessingRulesList := make([]interface{}, 0)
	for _, preprocessingRule := range fileSpecificationTemplatePreprocessingRules {
		preprocessingRuleMap := make(map[string]interface{})
		if preprocessingRule.Find != nil {
			preprocessingRuleMap["find"] = *preprocessingRule.Find
		}
		if preprocessingRule.ReplaceWith != nil {
			preprocessingRuleMap["replace_with"] = *preprocessingRule.ReplaceWith
		}
		if preprocessingRule.Global != nil {
			preprocessingRuleMap["global"] = *preprocessingRule.Global
		}
		if preprocessingRule.IgnoreCase != nil {
			preprocessingRuleMap["ignore_case"] = *preprocessingRule.IgnoreCase
		}
		preprocessingRulesList = append(preprocessingRulesList, preprocessingRuleMap)
	}
	return preprocessingRulesList
}
