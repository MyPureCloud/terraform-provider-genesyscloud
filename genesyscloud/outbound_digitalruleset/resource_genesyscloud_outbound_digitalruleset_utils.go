package outbound_digitalruleset

import (
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_outbound_digitalruleset_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getOutboundDigitalrulesetFromResourceData maps data from schema ResourceData object to a platformclientv2.Digitalruleset
func getOutboundDigitalrulesetFromResourceData(d *schema.ResourceData) platformclientv2.Digitalruleset {
	return platformclientv2.Digitalruleset{
		Name:        platformclientv2.String(d.Get("name").(string)),
		ContactList: util.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Rules:       buildDigitalRules(d.Get("rules").([]interface{})),
	}
}

// buildContactColumnConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Contactcolumnconditionsettings
func buildContactColumnConditionSettings(contactColumnConditionSettings *schema.Set) *platformclientv2.Contactcolumnconditionsettings {

	if contactColumnConditionSettings == nil {
		return nil
	}

	var sdkContactColumnConditionSettings platformclientv2.Contactcolumnconditionsettings
	sdkContactColumnConditionSettingsList := contactColumnConditionSettings.List()

	if len(sdkContactColumnConditionSettingsList) > 0 {
		contactColumnConditionSettingsMap := sdkContactColumnConditionSettingsList[0].(map[string]interface{})

		if columnName := contactColumnConditionSettingsMap["column_name"].(string); columnName != "" {
			sdkContactColumnConditionSettings.ColumnName = &columnName
		}

		if operator := contactColumnConditionSettingsMap["operator"].(string); operator != "" {
			sdkContactColumnConditionSettings.Operator = &operator
		}

		if value := contactColumnConditionSettingsMap["value"].(string); value != "" {
			sdkContactColumnConditionSettings.Value = &value
		}

		if valueType := contactColumnConditionSettingsMap["value_type"].(string); valueType != "" {
			sdkContactColumnConditionSettings.ValueType = &valueType
		}
	}

	return &sdkContactColumnConditionSettings
}

// buildContactAddressConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Contactaddressconditionsettings
func buildContactAddressConditionSettings(contactAddressConditionSettings *schema.Set) *platformclientv2.Contactaddressconditionsettings {
	if contactAddressConditionSettings == nil {
		return nil
	}

	var sdkContactAddressConditionSettings platformclientv2.Contactaddressconditionsettings
	contactAddressConditionSettingsList := contactAddressConditionSettings.List()

	if len(contactAddressConditionSettingsList) > 0 {
		contactAddressConditionSettingsMap := contactAddressConditionSettingsList[0].(map[string]interface{})

		if operator := contactAddressConditionSettingsMap["operator"].(string); operator != "" {
			sdkContactAddressConditionSettings.Operator = &operator
		}

		if value := contactAddressConditionSettingsMap["value"].(string); value != "" {
			sdkContactAddressConditionSettings.Value = &value
		}
	}

	return &sdkContactAddressConditionSettings
}

// buildContactAddressTypeConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Contactaddresstypeconditionsettings
func buildContactAddressTypeConditionSettings(contactAddressTypeConditionSettings *schema.Set) *platformclientv2.Contactaddresstypeconditionsettings {
	if contactAddressTypeConditionSettings == nil {
		return nil
	}

	var sdkContactAddressTypeConditionSettings platformclientv2.Contactaddresstypeconditionsettings
	contactAddressTypeConditionSettingsList := contactAddressTypeConditionSettings.List()

	if len(contactAddressTypeConditionSettingsList) > 0 {
		contactAddressTypeConditionSettingsMap := contactAddressTypeConditionSettingsList[0].(map[string]interface{})

		if operator := contactAddressTypeConditionSettingsMap["operator"].(string); operator != "" {
			sdkContactAddressTypeConditionSettings.Operator = &operator
		}

		if value := contactAddressTypeConditionSettingsMap["value"].(string); value != "" {
			sdkContactAddressTypeConditionSettings.Value = &value
		}
	}

	return &sdkContactAddressTypeConditionSettings
}

// buildLastAttemptByColumnConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Lastattemptbycolumnconditionsettings
func buildLastAttemptByColumnConditionSettings(lastAttemptByColumnConditionSettings *schema.Set) *platformclientv2.Lastattemptbycolumnconditionsettings {
	if lastAttemptByColumnConditionSettings == nil {
		return nil
	}

	var sdkLastAttemptByColumnConditionSettings platformclientv2.Lastattemptbycolumnconditionsettings
	lastAttemptByColumnConditionSettingsList := lastAttemptByColumnConditionSettings.List()

	if len(lastAttemptByColumnConditionSettingsList) > 0 {
		lastAttemptByColumnConditionSettingsMap := lastAttemptByColumnConditionSettingsList[0].(map[string]interface{})

		if emailColumnName := lastAttemptByColumnConditionSettingsMap["email_column_name"].(string); emailColumnName != "" {
			sdkLastAttemptByColumnConditionSettings.EmailColumnName = &emailColumnName
		}

		if smsColumnName := lastAttemptByColumnConditionSettingsMap["sms_column_name"].(string); smsColumnName != "" {
			sdkLastAttemptByColumnConditionSettings.SmsColumnName = &smsColumnName
		}

		if operator := lastAttemptByColumnConditionSettingsMap["operator"].(string); operator != "" {
			sdkLastAttemptByColumnConditionSettings.Operator = &operator
		}

		if value := lastAttemptByColumnConditionSettingsMap["value"].(string); value != "" {
			sdkLastAttemptByColumnConditionSettings.Value = &value
		}
	}

	return &sdkLastAttemptByColumnConditionSettings
}

// buildLastAttemptOverallConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Lastattemptoverallconditionsettings
func buildLastAttemptOverallConditionSettings(lastAttemptOverallConditionSettings *schema.Set) *platformclientv2.Lastattemptoverallconditionsettings {
	if lastAttemptOverallConditionSettings == nil {
		return nil
	}

	var sdkLastAttemptOverallConditionSettings platformclientv2.Lastattemptoverallconditionsettings
	lastAttemptOverallConditionSettingsList := lastAttemptOverallConditionSettings.List()

	if len(lastAttemptOverallConditionSettingsList) > 0 {
		lastAttemptOverallConditionSettingsMap := lastAttemptOverallConditionSettingsList[0].(map[string]interface{})

		if mediaTypes := lastAttemptOverallConditionSettingsMap["media_types"].([]string); len(mediaTypes) > 0 {
			sdkLastAttemptOverallConditionSettings.MediaTypes = &mediaTypes
		}

		if operator := lastAttemptOverallConditionSettingsMap["operator"].(string); operator != "" {
			sdkLastAttemptOverallConditionSettings.Operator = &operator
		}

		if value := lastAttemptOverallConditionSettingsMap["value"].(string); value != "" {
			sdkLastAttemptOverallConditionSettings.Value = &value
		}
	}

	return &sdkLastAttemptOverallConditionSettings
}

// buildLastResultByColumnConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Lastresultbycolumnconditionsettings
func buildLastResultByColumnConditionSettings(lastResultByColumnConditionSettings *schema.Set) *platformclientv2.Lastresultbycolumnconditionsettings {
	if lastResultByColumnConditionSettings == nil {
		return nil
	}

	var sdkLastResultByColumnConditionSettings platformclientv2.Lastresultbycolumnconditionsettings
	lastResultByColumnConditionSettingsList := lastResultByColumnConditionSettings.List()

	if len(lastResultByColumnConditionSettingsList) > 0 {
		lastResultByColumnConditionSettingsMap := lastResultByColumnConditionSettingsList[0].(map[string]interface{})

		if emailColumnName := lastResultByColumnConditionSettingsMap["email_column_name"].(string); emailColumnName != "" {
			sdkLastResultByColumnConditionSettings.EmailColumnName = &emailColumnName
		}

		if smsColumnName := lastResultByColumnConditionSettingsMap["sms_column_name"].(string); smsColumnName != "" {
			sdkLastResultByColumnConditionSettings.SmsColumnName = &smsColumnName
		}

		if emailWrapupCodes := lastResultByColumnConditionSettingsMap["email_wrapup_codes"].([]string); len(emailWrapupCodes) > 0 {
			sdkLastResultByColumnConditionSettings.EmailWrapupCodes = &emailWrapupCodes
		}

		if smsWrapupCodes := lastResultByColumnConditionSettingsMap["sms_wrapup_codes"].([]string); len(smsWrapupCodes) > 0 {
			sdkLastResultByColumnConditionSettings.SmsWrapupCodes = &smsWrapupCodes
		}
	}

	return &sdkLastResultByColumnConditionSettings
}

// buildLastResultOverallConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Lastresultoverallconditionsettings
func buildLastResultOverallConditionSettings(lastResultOverallConditionSettings *schema.Set) *platformclientv2.Lastresultoverallconditionsettings {
	if lastResultOverallConditionSettings == nil {
		return nil
	}

	var sdkLastResultOverallConditionSettings platformclientv2.Lastresultoverallconditionsettings
	lastResultOverallConditionSettingsList := lastResultOverallConditionSettings.List()

	if len(lastResultOverallConditionSettingsList) > 0 {
		lastResultOverallConditionSettingsMap := lastResultOverallConditionSettingsList[0].(map[string]interface{})

		if emailWrapupCodes := lastResultOverallConditionSettingsMap["email_wrapup_codes"].([]string); len(emailWrapupCodes) > 0 {
			sdkLastResultOverallConditionSettings.EmailWrapupCodes = &emailWrapupCodes
		}

		if smsWrapupCodes := lastResultOverallConditionSettingsMap["sms_wrapup_codes"].([]string); len(smsWrapupCodes) > 0 {
			sdkLastResultOverallConditionSettings.SmsWrapupCodes = &smsWrapupCodes
		}
	}

	return &sdkLastResultOverallConditionSettings
}

// buildDigitalDataActionConditionPredicates maps an []interface{} into a Genesys Cloud *[]platformclientv2.Digitaldataactionconditionpredicate
func buildDigitalDataActionConditionPredicates(digitalDataActionConditionPredicates []interface{}) *[]platformclientv2.Digitaldataactionconditionpredicate {
	digitalDataActionConditionPredicatesSlice := make([]platformclientv2.Digitaldataactionconditionpredicate, 0)
	for _, digitalDataActionConditionPredicate := range digitalDataActionConditionPredicates {
		var sdkDigitalDataActionConditionPredicate platformclientv2.Digitaldataactionconditionpredicate
		digitalDataActionConditionPredicatesMap, ok := digitalDataActionConditionPredicate.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDigitalDataActionConditionPredicate.OutputField, digitalDataActionConditionPredicatesMap, "output_field")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDigitalDataActionConditionPredicate.OutputOperator, digitalDataActionConditionPredicatesMap, "output_operator")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDigitalDataActionConditionPredicate.ComparisonValue, digitalDataActionConditionPredicatesMap, "comparison_value")
		sdkDigitalDataActionConditionPredicate.Inverted = platformclientv2.Bool(digitalDataActionConditionPredicatesMap["inverted"].(bool))
		sdkDigitalDataActionConditionPredicate.OutputFieldMissingResolution = platformclientv2.Bool(digitalDataActionConditionPredicatesMap["output_field_missing_resolution"].(bool))

		digitalDataActionConditionPredicatesSlice = append(digitalDataActionConditionPredicatesSlice, sdkDigitalDataActionConditionPredicate)
	}

	return &digitalDataActionConditionPredicatesSlice
}

