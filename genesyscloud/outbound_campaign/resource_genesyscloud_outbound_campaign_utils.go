package outbound_campaign

import (
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

func updateOutboundCampaignStatus(d *schema.ResourceData, outboundApi *platformclientv2.OutboundApi, campaign platformclientv2.Campaign) diag.Diagnostics {
	newCampaignStatus := d.Get("campaign_status").(string)
	if newCampaignStatus != "" &&
		// Campaign status can only go from ON -> OFF or OFF, COMPLETE, INVALID, ETC -> ON
		((*campaign.CampaignStatus == "on" && newCampaignStatus == "off") ||
			(*campaign.CampaignStatus != "on" && newCampaignStatus == "on")) {
		campaign.CampaignStatus = &newCampaignStatus
		log.Printf("Updating Outbound Campaign %s status to %s", *campaign.Name, newCampaignStatus)
		diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			// Get current Outbound Campaign version
			outboundCampaign, resp, getErr := outboundApi.GetOutboundCampaign(d.Id())
			if getErr != nil {
				return resp, diag.Errorf("Failed to read Outbound Campaign %s: %s", d.Id(), getErr)
			}
			campaign.Version = outboundCampaign.Version
			_, _, updateErr := outboundApi.PutOutboundCampaign(d.Id(), campaign)
			if updateErr != nil {
				return resp, diag.Errorf("Failed to update Outbound Campaign %s: %s", *campaign.Name, updateErr)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
	}
	return nil
}

func buildSdkoutboundcampaignPhonecolumnSlice(phonecolumnList []interface{}) *[]platformclientv2.Phonecolumn {
	if phonecolumnList == nil || len(phonecolumnList) == 0 {
		return nil
	}
	sdkPhonecolumnSlice := make([]platformclientv2.Phonecolumn, 0)
	for _, configphonecolumn := range phonecolumnList {
		var sdkPhonecolumn platformclientv2.Phonecolumn
		phonecolumnMap := configphonecolumn.(map[string]interface{})
		if columnName := phonecolumnMap["column_name"].(string); columnName != "" {
			sdkPhonecolumn.ColumnName = &columnName
		}

		sdkPhonecolumnSlice = append(sdkPhonecolumnSlice, sdkPhonecolumn)
	}
	return &sdkPhonecolumnSlice
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
