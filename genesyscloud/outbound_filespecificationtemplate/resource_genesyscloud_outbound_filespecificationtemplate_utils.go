package outbound_filespecificationtemplate

import (
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getFilespecificationtemplateFromResourceData(d *schema.ResourceData) platformclientv2.Filespecificationtemplate {
	description := d.Get("description").(string)
	delimiter := d.Get("delimiter").(string)

	sdkFileSpecificationTemplate := platformclientv2.Filespecificationtemplate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		Format:                       platformclientv2.String(d.Get("format").(string)),
		NumberOfHeadingLinesSkipped:  platformclientv2.Int(d.Get("number_of_header_lines_skipped").(int)),
		NumberOfTrailingLinesSkipped: platformclientv2.Int(d.Get("number_of_trailer_lines_skipped").(int)),
		Header:                       platformclientv2.Bool(d.Get("header").(bool)),
		DelimiterValue:               platformclientv2.String(d.Get("delimiter_value").(string)),
		ColumnInformation:            buildSdkOutboundFileSpecificationTemplateColumnInformationSlice(d.Get("column_information").([]interface{})),
		PreprocessingRules:           buildSdkOutboundFileSpecificationTemplatePreprocessingRulesSlice(d.Get("preprocessing_rule").([]interface{})),
	}

	if description != "" {
		sdkFileSpecificationTemplate.Description = &description
	}
	if delimiter != "" {
		sdkFileSpecificationTemplate.Delimiter = &delimiter
	}

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

			resourcedata.BuildSDKStringValueIfNotNil(&sdkColumnInformation.ColumnName, columnInfoMap, "column_name")

			if columnNumberInt, ok := columnInfoMap["column_number"].(int); ok {
				sdkColumnInformation.ColumnNumber = platformclientv2.Int(columnNumberInt)
			}
			if startPositionInt, ok := columnInfoMap["start_position"].(int); ok {
				sdkColumnInformation.StartPosition = platformclientv2.Int(startPositionInt)
			}
			if lengthInt, ok := columnInfoMap["length"].(int); ok {
				sdkColumnInformation.Length = platformclientv2.Int(lengthInt)
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

			resourcedata.BuildSDKStringValueIfNotNil(&sdkPreprocessingRule.Find, preprocessingRuleMap, "find")
			resourcedata.BuildSDKStringValueIfNotNil(&sdkPreprocessingRule.ReplaceWith, preprocessingRuleMap, "replace_with")

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

func flattenSdkOutboundFileSpecificationTemplateColumnInformationSlice(fileSpecificationTemplateColumnInformation *[]platformclientv2.Column) []interface{} {
	if len(*fileSpecificationTemplateColumnInformation) == 0 {
		return nil
	}
	columnInformationList := make([]interface{}, 0)
	for _, columnInformation := range *fileSpecificationTemplateColumnInformation {
		columnInformationMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(columnInformationMap, "column_name", columnInformation.ColumnName)
		resourcedata.SetMapValueIfNotNil(columnInformationMap, "column_number", columnInformation.ColumnNumber)
		resourcedata.SetMapValueIfNotNil(columnInformationMap, "start_position", columnInformation.StartPosition)
		resourcedata.SetMapValueIfNotNil(columnInformationMap, "length", columnInformation.Length)

		columnInformationList = append(columnInformationList, columnInformationMap)
	}
	return columnInformationList
}

func flattenSdkOutboundFileSpecificationTemplatePreprocessingRulesSlice(fileSpecificationTemplatePreprocessingRules *[]platformclientv2.Preprocessingrule) []interface{} {
	if len(*fileSpecificationTemplatePreprocessingRules) == 0 {
		return nil
	}
	preprocessingRulesList := make([]interface{}, 0)
	for _, preprocessingRule := range *fileSpecificationTemplatePreprocessingRules {
		preprocessingRuleMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(preprocessingRuleMap, "find", preprocessingRule.Find)
		resourcedata.SetMapValueIfNotNil(preprocessingRuleMap, "replace_with", preprocessingRule.ReplaceWith)
		resourcedata.SetMapValueIfNotNil(preprocessingRuleMap, "global", preprocessingRule.Global)
		resourcedata.SetMapValueIfNotNil(preprocessingRuleMap, "ignore_case", preprocessingRule.IgnoreCase)

		preprocessingRulesList = append(preprocessingRulesList, preprocessingRuleMap)
	}
	return preprocessingRulesList
}