// buildDataActionContactColumnFieldMappings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Dataactioncontactcolumnfieldmapping
func buildDataActionContactColumnFieldMappings(dataActionContactColumnFieldMappings []interface{}) *[]platformclientv2.Dataactioncontactcolumnfieldmapping {
	dataActionContactColumnFieldMappingsSlice := make([]platformclientv2.Dataactioncontactcolumnfieldmapping, 0)
	for _, dataActionContactColumnFieldMapping := range dataActionContactColumnFieldMappings {
		var sdkDataActionContactColumnFieldMapping platformclientv2.Dataactioncontactcolumnfieldmapping
		dataActionContactColumnFieldMappingsMap, ok := dataActionContactColumnFieldMapping.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDataActionContactColumnFieldMapping.ContactColumnName, dataActionContactColumnFieldMappingsMap, "contact_column_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDataActionContactColumnFieldMapping.DataActionField, dataActionContactColumnFieldMappingsMap, "data_action_field")

		dataActionContactColumnFieldMappingsSlice = append(dataActionContactColumnFieldMappingsSlice, sdkDataActionContactColumnFieldMapping)
	}

	return &dataActionContactColumnFieldMappingsSlice
}

// buildDataActionConditionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Dataactionconditionsettings
func buildDataActionConditionSettings(dataActionConditionSettings *schema.Set) *platformclientv2.Dataactionconditionsettings {
	if dataActionConditionSettings == nil {
		return nil
	}

	var sdkDataActionConditionSettings platformclientv2.Dataactionconditionsettings
	dataActionConditionSettingsList := dataActionConditionSettings.List()

	if len(dataActionConditionSettingsList) > 0 {
		dataActionConditionSettingsMap := dataActionConditionSettingsList[0].(map[string]interface{})

		if dataActionId := dataActionConditionSettingsMap["data_action_id"].(string); dataActionId != "" {
			sdkDataActionConditionSettings.DataActionId = &dataActionId
		}

		if contactIdField := dataActionConditionSettingsMap["contact_id_field"].(string); contactIdField != "" {
			sdkDataActionConditionSettings.ContactIdField = &contactIdField
		}

		dataNotFoundResolution := dataActionConditionSettingsMap["data_not_found_resolution"].(bool)
		sdkDataActionConditionSettings.DataNotFoundResolution = &dataNotFoundResolution

		sdkDataActionConditionSettings.Predicates = buildDigitalDataActionConditionPredicates(dataActionConditionSettingsMap["predicates"].([]interface{}))
		sdkDataActionConditionSettings.ContactColumnToDataActionFieldMappings = buildDataActionContactColumnFieldMappings(dataActionConditionSettingsMap["contact_column_to_data_action_field_mappings"].([]interface{}))

	}

	return &sdkDataActionConditionSettings
}

