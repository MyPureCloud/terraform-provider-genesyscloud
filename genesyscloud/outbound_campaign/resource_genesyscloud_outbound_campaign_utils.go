package outbound_campaign

import (
	"context"
	"fmt"
	"log"
	"strconv"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	obResponseSet "terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)

	campaign := platformclientv2.Campaign{
		Name:                           platformclientv2.String(d.Get("name").(string)),
		DialingMode:                    platformclientv2.String(d.Get("dialing_mode").(string)),
		CallerAddress:                  platformclientv2.String(d.Get("caller_address").(string)),
		CallerName:                     platformclientv2.String(d.Get("caller_name").(string)),
		CampaignStatus:                 platformclientv2.String("off"),
		ContactList:                    util.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:                          util.BuildSdkDomainEntityRef(d, "queue_id"),
		Script:                         util.BuildSdkDomainEntityRef(d, "script_id"),
		EdgeGroup:                      util.BuildSdkDomainEntityRef(d, "edge_group_id"),
		Site:                           util.BuildSdkDomainEntityRef(d, "site_id"),
		PhoneColumns:                   buildPhoneColumns(d.Get("phone_columns").([]interface{})),
		DncLists:                       util.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		CallableTimeSet:                util.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		CallAnalysisResponseSet:        util.BuildSdkDomainEntityRef(d, "call_analysis_response_set_id"),
		RuleSets:                       util.BuildSdkDomainEntityRefArr(d, "rule_set_ids"),
		SkipPreviewDisabled:            &skipPreviewDisabled,
		AlwaysRunning:                  &alwaysRunning,
		ContactSorts:                   buildContactSorts(d.Get("contact_sorts").([]interface{})),
		ContactListFilters:             util.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		Division:                       util.BuildSdkDomainEntityRef(d, "division_id"),
		DynamicContactQueueingSettings: buildSettings(d.Get("dynamic_contact_queueing_settings").([]interface{})),
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
	if maxCallsPerAgent != 0 {
		campaign.MaxCallsPerAgent = &maxCallsPerAgent
	}
	return campaign
}

func updateOutboundCampaignStatus(ctx context.Context, campaignId string, proxy *outboundCampaignProxy, campaign platformclientv2.Campaign, newCampaignStatus string) diag.Diagnostics {
	if newCampaignStatus == "" {
		return nil
	}
	// Campaign status can only go from ON -> OFF or OFF, COMPLETE, INVALID, ETC -> ON
	if (*campaign.CampaignStatus == "on" && newCampaignStatus == "off") || newCampaignStatus == "on" {
		campaign.CampaignStatus = &newCampaignStatus
		log.Printf("Updating Outbound Campaign %s status to %s", *campaign.Name, newCampaignStatus)
		_, resp, err := proxy.updateOutboundCampaign(ctx, campaignId, &campaign)
		if err != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Outbound Campaign %s error: %s", *campaign.Name, err), resp)
		}
	}
	return nil
}

func buildPhoneColumns(phonecolumns []interface{}) *[]platformclientv2.Phonecolumn {
	if len(phonecolumns) == 0 {
		return nil
	}
	phonecolumnSlice := make([]platformclientv2.Phonecolumn, 0)
	for _, phonecolumn := range phonecolumns {
		var sdkPhonecolumn platformclientv2.Phonecolumn
		phonecolumnMap, ok := phonecolumn.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkPhonecolumn.ColumnName, phonecolumnMap, "column_name")
		phonecolumnSlice = append(phonecolumnSlice, sdkPhonecolumn)
	}
	return &phonecolumnSlice
}

