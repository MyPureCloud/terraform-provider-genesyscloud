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

		resourcedata.BuildSDKStringMapValueIfNotNil(&sdkUpdateContactColumnActionSettings.Properties, updateContactColumnActionSettingsMap, "properties")
	}

	return &sdkUpdateContactColumnActionSettings
}

// buildAppendToDncActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Appendtodncactionsettings
func buildAppendToDncActionSettings(appendToDncActionSettings *schema.Set) *platformclientv2.Appendtodncactionsettings {
	if appendToDncActionSettings == nil {
		return nil
	}

	var sdkAppendToDncActionSettings platformclientv2.Appendtodncactionsettings
	appendToDncActionSettingsList := appendToDncActionSettings.List()

	if len(appendToDncActionSettingsList) > 0 {
		appendToDncActionSettingsMap := appendToDncActionSettingsList[0].(map[string]interface{})

		sdkAppendToDncActionSettings.Expire = platformclientv2.Bool(appendToDncActionSettingsMap["expire"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAppendToDncActionSettings.ExpirationDuration, appendToDncActionSettingsMap, "expiration_duration")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAppendToDncActionSettings.ListType, appendToDncActionSettingsMap, "list_type")
	}

	return &sdkAppendToDncActionSettings
}

// buildMarkContactUncontactableActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Markcontactuncontactableactionsettings
func buildMarkContactUncontactableActionSettings(markContactUncontactableActionSettings *schema.Set) *platformclientv2.Markcontactuncontactableactionsettings {
	if markContactUncontactableActionSettings == nil {
		return nil
	}

	var sdkMarkContactUncontactableActionSettings platformclientv2.Markcontactuncontactableactionsettings
	markContactUncontactableActionSettingsList := markContactUncontactableActionSettings.List()

	if len(markContactUncontactableActionSettingsList) > 0 {
		markContactUncontactableActionSettingsMap := markContactUncontactableActionSettingsList[0].(map[string]interface{})

		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkMarkContactUncontactableActionSettings.MediaTypes, markContactUncontactableActionSettingsMap, "media_types")

	}

	return &sdkMarkContactUncontactableActionSettings
}

// buildSetContentTemplateActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Setcontenttemplateactionsettings
func buildSetContentTemplateActionSettings(setContentTemplateActionSettings *schema.Set) *platformclientv2.Setcontenttemplateactionsettings {
	if setContentTemplateActionSettings == nil {
		return nil
	}

	var sdkSetContentTemplateActionSettings platformclientv2.Setcontenttemplateactionsettings
	setContentTemplateActionSettingsList := setContentTemplateActionSettings.List()

	if len(setContentTemplateActionSettingsList) > 0 {
		setContentTemplateActionSettingsMap := setContentTemplateActionSettingsList[0].(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSetContentTemplateActionSettings.SmsContentTemplateId, setContentTemplateActionSettingsMap, "sms_content_template_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSetContentTemplateActionSettings.EmailContentTemplateId, setContentTemplateActionSettingsMap, "email_content_template_id")

	}

	return &sdkSetContentTemplateActionSettings
}

// buildSetSmsPhoneNumberActionSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Setsmsphonenumberactionsettings
func buildSetSmsPhoneNumberActionSettings(setSmsPhoneNumberActionSettings *schema.Set) *platformclientv2.Setsmsphonenumberactionsettings {
	if setSmsPhoneNumberActionSettings == nil {
		return nil
	}
	var sdkSetSmsPhoneNumberActionSettings platformclientv2.Setsmsphonenumberactionsettings
	setSmsPhoneNumberActionSettingsList := setSmsPhoneNumberActionSettings.List()
	if len(setSmsPhoneNumberActionSettingsList) > 0 {
		setSmsPhoneNumberActionSettingsMap := setSmsPhoneNumberActionSettingsList[0].(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSetSmsPhoneNumberActionSettings.SenderSmsPhoneNumber, setSmsPhoneNumberActionSettingsMap, "sender_sms_phone_number")

	}

	return &sdkSetSmsPhoneNumberActionSettings
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
		if action := digitalActionsMap["do_not_send_action_settings"]; action != nil {
			sdkDigitalAction.DoNotSendActionSettings = &action
		}

		sdkDigitalAction.AppendToDncActionSettings = buildAppendToDncActionSettings(digitalActionsMap["append_to_dnc_action_settings"].(*schema.Set))
		sdkDigitalAction.MarkContactUncontactableActionSettings = buildMarkContactUncontactableActionSettings(digitalActionsMap["mark_contact_uncontactable_action_settings"].(*schema.Set))
		if action := digitalActionsMap["mark_contact_address_uncontactable_action_settings"]; action != nil {
			sdkDigitalAction.MarkContactAddressUncontactableActionSettings = &action
		}

		sdkDigitalAction.SetContentTemplateActionSettings = buildSetContentTemplateActionSettings(digitalActionsMap["set_content_template_action_settings"].(*schema.Set))
		sdkDigitalAction.SetSmsPhoneNumberActionSettings = buildSetSmsPhoneNumberActionSettings(digitalActionsMap["set_sms_phone_number_action_settings"].(*schema.Set))

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
func flattenContactColumnConditionSettings(contactColumnConditionSettings *platformclientv2.Contactcolumnconditionsettings) *schema.Set {
	if contactColumnConditionSettings == nil {
		return nil
	}

	contactColumnConditionSettingsSet := schema.NewSet(schema.HashResource(contactColumnConditionSettingsResource), []interface{}{})
	contactColumnConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "column_name", contactColumnConditionSettings.ColumnName)
	resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "operator", contactColumnConditionSettings.Operator)
	resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "value", contactColumnConditionSettings.Value)
	resourcedata.SetMapValueIfNotNil(contactColumnConditionSettingsMap, "value_type", contactColumnConditionSettings.ValueType)

	contactColumnConditionSettingsSet.Add(contactColumnConditionSettingsMap)

	return contactColumnConditionSettingsSet
}