// buildDigitalConditions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Digitalcondition
func buildDigitalConditions(digitalConditions []interface{}) *[]platformclientv2.Digitalcondition {
	digitalConditionsSlice := make([]platformclientv2.Digitalcondition, 0)
	for _, digitalCondition := range digitalConditions {
		var sdkDigitalCondition platformclientv2.Digitalcondition
		digitalConditionsMap, ok := digitalCondition.(map[string]interface{})
		if !ok {
			continue
		}

		sdkDigitalCondition.Inverted = platformclientv2.Bool(digitalConditionsMap["inverted"].(bool))
		sdkDigitalCondition.ContactColumnConditionSettings = buildContactColumnConditionSettings(digitalConditionsMap["contact_column_condition_settings"].(*schema.Set))
		sdkDigitalCondition.ContactAddressConditionSettings = buildContactAddressConditionSettings(digitalConditionsMap["contact_address_condition_settings"].(*schema.Set))
		sdkDigitalCondition.ContactAddressTypeConditionSettings = buildContactAddressTypeConditionSettings(digitalConditionsMap["contact_address_type_condition_settings"].(*schema.Set))
		sdkDigitalCondition.LastAttemptByColumnConditionSettings = buildLastAttemptByColumnConditionSettings(digitalConditionsMap["last_attempt_by_column_condition_settings"].(*schema.Set))
		sdkDigitalCondition.LastAttemptOverallConditionSettings = buildLastAttemptOverallConditionSettings(digitalConditionsMap["last_attempt_overall_condition_settings"].(*schema.Set))
		sdkDigitalCondition.LastResultByColumnConditionSettings = buildLastResultByColumnConditionSettings(digitalConditionsMap["last_result_by_column_condition_settings"].(*schema.Set))
		sdkDigitalCondition.LastResultOverallConditionSettings = buildLastResultOverallConditionSettings(digitalConditionsMap["last_result_overall_condition_settings"].(*schema.Set))
		sdkDigitalCondition.DataActionConditionSettings = buildDataActionConditionSettings(digitalConditionsMap["data_action_condition_settings"].(*schema.Set))

		digitalConditionsSlice = append(digitalConditionsSlice, sdkDigitalCondition)
	}

	return &digitalConditionsSlice
}

// buildUpdateContactColumnActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Updatecontactcolumnactionsettings
func buildUpdateContactColumnActionSettings(updateContactColumnActionSettings *schema.Set) *platformclientv2.Updatecontactcolumnactionsettings {
	if updateContactColumnActionSettings == nil {
		return nil
	}

	var sdkUpdateContactColumnActionSettings platformclientv2.Updatecontactcolumnactionsettings
	updateContactColumnActionSettingsList := updateContactColumnActionSettings.List()

	if len(updateContactColumnActionSettingsList) > 0 {
		updateContactColumnActionSettingsMap := updateContactColumnActionSettingsList[0].(map[string]interface{})

		if updateOption := updateContactColumnActionSettingsMap["update_option"].(string); updateOption != "" {
			sdkUpdateContactColumnActionSettings.UpdateOption = &updateOption
		}

		//resourcedata.BuildSDKStringMapValueIfNotNil(&sdkUpdateContactColumnActionSettings.Properties, updateContactColumnActionSettingsMap, "properties")
	}

	return &sdkUpdateContactColumnActionSettings
}

// buildAppendToDncActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Appendtodncactionsettings
func buildAppendToDncActionSettings(appendToDncActionSettingss []interface{}) *[]platformclientv2.Appendtodncactionsettings {
	appendToDncActionSettingssSlice := make([]platformclientv2.Appendtodncactionsettings, 0)
	for _, appendToDncActionSettings := range appendToDncActionSettingss {
		var sdkAppendToDncActionSettings platformclientv2.Appendtodncactionsettings
		appendToDncActionSettingssMap, ok := appendToDncActionSettings.(map[string]interface{})
		if !ok {
			continue
		}

		sdkAppendToDncActionSettings.Expire = platformclientv2.Bool(appendToDncActionSettingssMap["expire"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAppendToDncActionSettings.ExpirationDuration, appendToDncActionSettingssMap, "expiration_duration")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAppendToDncActionSettings.ListType, appendToDncActionSettingssMap, "list_type")

		appendToDncActionSettingssSlice = append(appendToDncActionSettingssSlice, sdkAppendToDncActionSettings)
	}

	return &appendToDncActionSettingssSlice
}

// buildMarkContactUncontactableActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Markcontactuncontactableactionsettings
func buildMarkContactUncontactableActionSettings(markContactUncontactableActionSettingss []interface{}) *[]platformclientv2.Markcontactuncontactableactionsettings {
	markContactUncontactableActionSettingssSlice := make([]platformclientv2.Markcontactuncontactableactionsettings, 0)
	for _, markContactUncontactableActionSettings := range markContactUncontactableActionSettingss {
		var sdkMarkContactUncontactableActionSettings platformclientv2.Markcontactuncontactableactionsettings
		markContactUncontactableActionSettingssMap, ok := markContactUncontactableActionSettings.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkMarkContactUncontactableActionSettings.MediaTypes, markContactUncontactableActionSettingssMap, "media_types")

		markContactUncontactableActionSettingssSlice = append(markContactUncontactableActionSettingssSlice, sdkMarkContactUncontactableActionSettings)
	}

	return &markContactUncontactableActionSettingssSlice
}

// buildSetContentTemplateActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Setcontenttemplateactionsettings
func buildSetContentTemplateActionSettings(setContentTemplateActionSettingss []interface{}) *[]platformclientv2.Setcontenttemplateactionsettings {
	setContentTemplateActionSettingssSlice := make([]platformclientv2.Setcontenttemplateactionsettings, 0)
	for _, setContentTemplateActionSettings := range setContentTemplateActionSettingss {
		var sdkSetContentTemplateActionSettings platformclientv2.Setcontenttemplateactionsettings
		setContentTemplateActionSettingssMap, ok := setContentTemplateActionSettings.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSetContentTemplateActionSettings.SmsContentTemplateId, setContentTemplateActionSettingssMap, "sms_content_template_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSetContentTemplateActionSettings.EmailContentTemplateId, setContentTemplateActionSettingssMap, "email_content_template_id")

		setContentTemplateActionSettingssSlice = append(setContentTemplateActionSettingssSlice, sdkSetContentTemplateActionSettings)
	}

	return &setContentTemplateActionSettingssSlice
}

// buildSetSmsPhoneNumberActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Setsmsphonenumberactionsettings
func buildSetSmsPhoneNumberActionSettings(setSmsPhoneNumberActionSettingss []interface{}) *[]platformclientv2.Setsmsphonenumberactionsettings {
	setSmsPhoneNumberActionSettingssSlice := make([]platformclientv2.Setsmsphonenumberactionsettings, 0)
	for _, setSmsPhoneNumberActionSettings := range setSmsPhoneNumberActionSettingss {
		var sdkSetSmsPhoneNumberActionSettings platformclientv2.Setsmsphonenumberactionsettings
		setSmsPhoneNumberActionSettingssMap, ok := setSmsPhoneNumberActionSettings.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSetSmsPhoneNumberActionSettings.SenderSmsPhoneNumber, setSmsPhoneNumberActionSettingssMap, "sender_sms_phone_number")

		setSmsPhoneNumberActionSettingssSlice = append(setSmsPhoneNumberActionSettingssSlice, sdkSetSmsPhoneNumberActionSettings)
	}

	return &setSmsPhoneNumberActionSettingssSlice
}

