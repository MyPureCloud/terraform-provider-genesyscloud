package outbound_messagingcampaign

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_outbound_messagingcampaign_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getOutboundMessagingcampaignFromResourceData maps data from schema ResourceData object to a platformclientv2.Messagingcampaign
func getOutboundMessagingcampaignFromResourceData(d *schema.ResourceData) platformclientv2.Messagingcampaign {

	return platformclientv2.Messagingcampaign{
		Name:                           platformclientv2.String(d.Get("name").(string)),
		Division:                       util.BuildSdkDomainEntityRef(d, "division_id"),
		CampaignStatus:                 platformclientv2.String(d.Get("campaign_status").(string)),
		CallableTimeSet:                util.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		ContactList:                    util.BuildSdkDomainEntityRef(d, "contact_list_id"),
		DncLists:                       util.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		AlwaysRunning:                  platformclientv2.Bool(d.Get("always_running").(bool)),
		ContactSorts:                   buildContactSorts(d.Get("contact_sorts").([]interface{})),
		MessagesPerMinute:              platformclientv2.Int(d.Get("messages_per_minute").(int)),
		RuleSets:                       util.BuildSdkDomainEntityRefArr(d, "rule_set_ids"),
		ContactListFilters:             util.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		Errors:                         buildRestErrorDetails(d.Get("errors").([]interface{})),
		DynamicContactQueueingSettings: buildDynamicContactQueueingSettingss(d.Get("dynamic_contact_queueing_settings").([]interface{})),
		SmsConfig:                      buildSmsConfigs(d.Get("sms_config").(*schema.Set)),
		EmailConfig:                    buildEmailConfigs(d.Get("email_config").(*schema.Set)),
	}
}

// buildContactSorts maps an []interface{} into a Genesys Cloud *[]platformclientv2.Contactsort
func buildContactSorts(contactSorts []interface{}) *[]platformclientv2.Contactsort {
	if contactSorts == nil {
		return nil
	}
	sdkContactSortSlice := make([]platformclientv2.Contactsort, 0)
	for _, configContactSort := range contactSorts {
		var sdkContactSort platformclientv2.Contactsort
		contactsortMap := configContactSort.(map[string]interface{})
		if fieldName := contactsortMap["field_name"].(string); fieldName != "" {
			sdkContactSort.FieldName = &fieldName
		}
		if direction := contactsortMap["direction"].(string); direction != "" {
			sdkContactSort.Direction = &direction
		}
		if numeric, ok := contactsortMap["numeric"].(bool); ok {
			sdkContactSort.Numeric = platformclientv2.Bool(numeric)
		}

		sdkContactSortSlice = append(sdkContactSortSlice, sdkContactSort)
	}
	return &sdkContactSortSlice
}

// buildRestErrorDetails maps an []interface{} into a Genesys Cloud *[]platformclientv2.Resterrordetail
func buildRestErrorDetails(restErrorDetails []interface{}) *[]platformclientv2.Resterrordetail {
	restErrorDetailsSlice := make([]platformclientv2.Resterrordetail, 0)
	for _, restErrorDetail := range restErrorDetails {
		var sdkRestErrorDetail platformclientv2.Resterrordetail
		restErrorDetailsMap, ok := restErrorDetail.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkRestErrorDetail.VarError, restErrorDetailsMap, "error")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkRestErrorDetail.Details, restErrorDetailsMap, "details")

		restErrorDetailsSlice = append(restErrorDetailsSlice, sdkRestErrorDetail)
	}

	return &restErrorDetailsSlice
}

// buildDynamicContactQueueingSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Dynamiccontactqueueingsettings
func buildDynamicContactQueueingSettingss(settings []interface{}) *platformclientv2.Dynamiccontactqueueingsettings {
	if settings == nil || len(settings) < 1 {
		return nil
	}
	var sdkDcqSettings platformclientv2.Dynamiccontactqueueingsettings
	dcqSetting, ok := settings[0].(map[string]interface{})
	if !ok {
		return nil
	}
	if sort, ok := dcqSetting["sort"].(bool); ok {
		sdkDcqSettings.Sort = &sort
	}
	if filter, ok := dcqSetting["filter"].(bool); ok {
		sdkDcqSettings.Filter = &filter
	}
	return &sdkDcqSettings
}

// buildFromEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Fromemailaddress
func buildFromEmailAddresss(fromEmailAddresss []interface{}) *platformclientv2.Fromemailaddress {
	fromEmailAddresssSlice := make([]platformclientv2.Fromemailaddress, 0)
	for _, fromEmailAddress := range fromEmailAddresss {
		var sdkFromEmailAddress platformclientv2.Fromemailaddress
		fromEmailAddresssMap, ok := fromEmailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		sdkFromEmailAddress.Domain = &platformclientv2.Domainentityref{Id: platformclientv2.String(fromEmailAddresssMap["domain_id"].(string))}
		resourcedata.BuildSDKStringValueIfNotNil(&sdkFromEmailAddress.FriendlyName, fromEmailAddresssMap, "friendly_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkFromEmailAddress.LocalPart, fromEmailAddresssMap, "local_part")

		fromEmailAddresssSlice = append(fromEmailAddresssSlice, sdkFromEmailAddress)
	}

	return &fromEmailAddresssSlice[0]
}

// buildReplyToEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Replytoemailaddress
func buildReplyToEmailAddresss(replyToEmailAddress []interface{}) *platformclientv2.Replytoemailaddress {
	if replyToEmailAddress == nil || len(replyToEmailAddress) < 1 {
		return nil
	}
	replyToEmailAddresssSlice := make([]platformclientv2.Replytoemailaddress, 0)
	for _, replyToEmailAddress := range replyToEmailAddress {
		var sdkReplyToEmailAddress platformclientv2.Replytoemailaddress
		replyToEmailAddresssMap, ok := replyToEmailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		sdkReplyToEmailAddress.Domain = &platformclientv2.Domainentityref{Id: platformclientv2.String(replyToEmailAddresssMap["domain_id"].(string))}
		sdkReplyToEmailAddress.Route = &platformclientv2.Domainentityref{Id: platformclientv2.String(replyToEmailAddresssMap["route_id"].(string))}

		replyToEmailAddresssSlice = append(replyToEmailAddresssSlice, sdkReplyToEmailAddress)
	}

	return &replyToEmailAddresssSlice[0]
}

// buildEmailConfigs maps an []interface{} into a Genesys Cloud *[]platformclientv2.Emailconfig
func buildEmailConfigs(emailConfigs *schema.Set) *platformclientv2.Emailconfig {
	if emailConfigs == nil || emailConfigs.Len() < 1 {
		return nil
	}
	var sdkEmailConfig platformclientv2.Emailconfig
	emailConfigList := emailConfigs.List()
	if len(emailConfigList) > 0 {
		emailConfigsMap := emailConfigList[0].(map[string]interface{})
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkEmailConfig.EmailColumns, emailConfigsMap, "email_columns")
		sdkEmailConfig.ContentTemplate = &platformclientv2.Domainentityref{Id: platformclientv2.String(emailConfigsMap["content_template_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkEmailConfig.FromAddress, emailConfigsMap, "from_address", buildFromEmailAddresss)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkEmailConfig.ReplyToAddress, emailConfigsMap, "reply_to_address", buildReplyToEmailAddresss)

		return &sdkEmailConfig
	}
	return nil
}

// buildSmsConfigs maps an []interface{} into a Genesys Cloud *platformclientv2.Smsconfig
func buildSmsConfigs(smsConfigs *schema.Set) *platformclientv2.Smsconfig {
	if smsConfigs == nil || smsConfigs.Len() < 1 {
		return nil
	}
	var sdkSmsconfig platformclientv2.Smsconfig
	smsconfigList := smsConfigs.List()
	if len(smsconfigList) > 0 {
		smsconfigMap := smsconfigList[0].(map[string]interface{})
		if messageColumn := smsconfigMap["message_column"].(string); messageColumn != "" {
			sdkSmsconfig.MessageColumn = &messageColumn
		}
		if phoneColumn := smsconfigMap["phone_column"].(string); phoneColumn != "" {
			sdkSmsconfig.PhoneColumn = &phoneColumn
		}
		if senderSmsPhoneNumber := smsconfigMap["sender_sms_phone_number"].(string); senderSmsPhoneNumber != "" {
			sdkSmsconfig.SenderSmsPhoneNumber = &platformclientv2.Smsphonenumberref{
				PhoneNumber: &senderSmsPhoneNumber,
			}
		}
		if contentTemplateId := smsconfigMap["content_template_id"].(string); contentTemplateId != "" {
			sdkSmsconfig.ContentTemplate = &platformclientv2.Domainentityref{Id: &contentTemplateId}
		}
		return &sdkSmsconfig
	}
	return nil
}

// flattenContactSorts maps a Genesys Cloud *[]platformclientv2.Contactsort into a []interface{}
func flattenContactSorts(contactSorts *[]platformclientv2.Contactsort) []interface{} {
	if len(*contactSorts) == 0 {
		return nil
	}

	var contactSortList []interface{}
	for _, contactSort := range *contactSorts {
		contactSortMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactSortMap, "field_name", contactSort.FieldName)
		resourcedata.SetMapValueIfNotNil(contactSortMap, "direction", contactSort.Direction)
		resourcedata.SetMapValueIfNotNil(contactSortMap, "numeric", contactSort.Numeric)

		contactSortList = append(contactSortList, contactSortMap)
	}

	return contactSortList
}

// flattenRestErrorDetails maps a Genesys Cloud *[]platformclientv2.Resterrordetail into a []interface{}
func flattenRestErrorDetails(restErrorDetails *[]platformclientv2.Resterrordetail) []interface{} {
	if len(*restErrorDetails) == 0 {
		return nil
	}

	var restErrorDetailList []interface{}
	for _, restErrorDetail := range *restErrorDetails {
		restErrorDetailMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(restErrorDetailMap, "error", restErrorDetail.VarError)
		resourcedata.SetMapValueIfNotNil(restErrorDetailMap, "details", restErrorDetail.Details)

		restErrorDetailList = append(restErrorDetailList, restErrorDetailMap)
	}

	return restErrorDetailList
}

// flattenDynamicContactQueueingSettingss maps a Genesys Cloud *[]platformclientv2.Dynamiccontactqueueingsettings into a []interface{}
func flattenDynamicContactQueueingSettingss(dynamicContactQueueingSettingss *platformclientv2.Dynamiccontactqueueingsettings) []interface{} {
	if dynamicContactQueueingSettingss == nil {
		return nil
	}

	var dynamicContactQueueingSettingsList []interface{}
	dynamicContactQueueingSettingsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(dynamicContactQueueingSettingsMap, "sort", dynamicContactQueueingSettingss.Sort)
	resourcedata.SetMapValueIfNotNil(dynamicContactQueueingSettingsMap, "filter", dynamicContactQueueingSettingss.Filter)

	dynamicContactQueueingSettingsList = append(dynamicContactQueueingSettingsList, dynamicContactQueueingSettingsMap)

	return dynamicContactQueueingSettingsList
}

// flattenFromEmailAddresss maps a Genesys Cloud *[]platformclientv2.Fromemailaddress into a []interface{}
func flattenFromEmailAddresss(fromEmailAddresss *platformclientv2.Fromemailaddress) []interface{} {
	if fromEmailAddresss == nil {
		return nil
	}

	var fromEmailAddressList []interface{}
	fromEmailAddressMap := make(map[string]interface{})

	resourcedata.SetMapReferenceValueIfNotNil(fromEmailAddressMap, "domain_id", fromEmailAddresss.Domain)
	resourcedata.SetMapValueIfNotNil(fromEmailAddressMap, "friendly_name", fromEmailAddresss.FriendlyName)
	resourcedata.SetMapValueIfNotNil(fromEmailAddressMap, "local_part", fromEmailAddresss.LocalPart)
	fromEmailAddressList = append(fromEmailAddressList, fromEmailAddressMap)
	return fromEmailAddressList
}

// flattenReplyToEmailAddresss maps a Genesys Cloud *[]platformclientv2.Replytoemailaddress into a []interface{}
func flattenReplyToEmailAddresss(replyToEmailAddresss *platformclientv2.Replytoemailaddress) []interface{} {
	if replyToEmailAddresss == nil {
		return nil
	}

	var replyToEmailAddressList []interface{}
	replyToEmailAddressMap := make(map[string]interface{})

	resourcedata.SetMapReferenceValueIfNotNil(replyToEmailAddressMap, "domain_id", replyToEmailAddresss.Domain)
	resourcedata.SetMapReferenceValueIfNotNil(replyToEmailAddressMap, "route_id", replyToEmailAddresss.Route)
	replyToEmailAddressList = append(replyToEmailAddressList, replyToEmailAddressMap)
	return replyToEmailAddressList
}

// flattenEmailConfigs maps a Genesys Cloud *[]platformclientv2.Emailconfig into a []interface{}
func flattenEmailConfigs(emailConfigs *platformclientv2.Emailconfig) []interface{} {
	if emailConfigs == nil {
		return nil
	}

	var emailConfigList []interface{}
	emailConfigMap := make(map[string]interface{})

	resourcedata.SetMapStringArrayValueIfNotNil(emailConfigMap, "email_columns", emailConfigs.EmailColumns)
	resourcedata.SetMapReferenceValueIfNotNil(emailConfigMap, "content_template_id", emailConfigs.ContentTemplate)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(emailConfigMap, "from_address", emailConfigs.FromAddress, flattenFromEmailAddresss)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(emailConfigMap, "reply_to_address", emailConfigs.ReplyToAddress, flattenReplyToEmailAddresss)

	emailConfigList = append(emailConfigList, emailConfigMap)

	return emailConfigList
}

// flattenSmsConfigs maps a Genesys Cloud *platformclientv2.Smsconfig into a []interface{}
func flattenSmsConfigs(smsconfig *platformclientv2.Smsconfig) *schema.Set {
	if smsconfig == nil {
		return nil
	}

	smsconfigSet := schema.NewSet(schema.HashResource(smsConfigResource), []interface{}{})
	smsconfigMap := make(map[string]interface{})

	if smsconfig.MessageColumn != nil {
		smsconfigMap["message_column"] = *smsconfig.MessageColumn
	}
	if smsconfig.PhoneColumn != nil {
		smsconfigMap["phone_column"] = *smsconfig.PhoneColumn
	}
	if smsconfig.SenderSmsPhoneNumber != nil {
		if smsconfig.SenderSmsPhoneNumber.PhoneNumber != nil {
			smsconfigMap["sender_sms_phone_number"] = *smsconfig.SenderSmsPhoneNumber.PhoneNumber
		}
	}
	if smsconfig.ContentTemplate != nil {
		smsconfigMap["content_template_id"] = *smsconfig.ContentTemplate.Id
	}

	smsconfigSet.Add(smsconfigMap)

	return smsconfigSet
}

func validateSmsconfig(smsconfig *schema.Set) error {
	if smsconfig == nil {
		return nil
	}

	smsconfigList := smsconfig.List()
	if len(smsconfigList) > 0 {
		smsconfigMap := smsconfigList[0].(map[string]interface{})
		messageColumn, _ := smsconfigMap["message_column"].(string)
		contentTemplateId, _ := smsconfigMap["content_template_id"].(string)
		if messageColumn == "" && contentTemplateId == "" {
			return fmt.Errorf("either message_column or content_template_id is required")
		} else if messageColumn != "" && contentTemplateId != "" {
			return fmt.Errorf("only one of message_column or content_template_id can be defined")
		}
	} else {
		return fmt.Errorf("error reading smsconfig")
	}

	return nil
}

func GenerateOutboundMessagingCampaignContactSort(fieldName string, direction string, numeric string) string {
	if direction != "" {
		direction = fmt.Sprintf(`direction = "%s"`, direction)
	}
	if numeric != "" {
		numeric = fmt.Sprintf(`numeric = %s`, numeric)
	}
	return fmt.Sprintf(`
	contact_sorts {
		field_name = "%s"
		%s
        %s
	}
`, fieldName, direction, numeric)
}

// GetOutboundDigitalrulesets invokes GET /api/v2/outbound/digitalrulesets
func GetOutboundDigitalRuleSets() (string, error) {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return "", err
	}

	outboundApi := platformclientv2.NewOutboundApiWithConfig(config)
	outboundRuleSets, _, err := outboundApi.GetOutboundDigitalrulesets(1, 1, "", "", "", []string{})

	if outboundRuleSets.Entities == nil || len(*outboundRuleSets.Entities) == 0 {
		return "", err
	}

	if len(*outboundRuleSets.Entities) > 0 {
		ruleSet := (*outboundRuleSets.Entities)[0]
		return *ruleSet.Id, nil
	}
	return "", err
}
