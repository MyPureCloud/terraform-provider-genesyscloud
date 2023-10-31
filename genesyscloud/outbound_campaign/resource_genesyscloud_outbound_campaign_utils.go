package outbound_campaign

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
The resource_genesyscloud_outbound_campaign_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

func getOutboundCampaignFromResourceData(d *schema.ResourceData) platformclientv2.Campaign {
	abandonRate := d.Get("abandon_rate").(float64)
	outboundLineCount := d.Get("outbound_line_count").(int)
	skipPreviewDisabled := d.Get("skip_preview_disabled").(bool)
	previewTimeOutSeconds := d.Get("preview_time_out_seconds").(int)
	alwaysRunning := d.Get("always_running").(bool)
	noAnswerTimeout := d.Get("no_answer_timeout").(int)
	callAnalysisLanguage := d.Get("call_analysis_language").(string)
	priority := d.Get("priority").(int)

	campaign := platformclientv2.Campaign{
		Name:                           platformclientv2.String(d.Get("name").(string)),
		DialingMode:                    platformclientv2.String(d.Get("dialing_mode").(string)),
		CallerAddress:                  platformclientv2.String(d.Get("caller_address").(string)),
		CallerName:                     platformclientv2.String(d.Get("caller_name").(string)),
		CampaignStatus:                 platformclientv2.String("off"),
		ContactList:                    gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:                          gcloud.BuildSdkDomainEntityRef(d, "queue_id"),
		Script:                         gcloud.BuildSdkDomainEntityRef(d, "script_id"),
		EdgeGroup:                      gcloud.BuildSdkDomainEntityRef(d, "edge_group_id"),
		Site:                           gcloud.BuildSdkDomainEntityRef(d, "site_id"),
		PhoneColumns:                   buildSdkoutboundcampaignPhonecolumnSlice(d.Get("phone_columns").([]interface{})),
		DncLists:                       gcloud.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		CallableTimeSet:                gcloud.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		CallAnalysisResponseSet:        gcloud.BuildSdkDomainEntityRef(d, "call_analysis_response_set_id"),
		RuleSets:                       gcloud.BuildSdkDomainEntityRefArr(d, "rule_set_ids"),
		SkipPreviewDisabled:            &skipPreviewDisabled,
		AlwaysRunning:                  &alwaysRunning,
		ContactSorts:                   buildSdkoutboundcampaignContactsortSlice(d.Get("contact_sorts").([]interface{})),
		ContactListFilters:             gcloud.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		Division:                       gcloud.BuildSdkDomainEntityRef(d, "division_id"),
		DynamicContactQueueingSettings: buildSdkDynamicContactQueueingSettings(d.Get("dynamic_contact_queueing_settings").([]interface{})),
	}

	if abandonRate != 0 {
		campaign.AbandonRate = &abandonRate
	}
	if outboundLineCount != 0 {
		campaign.OutboundLineCount = &outboundLineCount
	}
	if previewTimeOutSeconds != 0 {
		campaign.PreviewTimeOutSeconds = &previewTimeOutSeconds
	}
	if noAnswerTimeout != 0 {
		campaign.NoAnswerTimeout = &noAnswerTimeout
	}
	if callAnalysisLanguage != "" {
		campaign.CallAnalysisLanguage = &callAnalysisLanguage
	}
	if priority != 0 {
		campaign.Priority = &priority
	}

	return campaign
}

func updateOutboundCampaignStatus(ctx context.Context, d *schema.ResourceData, proxy *outboundCampaignProxy, campaign platformclientv2.Campaign, newCampaignStatus string) diag.Diagnostics {
	if newCampaignStatus != "" &&
		// Campaign status can only go from ON -> OFF or OFF, COMPLETE, INVALID, ETC -> ON
		((*campaign.CampaignStatus == "on" && newCampaignStatus == "off") ||
			(*campaign.CampaignStatus != "on" && newCampaignStatus == "on")) {
		campaign.CampaignStatus = &newCampaignStatus
		log.Printf("Updating Outbound Campaign %s status to %s", *campaign.Name, newCampaignStatus)
		_, err := proxy.updateOutboundCampaign(ctx, d.Id(), &campaign)
		if err != nil {
			return diag.Errorf("Failed to update Outbound Campaign %s: %s", *campaign.Name, err)
		}
	}
	return nil
}