// buildDigitalActions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Digitalaction
func buildDigitalActions(digitalActions []interface{}) *[]platformclientv2.Digitalaction {
	digitalActionsSlice := make([]platformclientv2.Digitalaction, 0)
	for _, digitalAction := range digitalActions {
		var sdkDigitalAction platformclientv2.Digitalaction
		digitalActionsMap, ok := digitalAction.(map[string]interface{})
		if !ok {
			continue
		}

		sdkDigitalAction.UpdateContactColumnActionSettings = buildUpdateContactColumnActionSettings(digitalActionsMap["update_contact_column_action_settings"].(*schema.Set))
		// if action := digitalActionsMap["do_not_send_action_settings"]; action != nil {
		// 	sdkDigitalAction.DoNotSendActionSettings = &action
		// }

		// TODO: Handle do_not_send_action_settings property
		// resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDigitalAction.AppendToDncActionSettings, digitalActionsMap, "append_to_dnc_action_settings", buildAppendToDncActionSettings)
		// resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDigitalAction.MarkContactUncontactableActionSettings, digitalActionsMap, "mark_contact_uncontactable_action_settings", buildMarkContactUncontactableActionSettings)
		// // TODO: Handle mark_contact_address_uncontactable_action_settings property
		// resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDigitalAction.SetContentTemplateActionSettings, digitalActionsMap, "set_content_template_action_settings", buildSetContentTemplateActionSettings)
		// resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDigitalAction.SetSmsPhoneNumberActionSettings, digitalActionsMap, "set_sms_phone_number_action_settings", buildSetSmsPhoneNumberActionSettings)

		digitalActionsSlice = append(digitalActionsSlice, sdkDigitalAction)
	}

	return &digitalActionsSlice
}

// buildDigitalRules maps an []interface{} into a Genesys Cloud *[]platformclientv2.Digitalrule
func buildDigitalRules(digitalRules []interface{}) *[]platformclientv2.Digitalrule {
	digitalRulesSlice := make([]platformclientv2.Digitalrule, 0)
	for _, digitalRule := range digitalRules {
		var sdkDigitalRule platformclientv2.Digitalrule
		digitalRulesMap, ok := digitalRule.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDigitalRule.Name, digitalRulesMap, "name")
		sdkDigitalRule.Order = platformclientv2.Int(digitalRulesMap["order"].(int))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDigitalRule.Category, digitalRulesMap, "category")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDigitalRule.Conditions, digitalRulesMap, "conditions", buildDigitalConditions)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDigitalRule.Actions, digitalRulesMap, "actions", buildDigitalActions)

		digitalRulesSlice = append(digitalRulesSlice, sdkDigitalRule)
	}

	return &digitalRulesSlice
}

// flattenContactColumnConditionSettingss maps a Genesys Cloud *[]platformclientv2.Contactcolumnconditionsettings into a []interface{}
func flattenContactColumnConditionSettings(contactColumnConditionSettingss *[]platformclientv2.Contactcolumnconditionsettings) []interface{} {
	if len(*contactColumnConditionSettingss) == 0 {
		return nil
	}

	var contactColumnConditionSettingsList []interface{}
	for _, contactColumnConditionSettings := range *contactColumnConditionSettingss {
		contactColumnConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "column_name", contactColumnConditionSettings.ColumnName)
		resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "operator", contactColumnConditionSettings.Operator)
		resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "value", contactColumnConditionSettings.Value)
		resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "value_type", contactColumnConditionSettings.ValueType)

		contactColumnConditionSettingsList = append(contactColumnConditionSettingsList, contactColumnConditionSettingsMap)
	}

	return contactColumnConditionSettingsList
}

// flattenContactAddressConditionSettingss maps a Genesys Cloud *[]platformclientv2.Contactaddressconditionsettings into a []interface{}
func flattenContactAddressConditionSettings(contactAddressConditionSettingss *[]platformclientv2.Contactaddressconditionsettings) []interface{} {
	if len(*contactAddressConditionSettingss) == 0 {
		return nil
	}

	var contactAddressConditionSettingsList []interface{}
	for _, contactAddressConditionSettings := range *contactAddressConditionSettingss {
		contactAddressConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactAddressConditionSettingsMap, "operator", contactAddressConditionSettings.Operator)
		resourcedata.SetMapValueIfNotNil(contactAddressConditionSettingsMap, "value", contactAddressConditionSettings.Value)

		contactAddressConditionSettingsList = append(contactAddressConditionSettingsList, contactAddressConditionSettingsMap)
	}

	return contactAddressConditionSettingsList
}

// flattenContactAddressTypeConditionSettingss maps a Genesys Cloud *[]platformclientv2.Contactaddresstypeconditionsettings into a []interface{}
func flattenContactAddressTypeConditionSettings(contactAddressTypeConditionSettingss *[]platformclientv2.Contactaddresstypeconditionsettings) []interface{} {
	if len(*contactAddressTypeConditionSettingss) == 0 {
		return nil
	}

	var contactAddressTypeConditionSettingsList []interface{}
	for _, contactAddressTypeConditionSettings := range *contactAddressTypeConditionSettingss {
		contactAddressTypeConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactAddressTypeConditionSettingsMap, "operator", contactAddressTypeConditionSettings.Operator)
		resourcedata.SetMapValueIfNotNil(contactAddressTypeConditionSettingsMap, "value", contactAddressTypeConditionSettings.Value)

		contactAddressTypeConditionSettingsList = append(contactAddressTypeConditionSettingsList, contactAddressTypeConditionSettingsMap)
	}

	return contactAddressTypeConditionSettingsList
}

// flattenLastAttemptByColumnConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastattemptbycolumnconditionsettings into a []interface{}
func flattenLastAttemptByColumnConditionSettings(lastAttemptByColumnConditionSettingss *[]platformclientv2.Lastattemptbycolumnconditionsettings) []interface{} {
	if len(*lastAttemptByColumnConditionSettingss) == 0 {
		return nil
	}

	var lastAttemptByColumnConditionSettingsList []interface{}
	for _, lastAttemptByColumnConditionSettings := range *lastAttemptByColumnConditionSettingss {
		lastAttemptByColumnConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "email_column_name", lastAttemptByColumnConditionSettings.EmailColumnName)
		resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "sms_column_name", lastAttemptByColumnConditionSettings.SmsColumnName)
		resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "operator", lastAttemptByColumnConditionSettings.Operator)
		resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "value", lastAttemptByColumnConditionSettings.Value)

		lastAttemptByColumnConditionSettingsList = append(lastAttemptByColumnConditionSettingsList, lastAttemptByColumnConditionSettingsMap)
	}

	return lastAttemptByColumnConditionSettingsList
}

// flattenLastAttemptOverallConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastattemptoverallconditionsettings into a []interface{}
func flattenLastAttemptOverallConditionSettings(lastAttemptOverallConditionSettingss *[]platformclientv2.Lastattemptoverallconditionsettings) []interface{} {
	if len(*lastAttemptOverallConditionSettingss) == 0 {
		return nil
	}

	var lastAttemptOverallConditionSettingsList []interface{}
	for _, lastAttemptOverallConditionSettings := range *lastAttemptOverallConditionSettingss {
		lastAttemptOverallConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapStringArrayValueIfNotNil(lastAttemptOverallConditionSettingsMap, "media_types", lastAttemptOverallConditionSettings.MediaTypes)
		resourcedata.SetMapValueIfNotNil(lastAttemptOverallConditionSettingsMap, "operator", lastAttemptOverallConditionSettings.Operator)
		resourcedata.SetMapValueIfNotNil(lastAttemptOverallConditionSettingsMap, "value", lastAttemptOverallConditionSettings.Value)

		lastAttemptOverallConditionSettingsList = append(lastAttemptOverallConditionSettingsList, lastAttemptOverallConditionSettingsMap)
	}

	return lastAttemptOverallConditionSettingsList
}

// flattenLastResultByColumnConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastresultbycolumnconditionsettings into a []interface{}
func flattenLastResultByColumnConditionSettings(lastResultByColumnConditionSettingss *[]platformclientv2.Lastresultbycolumnconditionsettings) []interface{} {
	if len(*lastResultByColumnConditionSettingss) == 0 {
		return nil
	}

	var lastResultByColumnConditionSettingsList []interface{}
	for _, lastResultByColumnConditionSettings := range *lastResultByColumnConditionSettingss {
		lastResultByColumnConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(lastResultByColumnConditionSettingsMap, "email_column_name", lastResultByColumnConditionSettings.EmailColumnName)
		resourcedata.SetMapStringArrayValueIfNotNil(lastResultByColumnConditionSettingsMap, "email_wrapup_codes", lastResultByColumnConditionSettings.EmailWrapupCodes)
		resourcedata.SetMapValueIfNotNil(lastResultByColumnConditionSettingsMap, "sms_column_name", lastResultByColumnConditionSettings.SmsColumnName)
		resourcedata.SetMapStringArrayValueIfNotNil(lastResultByColumnConditionSettingsMap, "sms_wrapup_codes", lastResultByColumnConditionSettings.SmsWrapupCodes)

		lastResultByColumnConditionSettingsList = append(lastResultByColumnConditionSettingsList, lastResultByColumnConditionSettingsMap)
	}

	return lastResultByColumnConditionSettingsList
}

// flattenLastResultOverallConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastresultoverallconditionsettings into a []interface{}
func flattenLastResultOverallConditionSettings(lastResultOverallConditionSettingss *[]platformclientv2.Lastresultoverallconditionsettings) []interface{} {
	if len(*lastResultOverallConditionSettingss) == 0 {
		return nil
	}

	var lastResultOverallConditionSettingsList []interface{}
	for _, lastResultOverallConditionSettings := range *lastResultOverallConditionSettingss {
		lastResultOverallConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapStringArrayValueIfNotNil(lastResultOverallConditionSettingsMap, "email_wrapup_codes", lastResultOverallConditionSettings.EmailWrapupCodes)
		resourcedata.SetMapStringArrayValueIfNotNil(lastResultOverallConditionSettingsMap, "sms_wrapup_codes", lastResultOverallConditionSettings.SmsWrapupCodes)

		lastResultOverallConditionSettingsList = append(lastResultOverallConditionSettingsList, lastResultOverallConditionSettingsMap)
	}

	return lastResultOverallConditionSettingsList
}

// flattenDigitalDataActionConditionPredicates maps a Genesys Cloud *[]platformclientv2.Digitaldataactionconditionpredicate into a []interface{}
func flattenDigitalDataActionConditionPredicates(digitalDataActionConditionPredicates *[]platformclientv2.Digitaldataactionconditionpredicate) []interface{} {
	if len(*digitalDataActionConditionPredicates) == 0 {
		return nil
	}

	var digitalDataActionConditionPredicateList []interface{}
	for _, digitalDataActionConditionPredicate := range *digitalDataActionConditionPredicates {
		digitalDataActionConditionPredicateMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(digitalDataActionConditionPredicateMap, "output_field", digitalDataActionConditionPredicate.OutputField)
		resourcedata.SetMapValueIfNotNil(digitalDataActionConditionPredicateMap, "output_operator", digitalDataActionConditionPredicate.OutputOperator)
		resourcedata.SetMapValueIfNotNil(digitalDataActionConditionPredicateMap, "comparison_value", digitalDataActionConditionPredicate.ComparisonValue)
		resourcedata.SetMapValueIfNotNil(digitalDataActionConditionPredicateMap, "inverted", digitalDataActionConditionPredicate.Inverted)
		resourcedata.SetMapValueIfNotNil(digitalDataActionConditionPredicateMap, "output_field_missing_resolution", digitalDataActionConditionPredicate.OutputFieldMissingResolution)

		digitalDataActionConditionPredicateList = append(digitalDataActionConditionPredicateList, digitalDataActionConditionPredicateMap)
	}

	return digitalDataActionConditionPredicateList
}