// flattenContactAddressConditionSettingss maps a Genesys Cloud *[]platformclientv2.Contactaddressconditionsettings into a []interface{}
func flattenContactAddressConditionSettings(contactAddressConditionSettings *platformclientv2.Contactaddressconditionsettings) *schema.Set {
	if contactAddressConditionSettings == nil {
		return nil
	}

	contactAddressConditionSettingsSet := schema.NewSet(schema.HashResource(contactAddressConditionSettingsResource), []interface{}{})
	contactAddressConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(contactAddressConditionSettingsMap, "operator", contactAddressConditionSettings.Operator)
	resourcedata.SetMapValueIfNotNil(contactAddressConditionSettingsMap, "value", contactAddressConditionSettings.Value)

	contactAddressConditionSettingsSet.Add(contactAddressConditionSettingsMap)
	return contactAddressConditionSettingsSet
}

// flattenContactAddressTypeConditionSettingss maps a Genesys Cloud *[]platformclientv2.Contactaddresstypeconditionsettings into a []interface{}
func flattenContactAddressTypeConditionSettings(contactAddressTypeConditionSettings *platformclientv2.Contactaddresstypeconditionsettings) *schema.Set {
	if contactAddressTypeConditionSettings == nil {
		return nil
	}

	contactAddressTypeConditionSettingsSet := schema.NewSet(schema.HashResource(contactAddressTypeConditionSettingsResource), []interface{}{})
	contactAddressTypeConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(contactAddressTypeConditionSettingsMap, "operator", contactAddressTypeConditionSettings.Operator)
	resourcedata.SetMapValueIfNotNil(contactAddressTypeConditionSettingsMap, "value", contactAddressTypeConditionSettings.Value)

	contactAddressTypeConditionSettingsSet.Add(contactAddressTypeConditionSettingsMap)

	return contactAddressTypeConditionSettingsSet
}

// flattenLastAttemptByColumnConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastattemptbycolumnconditionsettings into a []interface{}
func flattenLastAttemptByColumnConditionSettings(lastAttemptByColumnConditionSettings *platformclientv2.Lastattemptbycolumnconditionsettings) *schema.Set {
	if lastAttemptByColumnConditionSettings == nil {
		return nil
	}

	lastAttemptByColumnConditionSettingsSet := schema.NewSet(schema.HashResource(lastAttemptByColumnConditionSettingsResource), []interface{}{})
	lastAttemptByColumnConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "email_column_name", lastAttemptByColumnConditionSettings.EmailColumnName)
	resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "sms_column_name", lastAttemptByColumnConditionSettings.SmsColumnName)
	resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "operator", lastAttemptByColumnConditionSettings.Operator)
	resourcedata.SetMapValueIfNotNil(lastAttemptByColumnConditionSettingsMap, "value", lastAttemptByColumnConditionSettings.Value)

	lastAttemptByColumnConditionSettingsSet.Add(lastAttemptByColumnConditionSettingsMap)
	return lastAttemptByColumnConditionSettingsSet
}

// flattenLastAttemptOverallConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastattemptoverallconditionsettings into a []interface{}
func flattenLastAttemptOverallConditionSettings(lastAttemptOverallConditionSettings *platformclientv2.Lastattemptoverallconditionsettings) *schema.Set {
	if lastAttemptOverallConditionSettings == nil {
		return nil
	}

	lastAttemptOverallConditionSettingsSet := schema.NewSet(schema.HashResource(lastAttemptOverallConditionSettingsResource), []interface{}{})
	lastAttemptOverallConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapStringArrayValueIfNotNil(lastAttemptOverallConditionSettingsMap, "media_types", lastAttemptOverallConditionSettings.MediaTypes)
	resourcedata.SetMapValueIfNotNil(lastAttemptOverallConditionSettingsMap, "operator", lastAttemptOverallConditionSettings.Operator)
	resourcedata.SetMapValueIfNotNil(lastAttemptOverallConditionSettingsMap, "value", lastAttemptOverallConditionSettings.Value)

	lastAttemptOverallConditionSettingsSet.Add(lastAttemptOverallConditionSettingsMap)
	return lastAttemptOverallConditionSettingsSet
}

// flattenLastResultByColumnConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastresultbycolumnconditionsettings into a []interface{}
func flattenLastResultByColumnConditionSettings(lastResultByColumnConditionSettings *platformclientv2.Lastresultbycolumnconditionsettings) *schema.Set {
	if lastResultByColumnConditionSettings == nil {
		return nil
	}

	lastResultByColumnConditionSettingsSet := schema.NewSet(schema.HashResource(lastResultByColumnConditionSettingsResource), []interface{}{})
	lastResultByColumnConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(lastResultByColumnConditionSettingsMap, "email_column_name", lastResultByColumnConditionSettings.EmailColumnName)
	resourcedata.SetMapStringArrayValueIfNotNil(lastResultByColumnConditionSettingsMap, "email_wrapup_codes", lastResultByColumnConditionSettings.EmailWrapupCodes)
	resourcedata.SetMapValueIfNotNil(lastResultByColumnConditionSettingsMap, "sms_column_name", lastResultByColumnConditionSettings.SmsColumnName)
	resourcedata.SetMapStringArrayValueIfNotNil(lastResultByColumnConditionSettingsMap, "sms_wrapup_codes", lastResultByColumnConditionSettings.SmsWrapupCodes)

	lastResultByColumnConditionSettingsSet.Add(lastResultByColumnConditionSettingsMap)
	return lastResultByColumnConditionSettingsSet
}