func buildSettings(settings []interface{}) *platformclientv2.Dynamiccontactqueueingsettings {
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

func buildContactSorts(contactSortList []interface{}) *[]platformclientv2.Contactsort {
	if len(contactSortList) == 0 {
		return nil
	}
	sdkContactsortSlice := make([]platformclientv2.Contactsort, 0)
	for _, configcontactsort := range contactSortList {
		var sdkContactsort platformclientv2.Contactsort
		contactsortMap := configcontactsort.(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkContactsort.FieldName, contactsortMap, "field_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContactsort.Direction, contactsortMap, "direction")

		sdkContactsort.Numeric = platformclientv2.Bool(contactsortMap["numeric"].(bool))
		sdkContactsortSlice = append(sdkContactsortSlice, sdkContactsort)
	}
	return &sdkContactsortSlice
}

func flattenSettings(settings *platformclientv2.Dynamiccontactqueueingsettings) []interface{} {
	settingsMap := make(map[string]interface{}, 0)
	settingsMap["sort"] = *settings.Sort
	return []interface{}{settingsMap}
}

func flattenPhoneColumn(phonecolumns *[]platformclientv2.Phonecolumn) []interface{} {
	if len(*phonecolumns) == 0 {
		return nil
	}

	phonecolumnList := make([]interface{}, 0)
	for _, phonecolumn := range *phonecolumns {
		phonecolumnMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(phonecolumnMap, "column_name", phonecolumn.ColumnName)
		phonecolumnList = append(phonecolumnList, phonecolumnMap)
	}

	return phonecolumnList
}

func flattenContactSorts(contactSorts *[]platformclientv2.Contactsort) []interface{} {
	if len(*contactSorts) == 0 {
		return nil
	}

	contactSortList := make([]interface{}, 0)
	for _, contactSort := range *contactSorts {
		contactSortMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactSortMap, "field_name", contactSort.FieldName)
		resourcedata.SetMapValueIfNotNil(contactSortMap, "direction", contactSort.Direction)
		resourcedata.SetMapValueIfNotNil(contactSortMap, "numeric", contactSort.Numeric)

		contactSortList = append(contactSortList, contactSortMap)
	}

	return contactSortList
}

func GenerateOutboundCampaignBasic(resourceId string,
	name string,
	contactListResourceId string,
	siteResourceId string,
	emergencyNumber string,
	carResourceId string,
	campaignStatus string,
	outboundFlowFilePath string,
	flowResourceId string,
	flowName string,
	divisionName,
	locationResourceId string,
	wrapupcodeResourceId string) string {
	referencedResources := GenerateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		"",
		"",
		carResourceId,
		outboundFlowFilePath,
		flowResourceId,
		flowName,
		"",
		siteResourceId,
		emergencyNumber,
		"",
		"",
		divisionName,
		locationResourceId,
		wrapupcodeResourceId,
	)
	return fmt.Sprintf(`
resource "genesyscloud_outbound_campaign" "%s" {
	name                          = "%s"
	dialing_mode                  = "agentless"
	caller_name                   = "Test Name"
	caller_address                = "+353371111111"
	outbound_line_count           = 2
	campaign_status               = %s
	contact_list_id 			  = genesyscloud_outbound_contact_list.%s.id
	site_id 				      = genesyscloud_telephony_providers_edges_site.%s.id
	call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.%s.id
	phone_columns {
		column_name = "Cell"
	}
}
%s
`, resourceId, name, campaignStatus, contactListResourceId, siteResourceId, carResourceId, referencedResources)
}