// flattenDataActionContactColumnFieldMappings maps a Genesys Cloud *[]platformclientv2.Dataactioncontactcolumnfieldmapping into a []interface{}
func flattenDataActionContactColumnFieldMappings(dataActionContactColumnFieldMappings *[]platformclientv2.Dataactioncontactcolumnfieldmapping) []interface{} {
	if len(*dataActionContactColumnFieldMappings) == 0 {
		return nil
	}

	var dataActionContactColumnFieldMappingList []interface{}
	for _, dataActionContactColumnFieldMapping := range *dataActionContactColumnFieldMappings {
		dataActionContactColumnFieldMappingMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(dataActionContactColumnFieldMappingMap, "contact_column_name", dataActionContactColumnFieldMapping.ContactColumnName)
		resourcedata.SetMapValueIfNotNil(dataActionContactColumnFieldMappingMap, "data_action_field", dataActionContactColumnFieldMapping.DataActionField)

		dataActionContactColumnFieldMappingList = append(dataActionContactColumnFieldMappingList, dataActionContactColumnFieldMappingMap)
	}

	return dataActionContactColumnFieldMappingList
}

// flattenDataActionConditionSettingss maps a Genesys Cloud *[]platformclientv2.Dataactionconditionsettings into a []interface{}
func flattenDataActionConditionSettings(dataActionConditionSettingss *[]platformclientv2.Dataactionconditionsettings) []interface{} {
	if len(*dataActionConditionSettingss) == 0 {
		return nil
	}

	var dataActionConditionSettingsList []interface{}
	for _, dataActionConditionSettings := range *dataActionConditionSettingss {
		dataActionConditionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(dataActionConditionSettingsMap, "data_action_id", dataActionConditionSettings.DataActionId)
		resourcedata.SetMapValueIfNotNil(dataActionConditionSettingsMap, "contact_id_field", dataActionConditionSettings.ContactIdField)
		resourcedata.SetMapValueIfNotNil(dataActionConditionSettingsMap, "data_not_found_resolution", dataActionConditionSettings.DataNotFoundResolution)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(dataActionConditionSettingsMap, "predicates", dataActionConditionSettings.Predicates, flattenDigitalDataActionConditionPredicates)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(dataActionConditionSettingsMap, "contact_column_to_data_action_field_mappings", dataActionConditionSettings.ContactColumnToDataActionFieldMappings, flattenDataActionContactColumnFieldMappings)

		dataActionConditionSettingsList = append(dataActionConditionSettingsList, dataActionConditionSettingsMap)
	}

	return dataActionConditionSettingsList
}

// flattenDigitalConditions maps a Genesys Cloud *[]platformclientv2.Digitalcondition into a []interface{}
func flattenDigitalConditions(digitalConditions *[]platformclientv2.Digitalcondition) []interface{} {
	if len(*digitalConditions) == 0 {
		return nil
	}

	var digitalConditionList []interface{}
	for _, digitalCondition := range *digitalConditions {
		digitalConditionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(digitalConditionMap, "inverted", digitalCondition.Inverted)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "contact_column_condition_settings", digitalCondition.ContactColumnConditionSettings, flattenContactColumnConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "contact_address_condition_settings", digitalCondition.ContactAddressConditionSettings, flattenContactAddressConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "contact_address_type_condition_settings", digitalCondition.ContactAddressTypeConditionSettings, flattenContactAddressTypeConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "last_attempt_by_column_condition_settings", digitalCondition.LastAttemptByColumnConditionSettings, flattenLastAttemptByColumnConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "last_attempt_overall_condition_settings", digitalCondition.LastAttemptOverallConditionSettings, flattenLastAttemptOverallConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "last_result_by_column_condition_settings", digitalCondition.LastResultByColumnConditionSettings, flattenLastResultByColumnConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "last_result_overall_condition_settings", digitalCondition.LastResultOverallConditionSettings, flattenLastResultOverallConditionSettings)
		// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalConditionMap, "data_action_condition_settings", digitalCondition.DataActionConditionSettings, flattenDataActionConditionSettings)

		digitalConditionList = append(digitalConditionList, digitalConditionMap)
	}

	return digitalConditionList
}

// flattenUpdateContactColumnActionSettingss maps a Genesys Cloud *[]platformclientv2.Updatecontactcolumnactionsettings into a []interface{}
func flattenUpdateContactColumnActionSettings(updateContactColumnActionSettingss *[]platformclientv2.Updatecontactcolumnactionsettings) []interface{} {
	if len(*updateContactColumnActionSettingss) == 0 {
		return nil
	}

	var updateContactColumnActionSettingsList []interface{}
	for _, updateContactColumnActionSettings := range *updateContactColumnActionSettingss {
		updateContactColumnActionSettingsMap := make(map[string]interface{})

		// TODO: Handle properties property
		resourcedata.SetMapValueIfNotNil(updateContactColumnActionSettingsMap, "update_option", updateContactColumnActionSettings.UpdateOption)

		updateContactColumnActionSettingsList = append(updateContactColumnActionSettingsList, updateContactColumnActionSettingsMap)
	}

	return updateContactColumnActionSettingsList
}

// flattenAppendToDncActionSettingss maps a Genesys Cloud *[]platformclientv2.Appendtodncactionsettings into a []interface{}
func flattenAppendToDncActionSettings(appendToDncActionSettingss *[]platformclientv2.Appendtodncactionsettings) []interface{} {
	if len(*appendToDncActionSettingss) == 0 {
		return nil
	}

	var appendToDncActionSettingsList []interface{}
	for _, appendToDncActionSettings := range *appendToDncActionSettingss {
		appendToDncActionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(appendToDncActionSettingsMap, "expire", appendToDncActionSettings.Expire)
		resourcedata.SetMapValueIfNotNil(appendToDncActionSettingsMap, "expiration_duration", appendToDncActionSettings.ExpirationDuration)
		resourcedata.SetMapValueIfNotNil(appendToDncActionSettingsMap, "list_type", appendToDncActionSettings.ListType)

		appendToDncActionSettingsList = append(appendToDncActionSettingsList, appendToDncActionSettingsMap)
	}

	return appendToDncActionSettingsList
}