// flattenLastResultOverallConditionSettingss maps a Genesys Cloud *[]platformclientv2.Lastresultoverallconditionsettings into a []interface{}
func flattenLastResultOverallConditionSettings(lastResultOverallConditionSettings *platformclientv2.Lastresultoverallconditionsettings) *schema.Set {
	if lastResultOverallConditionSettings == nil {
		return nil
	}

	lastResultOverallConditionSettingsSet := schema.NewSet(schema.HashResource(lastResultOverallConditionSettingsResource), []interface{}{})
	lastResultOverallConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapStringArrayValueIfNotNil(lastResultOverallConditionSettingsMap, "email_wrapup_codes", lastResultOverallConditionSettings.EmailWrapupCodes)
	resourcedata.SetMapStringArrayValueIfNotNil(lastResultOverallConditionSettingsMap, "sms_wrapup_codes", lastResultOverallConditionSettings.SmsWrapupCodes)

	lastResultOverallConditionSettingsSet.Add(lastResultOverallConditionSettingsMap)
	return lastResultOverallConditionSettingsSet
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
func flattenDataActionConditionSettings(dataActionConditionSettings *platformclientv2.Dataactionconditionsettings) *schema.Set {
	if dataActionConditionSettings == nil {
		return nil
	}

	dataActionConditionSettingsSet := schema.NewSet(schema.HashResource(dataActionConditionSettingsResource), []interface{}{})
	dataActionConditionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(dataActionConditionSettingsMap, "data_action_id", dataActionConditionSettings.DataActionId)
	resourcedata.SetMapValueIfNotNil(dataActionConditionSettingsMap, "contact_id_field", dataActionConditionSettings.ContactIdField)
	resourcedata.SetMapValueIfNotNil(dataActionConditionSettingsMap, "data_not_found_resolution", dataActionConditionSettings.DataNotFoundResolution)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(dataActionConditionSettingsMap, "predicates", dataActionConditionSettings.Predicates, flattenDigitalDataActionConditionPredicates)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(dataActionConditionSettingsMap, "contact_column_to_data_action_field_mappings", dataActionConditionSettings.ContactColumnToDataActionFieldMappings, flattenDataActionContactColumnFieldMappings)

	dataActionConditionSettingsSet.Add(dataActionConditionSettingsMap)
	return dataActionConditionSettingsSet
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

		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "contact_column_condition_settings", digitalCondition.ContactColumnConditionSettings, flattenContactColumnConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "contact_address_condition_settings", digitalCondition.ContactAddressConditionSettings, flattenContactAddressConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "contact_address_type_condition_settings", digitalCondition.ContactAddressTypeConditionSettings, flattenContactAddressTypeConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "last_attempt_by_column_condition_settings", digitalCondition.LastAttemptByColumnConditionSettings, flattenLastAttemptByColumnConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "last_attempt_overall_condition_settings", digitalCondition.LastAttemptOverallConditionSettings, flattenLastAttemptOverallConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "last_result_by_column_condition_settings", digitalCondition.LastResultByColumnConditionSettings, flattenLastResultByColumnConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "last_result_overall_condition_settings", digitalCondition.LastResultOverallConditionSettings, flattenLastResultOverallConditionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalConditionMap, "data_action_condition_settings", digitalCondition.DataActionConditionSettings, flattenDataActionConditionSettings)

		digitalConditionList = append(digitalConditionList, digitalConditionMap)
	}

	return digitalConditionList
}

// flattenUpdateContactColumnActionSettingss maps a Genesys Cloud *[]platformclientv2.Updatecontactcolumnactionsettings into a []interface{}
func flattenUpdateContactColumnActionSettings(updateContactColumnActionSettings *platformclientv2.Updatecontactcolumnactionsettings) *schema.Set {
	if updateContactColumnActionSettings == nil {
		return nil
	}

	updateContactColumnActionSettingsSet := schema.NewSet(schema.HashResource(updateContactColumnActionSettingsResource), []interface{}{})
	updateContactColumnActionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapStringMapValueIfNotNil(updateContactColumnActionSettingsMap, "properties", updateContactColumnActionSettings.Properties)
	resourcedata.SetMapValueIfNotNil(updateContactColumnActionSettingsMap, "update_option", updateContactColumnActionSettings.UpdateOption)

	updateContactColumnActionSettingsSet.Add(updateContactColumnActionSettingsMap)
	return updateContactColumnActionSettingsSet
}

// flattenAppendToDncActionSettingss maps a Genesys Cloud *[]platformclientv2.Appendtodncactionsettings into a []interface{}
func flattenAppendToDncActionSettings(appendToDncActionSettings *platformclientv2.Appendtodncactionsettings) *schema.Set {
	if appendToDncActionSettings == nil {
		return nil
	}

	appendToDncActionSettingsSet := schema.NewSet(schema.HashResource(appendToDncActionSettingsResource), []interface{}{})
	appendToDncActionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(appendToDncActionSettingsMap, "expire", appendToDncActionSettings.Expire)
	resourcedata.SetMapValueIfNotNil(appendToDncActionSettingsMap, "expiration_duration", appendToDncActionSettings.ExpirationDuration)
	resourcedata.SetMapValueIfNotNil(appendToDncActionSettingsMap, "list_type", appendToDncActionSettings.ListType)

	appendToDncActionSettingsSet.Add(appendToDncActionSettingsMap)
	return appendToDncActionSettingsSet
}