func GenerateReferencedResourcesForOutboundCampaignTests(
	contactListResourceId string,
	dncListResourceId string,
	queueResourceId string,
	carResourceId string,
	outboundFlowFilePath string,
	flowResourceId string,
	flowName string,
	clfResourceId string,
	siteId string,
	emergencyNumber string,
	ruleSetId string,
	callableTimeSetResourceId string,
	divisionName string,
	locationResourceId string,
	wrapUpCodeResourceId string,
) string {
	var (
		contactList             string
		dncList                 string
		queue                   string
		callAnalysisResponseSet string
		contactListFilter       string
		site                    string
		ruleSet                 string
		callableTimeSet         string
	)
	if contactListResourceId != "" {
		contactList = obContactList.GenerateOutboundContactList(
			contactListResourceId,
			"terraform contact list "+uuid.NewString(),
			util.NullValue,
			strconv.Quote("Cell"),
			[]string{strconv.Quote("Cell")},
			[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
			util.FalseValue,
			util.NullValue,
			util.NullValue,
			obContactList.GeneratePhoneColumnsBlock("Cell", "cell", strconv.Quote("Cell")),
			obContactList.GeneratePhoneColumnsBlock("Home", "home", strconv.Quote("Home")))
	}
	if dncListResourceId != "" {
		dncList = obDnclist.GenerateOutboundDncListBasic(dncListResourceId, "tf dnc list "+uuid.NewString())
	}
	if queueResourceId != "" {
		queue = routingQueue.GenerateRoutingQueueResourceBasic(queueResourceId, "tf test queue "+uuid.NewString())
	}
	if carResourceId != "" {
		if outboundFlowFilePath != "" {
			callAnalysisResponseSet = gcloud.GenerateRoutingWrapupcodeResource(
				wrapUpCodeResourceId,
				"wrapupcode "+uuid.NewString(),
			) + architect_flow.GenerateFlowResource(
				flowResourceId,
				outboundFlowFilePath,
				"",
				false,
				util.GenerateSubstitutionsMap(map[string]string{
					"flow_name":          flowName,
					"home_division_name": divisionName,
					"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceId + ".name}",
					"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapUpCodeResourceId + ".name}",
				}),
			) + obResponseSet.GenerateOutboundCallAnalysisResponseSetResource(
				carResourceId,
				"tf test car "+uuid.NewString(),
				util.FalseValue,
				obResponseSet.GenerateCarsResponsesBlock(
					obResponseSet.GenerateCarsResponse(
						"callable_person",
						"transfer_flow",
						flowName,
						"${genesyscloud_flow."+flowResourceId+".id}",
					),
				))
		} else {
			callAnalysisResponseSet = obResponseSet.GenerateOutboundCallAnalysisResponseSetResource(
				carResourceId,
				"tf test car "+uuid.NewString(),
				util.TrueValue,
				obResponseSet.GenerateCarsResponsesBlock(
					obResponseSet.GenerateCarsResponse(
						"callable_machine",
						"transfer",
						"",
						"",
					),
				),
			)
		}
	}
	if clfResourceId != "" {
		contactListFilter = obContactListFilter.GenerateOutboundContactListFilter(
			clfResourceId,
			"tf test clf "+uuid.NewString(),
			"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
			"",
			obContactListFilter.GenerateOutboundContactListFilterClause(
				"",
				obContactListFilter.GenerateOutboundContactListFilterPredicates(
					"Cell",
					"alphabetic",
					"EQUALS",
					"+12345123456",
					"",
					"",
				),
			),
		)
	}
	if siteId != "" {
		siteName := "site " + uuid.NewString()
		locationName := "location " + uuid.NewString()
		site = fmt.Sprintf(`
resource "genesyscloud_location" "%s" {
	name  = "%s"
	notes = "HQ1"
	path  = []
	emergency_number {
		number = "%s"
		type   = null
	}
	address {
		street1  = "7601 Interactive Way"
		city     = "Indianapolis"
		state    = "IN"
		country  = "US"
		zip_code = "46278"
	}
}

resource "genesyscloud_telephony_providers_edges_site" "%s" {
	name                            = "%s"
	description                     = "TestAccResourceSite description 1"
	location_id                     = genesyscloud_location.%s.id
	media_model                     = "Cloud"
	media_regions_use_latency_based = false
}
`, locationResourceId, locationName, emergencyNumber, siteId, siteName, locationResourceId)
	}
	if ruleSetId != "" {
		ruleSetName := "ruleset " + uuid.NewString()
		ruleSet = fmt.Sprintf(`
resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
}
`, ruleSetId, ruleSetName, contactListResourceId)
	}
	if callableTimeSetResourceId != "" {
		callableTimeSetName := "test time set " + uuid.NewString()
		callableTimeSet = fmt.Sprintf(`
resource "genesyscloud_outbound_callabletimeset" "%s"{
	name = "%s"
	callable_times {
		time_zone_id = "Africa/Abidjan"
		time_slots {
			start_time = "07:00:00"
			stop_time = "18:00:00"
			day = 3
		}
	}
}
`, callableTimeSetResourceId, callableTimeSetName)
	}
	return fmt.Sprintf(`
%s
%s
%s
%s
%s
%s
%s
%s
`, contactList, dncList, queue, callAnalysisResponseSet, contactListFilter, site, ruleSet, callableTimeSet)
}
