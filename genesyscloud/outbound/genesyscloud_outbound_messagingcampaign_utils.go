package outbound

import (
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func gatherExtraErrorMessagesFromResponseBody(resp *platformclientv2.APIResponse) (string, error) {
	if resp == nil || resp.RawBody == nil {
		return "", errors.New("no raw body to parse from API response")
	}

	type respBodyStructure struct {
		MessageParams map[string]string `json:"messageParams"`
	}

	var (
		extraDetails string
		rb           respBodyStructure
	)

	if err := json.Unmarshal(resp.RawBody, &rb); err != nil {
		return "", err
	}

	if rb.MessageParams != nil {
		for code, message := range rb.MessageParams {
			extraDetails += fmt.Sprintf("%s: %s\n", code, message)
		}
	}

	return extraDetails, nil
}

func buildSdkoutboundmessagingcampaignContactsortSlice(contactSort []interface{}) *[]platformclientv2.Contactsort {
	if contactSort == nil {
		return nil
	}
	sdkContactSortSlice := make([]platformclientv2.Contactsort, 0)
	for _, configContactSort := range contactSort {
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

func buildSmsconfig(d *schema.ResourceData) *platformclientv2.Smsconfig {
	smsConfigSet, ok := d.Get("sms_config").(*schema.Set)
	if !ok || len(smsConfigSet.List()) == 0 {
		return nil
	}
	var sdkSmsconfig platformclientv2.Smsconfig
	smsconfigList := smsConfigSet.List()
	if len(smsconfigList) > 0 {
		smsconfigMap, ok := smsconfigList[0].(map[string]interface{})
		if !ok {
			return nil
		}
		if messageColumn, _ := smsconfigMap["message_column"].(string); messageColumn != "" {
			sdkSmsconfig.MessageColumn = &messageColumn
		}
		if phoneColumn, _ := smsconfigMap["phone_column"].(string); phoneColumn != "" {
			sdkSmsconfig.PhoneColumn = &phoneColumn
		}
		if senderSmsPhoneNumber, _ := smsconfigMap["sender_sms_phone_number"].(string); senderSmsPhoneNumber != "" {
			sdkSmsconfig.SenderSmsPhoneNumber = &platformclientv2.Smsphonenumberref{
				PhoneNumber: &senderSmsPhoneNumber,
			}
		}
		if contentTemplateId, _ := smsconfigMap["content_template_id"].(string); contentTemplateId != "" {
			sdkSmsconfig.ContentTemplate = &platformclientv2.Domainentityref{Id: &contentTemplateId}
		}
	}

	return &sdkSmsconfig
}

func buildDynamicContactQueueingSettings(d *schema.ResourceData) *platformclientv2.Dynamiccontactqueueingsettings {
	settings, ok := d.Get("dynamic_contact_queueing_settings").([]any)
	if !ok || len(settings) == 0 {
		return nil
	}
	settingsMap, ok := settings[0].(map[string]any)
	if !ok {
		return nil
	}

	var dcqSettings platformclientv2.Dynamiccontactqueueingsettings

	if sort, ok := settingsMap["sort"].(bool); ok {
		dcqSettings.Sort = &sort
	}

	if filter, ok := settingsMap["filter"].(bool); ok {
		dcqSettings.Filter = &filter
	}

	return &dcqSettings
}

func buildEmailConfig(d *schema.ResourceData) *platformclientv2.Emailconfig {
	emailConfigList, ok := d.Get("email_config").([]any)
	if !ok || len(emailConfigList) == 0 {
		return nil
	}
	emailConfigMap, ok := emailConfigList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var emailConfig platformclientv2.Emailconfig

	if emailColumns := lists.BuildStringListFromSetInMap(emailConfigMap, "email_columns"); len(emailColumns) > 0 {
		emailConfig.EmailColumns = &emailColumns
	}

	if contentTemplateId, _ := emailConfigMap["content_template_id"].(string); contentTemplateId != "" {
		emailConfig.ContentTemplate = &platformclientv2.Domainentityref{Id: &contentTemplateId}
	}

	if fromAddress := buildFromAddress(emailConfigMap); fromAddress != nil {
		emailConfig.FromAddress = fromAddress
	}

	if replyToAddress := buildReplyToAddress(emailConfigMap); replyToAddress != nil {
		emailConfig.ReplyToAddress = replyToAddress
	}

	return &emailConfig
}

func flattenEmailConfig(emailConfig platformclientv2.Emailconfig) []any {
	emailConfigMap := make(map[string]any)

	resourcedata.SetMapReferenceValueIfNotNil(emailConfigMap, "content_template_id", emailConfig.ContentTemplate)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(emailConfigMap, "from_address", emailConfig.FromAddress, flattenFromAddress)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(emailConfigMap, "reply_to_address", emailConfig.ReplyToAddress, flattenReplyToAddress)
	if emailConfig.EmailColumns != nil {
		emailConfigMap["email_columns"] = lists.StringListToInterfaceList(*emailConfig.EmailColumns)
	}

	return []any{emailConfigMap}
}

func buildFromAddress(emailConfigMap map[string]any) *platformclientv2.Fromemailaddress {
	fromAddressList, ok := emailConfigMap["from_address"].([]any)
	if !ok || len(fromAddressList) == 0 {
		return nil
	}
	fromAddressMap, ok := fromAddressList[0].(map[string]any)
	if !ok {
		return nil
	}

	var fromAddress platformclientv2.Fromemailaddress

	if domainId, _ := fromAddressMap["domain_id"].(string); domainId != "" {
		fromAddress.Domain = &platformclientv2.Domainentityref{Id: &domainId}
	}
	if friendlyName, _ := fromAddressMap["friendly_name"].(string); friendlyName != "" {
		fromAddress.FriendlyName = &friendlyName
	}
	if localPart, _ := fromAddressMap["local_part"].(string); localPart != "" {
		fromAddress.LocalPart = &localPart
	}

	return &fromAddress
}

func flattenFromAddress(fromAddress *platformclientv2.Fromemailaddress) []any {
	if fromAddress == nil {
		return nil
	}
	fromAddressMap := make(map[string]any)
	resourcedata.SetMapValueIfNotNil(fromAddressMap, "local_part", fromAddress.LocalPart)
	resourcedata.SetMapValueIfNotNil(fromAddressMap, "friendly_name", fromAddress.FriendlyName)
	resourcedata.SetMapReferenceValueIfNotNil(fromAddressMap, "domain_id", fromAddress.Domain)
	return []any{fromAddressMap}
}

func flattenReplyToAddress(replyToAddress *platformclientv2.Replytoemailaddress) []any {
	replyToAddressMap := make(map[string]any)

	resourcedata.SetMapReferenceValueIfNotNil(replyToAddressMap, "domain_id", replyToAddress.Domain)
	resourcedata.SetMapReferenceValueIfNotNil(replyToAddressMap, "route_id", replyToAddress.Route)

	return []any{replyToAddressMap}
}

func buildReplyToAddress(emailConfigMap map[string]any) *platformclientv2.Replytoemailaddress {
	replyToAddressList, ok := emailConfigMap["reply_to_address"].([]any)
	if !ok || len(replyToAddressList) == 0 {
		return nil
	}
	replyToAddressMap, ok := replyToAddressList[0].(map[string]any)
	if !ok {
		return nil
	}

	var replyToAddress platformclientv2.Replytoemailaddress

	if domainId, _ := replyToAddressMap["domain_id"].(string); domainId != "" {
		replyToAddress.Domain = &platformclientv2.Domainentityref{Id: &domainId}
	}
	if routeId, _ := replyToAddressMap["route_id"].(string); routeId != "" {
		replyToAddress.Route = &platformclientv2.Domainentityref{Id: &routeId}
	}

	return &replyToAddress
}

func flattenDynamicContactQueueingSettings(dcqSettings platformclientv2.Dynamiccontactqueueingsettings) []any {
	settingsMap := make(map[string]any)
	resourcedata.SetMapValueIfNotNil(settingsMap, "filter", dcqSettings.Filter)
	resourcedata.SetMapValueIfNotNil(settingsMap, "sort", dcqSettings.Sort)
	return []any{settingsMap}
}

func flattenSdkOutboundMessagingCampaignContactsortSlice(contactSorts []platformclientv2.Contactsort) []interface{} {
	if len(contactSorts) == 0 {
		return nil
	}

	contactSortList := make([]interface{}, 0)
	for _, contactsort := range contactSorts {
		contactSortMap := make(map[string]interface{})

		if contactsort.FieldName != nil {
			contactSortMap["field_name"] = *contactsort.FieldName
		}
		if contactsort.Direction != nil {
			contactSortMap["direction"] = *contactsort.Direction
		}
		if contactsort.Numeric != nil {
			contactSortMap["numeric"] = *contactsort.Numeric
		}

		contactSortList = append(contactSortList, contactSortMap)
	}

	return contactSortList
}

func flattenSmsconfig(smsconfig platformclientv2.Smsconfig) *schema.Set {
	smsconfigSet := schema.NewSet(schema.HashResource(smsConfigResource), []interface{}{})
	smsconfigMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(smsconfigMap, "message_column", smsconfig.MessageColumn)
	resourcedata.SetMapValueIfNotNil(smsconfigMap, "phone_column", smsconfig.PhoneColumn)

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

func validateSmsconfig(smsconfig *schema.Set) (string, bool) {
	if smsconfig == nil {
		return "", true
	}

	smsconfigList := smsconfig.List()
	if len(smsconfigList) > 0 {
		smsconfigMap := smsconfigList[0].(map[string]interface{})
		messageColumn, _ := smsconfigMap["message_column"].(string)
		contentTemplateId, _ := smsconfigMap["content_template_id"].(string)
		if messageColumn == "" && contentTemplateId == "" {
			return "Either message_column or content_template_id is required.", false
		} else if messageColumn != "" && contentTemplateId != "" {
			return "Only one of message_column or content_template_id can be defined", false
		}
	}

	return "", true
}