func buildSdkoutboundcampaignPhonecolumnSlice(phonecolumns []interface{}) *[]platformclientv2.Phonecolumn {
	if phonecolumns == nil || len(phonecolumns) == 0 {
		return nil
	}
	phonecolumnSlice := make([]platformclientv2.Phonecolumn, 0)
	for _, phonecolumn := range phonecolumns {
		var sdkPhonecolumn platformclientv2.Phonecolumn

		phonecolumnMap := phonecolumn.(map[string]interface{})
		if columnName := phonecolumnMap["column_name"].(string); columnName != "" {
			sdkPhonecolumn.ColumnName = &columnName
		}

		phonecolumnSlice = append(phonecolumnSlice, sdkPhonecolumn)
	}
	return &phonecolumnSlice
}

func buildSdkDynamicContactQueueingSettings(settings []interface{}) *platformclientv2.Dynamiccontactqueueingsettings {
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
	return &sdkDcqSettings
}

func flattenSdkDynamicContactQueueingSettings(settings platformclientv2.Dynamiccontactqueueingsettings) []interface{} {
	settingsMap := make(map[string]interface{}, 0)
	settingsMap["sort"] = *settings.Sort
	return []interface{}{settingsMap}
}

func buildSdkoutboundcampaignContactsortSlice(contactSortList []interface{}) *[]platformclientv2.Contactsort {
	if contactSortList == nil || len(contactSortList) == 0 {
		return nil
	}
	sdkContactsortSlice := make([]platformclientv2.Contactsort, 0)
	for _, configcontactsort := range contactSortList {
		var sdkContactsort platformclientv2.Contactsort
		contactsortMap := configcontactsort.(map[string]interface{})
		if fieldName := contactsortMap["field_name"].(string); fieldName != "" {
			sdkContactsort.FieldName = &fieldName
		}
		if direction := contactsortMap["direction"].(string); direction != "" {
			sdkContactsort.Direction = &direction
		}
		sdkContactsort.Numeric = platformclientv2.Bool(contactsortMap["numeric"].(bool))
		sdkContactsortSlice = append(sdkContactsortSlice, sdkContactsort)
	}
	return &sdkContactsortSlice
}

func flattenSdkoutboundcampaignPhonecolumnSlice(phonecolumns []platformclientv2.Phonecolumn) []interface{} {
	if len(phonecolumns) == 0 {
		return nil
	}

	phonecolumnList := make([]interface{}, 0)
	for _, phonecolumn := range phonecolumns {
		phonecolumnMap := make(map[string]interface{})

		if phonecolumn.ColumnName != nil {
			phonecolumnMap["column_name"] = *phonecolumn.ColumnName
		}

		phonecolumnList = append(phonecolumnList, phonecolumnMap)
	}

	return phonecolumnList
}

func flattenSdkoutboundcampaignContactsortSlice(contactSorts []platformclientv2.Contactsort) []interface{} {
	if len(contactSorts) == 0 {
		return nil
	}

	contactSortList := make([]interface{}, 0)
	for _, contactSort := range contactSorts {
		contactSortMap := make(map[string]interface{})

		if contactSort.FieldName != nil {
			contactSortMap["field_name"] = *contactSort.FieldName
		}
		if contactSort.Direction != nil {
			contactSortMap["direction"] = *contactSort.Direction
		}
		if contactSort.Numeric != nil {
			contactSortMap["numeric"] = *contactSort.Numeric
		}

		contactSortList = append(contactSortList, contactSortMap)
	}

	return contactSortList
}