// flattenMarkContactUncontactableActionSettingss maps a Genesys Cloud *[]platformclientv2.Markcontactuncontactableactionsettings into a []interface{}
func flattenMarkContactUncontactableActionSettings(markContactUncontactableActionSettingss *[]platformclientv2.Markcontactuncontactableactionsettings) []interface{} {
	if len(*markContactUncontactableActionSettingss) == 0 {
		return nil
	}

	var markContactUncontactableActionSettingsList []interface{}
	for _, markContactUncontactableActionSettings := range *markContactUncontactableActionSettingss {
		markContactUncontactableActionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapStringArrayValueIfNotNil(markContactUncontactableActionSettingsMap, "media_types", markContactUncontactableActionSettings.MediaTypes)

		markContactUncontactableActionSettingsList = append(markContactUncontactableActionSettingsList, markContactUncontactableActionSettingsMap)
	}

	return markContactUncontactableActionSettingsList
}

// flattenSetContentTemplateActionSettingss maps a Genesys Cloud *[]platformclientv2.Setcontenttemplateactionsettings into a []interface{}
func flattenSetContentTemplateActionSettings(setContentTemplateActionSettingss *[]platformclientv2.Setcontenttemplateactionsettings) []interface{} {
	if len(*setContentTemplateActionSettingss) == 0 {
		return nil
	}

	var setContentTemplateActionSettingsList []interface{}
	for _, setContentTemplateActionSettings := range *setContentTemplateActionSettingss {
		setContentTemplateActionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(setContentTemplateActionSettingsMap, "sms_content_template_id", setContentTemplateActionSettings.SmsContentTemplateId)
		resourcedata.SetMapValueIfNotNil(setContentTemplateActionSettingsMap, "email_content_template_id", setContentTemplateActionSettings.EmailContentTemplateId)

		setContentTemplateActionSettingsList = append(setContentTemplateActionSettingsList, setContentTemplateActionSettingsMap)
	}

	return setContentTemplateActionSettingsList
}

// flattenSetSmsPhoneNumberActionSettingss maps a Genesys Cloud *[]platformclientv2.Setsmsphonenumberactionsettings into a []interface{}
func flattenSetSmsPhoneNumberActionSettingss(setSmsPhoneNumberActionSettingss *[]platformclientv2.Setsmsphonenumberactionsettings) []interface{} {
	if len(*setSmsPhoneNumberActionSettingss) == 0 {
		return nil
	}

	var setSmsPhoneNumberActionSettingsList []interface{}
	for _, setSmsPhoneNumberActionSettings := range *setSmsPhoneNumberActionSettingss {
		setSmsPhoneNumberActionSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(setSmsPhoneNumberActionSettingsMap, "sender_sms_phone_number", setSmsPhoneNumberActionSettings.SenderSmsPhoneNumber)

		setSmsPhoneNumberActionSettingsList = append(setSmsPhoneNumberActionSettingsList, setSmsPhoneNumberActionSettingsMap)
	}

	return setSmsPhoneNumberActionSettingsList
}

// flattenDigitalActions maps a Genesys Cloud *[]platformclientv2.Digitalaction into a []interface{}
func flattenDigitalActions(digitalActions *[]platformclientv2.Digitalaction) []interface{} {
	if len(*digitalActions) == 0 {
		return nil
	}

	var digitalActionList []interface{}
	// for _, digitalAction := range *digitalActions {
	// 	digitalActionMap := make(map[string]interface{})

	// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalActionMap, "update_contact_column_action_settings", digitalAction.UpdateContactColumnActionSettings, flattenUpdateContactColumnActionSettings)
	// // TODO: Handle do_not_send_action_settings property
	// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalActionMap, "append_to_dnc_action_settings", digitalAction.AppendToDncActionSettings, flattenAppendToDncActionSettings)
	// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalActionMap, "mark_contact_uncontactable_action_settings", digitalAction.MarkContactUncontactableActionSettings, flattenMarkContactUncontactableActionSettings)
	// // TODO: Handle mark_contact_address_uncontactable_action_settings property
	// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalActionMap, "set_content_template_action_settings", digitalAction.SetContentTemplateActionSettings, flattenSetContentTemplateActionSettings)
	// resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalActionMap, "set_sms_phone_number_action_settings", digitalAction.SetSmsPhoneNumberActionSettings, flattenSetSmsPhoneNumberActionSettings)

	// 	digitalActionList = append(digitalActionList, digitalActionMap)
	// }

	return digitalActionList
}

// flattenDigitalRules maps a Genesys Cloud *[]platformclientv2.Digitalrule into a []interface{}
func flattenDigitalRules(digitalRules *[]platformclientv2.Digitalrule) []interface{} {
	if len(*digitalRules) == 0 {
		return nil
	}

	var digitalRuleList []interface{}
	for _, digitalRule := range *digitalRules {
		digitalRuleMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(digitalRuleMap, "name", digitalRule.Name)
		resourcedata.SetMapValueIfNotNil(digitalRuleMap, "order", digitalRule.Order)
		resourcedata.SetMapValueIfNotNil(digitalRuleMap, "category", digitalRule.Category)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalRuleMap, "conditions", digitalRule.Conditions, flattenDigitalConditions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(digitalRuleMap, "actions", digitalRule.Actions, flattenDigitalActions)

		digitalRuleList = append(digitalRuleList, digitalRuleMap)
	}

	return digitalRuleList
}