// flattenMarkContactUncontactableActionSettingss maps a Genesys Cloud *[]platformclientv2.Markcontactuncontactableactionsettings into a []interface{}
func flattenMarkContactUncontactableActionSettings(markContactUncontactableActionSettings *platformclientv2.Markcontactuncontactableactionsettings) *schema.Set {
	if markContactUncontactableActionSettings == nil {
		return nil
	}

	markContactUncontactableActionSettingsSet := schema.NewSet(schema.HashResource(markContactUncontactableActionSettingsResource), []interface{}{})
	markContactUncontactableActionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapStringArrayValueIfNotNil(markContactUncontactableActionSettingsMap, "media_types", markContactUncontactableActionSettings.MediaTypes)

	markContactUncontactableActionSettingsSet.Add(markContactUncontactableActionSettingsMap)
	return markContactUncontactableActionSettingsSet
}

// flattenSetContentTemplateActionSettingss maps a Genesys Cloud *[]platformclientv2.Setcontenttemplateactionsettings into a []interface{}
func flattenSetContentTemplateActionSettings(setContentTemplateActionSettings *platformclientv2.Setcontenttemplateactionsettings) *schema.Set {
	if setContentTemplateActionSettings == nil {
		return nil
	}

	setContentTemplateActionSettingsSet := schema.NewSet(schema.HashResource(setContentTemplateActionSettingsResource), []interface{}{})
	setContentTemplateActionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(setContentTemplateActionSettingsMap, "sms_content_template_id", setContentTemplateActionSettings.SmsContentTemplateId)
	resourcedata.SetMapValueIfNotNil(setContentTemplateActionSettingsMap, "email_content_template_id", setContentTemplateActionSettings.EmailContentTemplateId)

	setContentTemplateActionSettingsSet.Add(setContentTemplateActionSettingsMap)
	return setContentTemplateActionSettingsSet
}

// flattenSetSmsPhoneNumberActionSettingss maps a Genesys Cloud *[]platformclientv2.Setsmsphonenumberactionsettings into a []interface{}
func flattenSetSmsPhoneNumberActionSettings(setSmsPhoneNumberActionSettings *platformclientv2.Setsmsphonenumberactionsettings) *schema.Set {
	if setSmsPhoneNumberActionSettings == nil {
		return nil
	}

	setSmsPhoneNumberActionSettingsSet := schema.NewSet(schema.HashResource(setSmsPhoneNumberActionSettingsResource), []interface{}{})
	setSmsPhoneNumberActionSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(setSmsPhoneNumberActionSettingsMap, "sender_sms_phone_number", setSmsPhoneNumberActionSettings.SenderSmsPhoneNumber)

	setSmsPhoneNumberActionSettingsSet.Add(setSmsPhoneNumberActionSettingsMap)
	return setSmsPhoneNumberActionSettingsSet
}

// flattenDigitalActions maps a Genesys Cloud *[]platformclientv2.Digitalaction into a []interface{}
func flattenDigitalActions(digitalActions *[]platformclientv2.Digitalaction) []interface{} {
	if len(*digitalActions) == 0 {
		return nil
	}

	var digitalActionList []interface{}
	for _, digitalAction := range *digitalActions {
		digitalActionMap := make(map[string]interface{})

		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalActionMap, "update_contact_column_action_settings", digitalAction.UpdateContactColumnActionSettings, flattenUpdateContactColumnActionSettings)
		resourcedata.SetMapValueIfNotNil(digitalActionMap, "do_not_send_action_settings", digitalAction.DoNotSendActionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalActionMap, "append_to_dnc_action_settings", digitalAction.AppendToDncActionSettings, flattenAppendToDncActionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalActionMap, "mark_contact_uncontactable_action_settings", digitalAction.MarkContactUncontactableActionSettings, flattenMarkContactUncontactableActionSettings)
		resourcedata.SetMapValueIfNotNil(digitalActionMap, "mark_contact_address_uncontactable_action_settings", digitalAction.DoNotSendActionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalActionMap, "set_content_template_action_settings", digitalAction.SetContentTemplateActionSettings, flattenSetContentTemplateActionSettings)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(digitalActionMap, "set_sms_phone_number_action_settings", digitalAction.SetSmsPhoneNumberActionSettings, flattenSetSmsPhoneNumberActionSettings)

		digitalActionList = append(digitalActionList, digitalActionMap)
	}

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
