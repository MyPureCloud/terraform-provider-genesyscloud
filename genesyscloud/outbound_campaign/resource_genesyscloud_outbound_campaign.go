package outbound_campaign

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"
)

/*
The resource_genesyscloud_outbound_campaign.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundCampaign retrieves all of the outbound campaign via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundCampaign(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundCampaignProxy(clientConfig)

	campaigns, err := proxy.getAllOutboundCampaign(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get campaigns: %s", err)
	}

	for _, campaign := range *campaigns {
		if *campaign.CampaignStatus != "off" && *campaign.CampaignStatus != "on" {
			*campaign.CampaignStatus = "off"
		}
		resources[*campaign.Id] = &resourceExporter.ResourceMeta{Name: *campaign.Name}
	}

	return resources, nil
}

// createOutboundCampaign is used by the outbound_campaign resource to create Genesys cloud outbound campaign
func createOutboundCampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	dialingMode := d.Get("dialing_mode").(string)
	campaignStatus := d.Get("campaign_status").(string)
	abandonRate := d.Get("abandon_rate").(float64)
	callerName := d.Get("caller_name").(string)
	callerAddress := d.Get("caller_address").(string)
	outboundLineCount := d.Get("outbound_line_count").(int)
	skipPreviewDisabled := d.Get("skip_preview_disabled").(bool)
	previewTimeOutSeconds := d.Get("preview_time_out_seconds").(int)
	alwaysRunning := d.Get("always_running").(bool)
	noAnswerTimeout := d.Get("no_answer_timeout").(int)
	callAnalysisLanguage := d.Get("call_analysis_language").(string)
	priority := d.Get("priority").(int)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcampaign := platformclientv2.Campaign{
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

	if name != "" {
		sdkcampaign.Name = &name
	}
	if dialingMode != "" {
		sdkcampaign.DialingMode = &dialingMode
	}
	if abandonRate != 0 {
		sdkcampaign.AbandonRate = &abandonRate
	}
	if callerName != "" {
		sdkcampaign.CallerName = &callerName
	}
	if callerAddress != "" {
		sdkcampaign.CallerAddress = &callerAddress
	}
	if outboundLineCount != 0 {
		sdkcampaign.OutboundLineCount = &outboundLineCount
	}
	if previewTimeOutSeconds != 0 {
		sdkcampaign.PreviewTimeOutSeconds = &previewTimeOutSeconds
	}
	if noAnswerTimeout != 0 {
		sdkcampaign.NoAnswerTimeout = &noAnswerTimeout
	}
	if callAnalysisLanguage != "" {
		sdkcampaign.CallAnalysisLanguage = &callAnalysisLanguage
	}
	if priority != 0 {
		sdkcampaign.Priority = &priority
	}

	// All campaigns have to be created in an "off" state to start out with
	defaultCampaignStatus := "off"
	sdkcampaign.CampaignStatus = &defaultCampaignStatus

	log.Printf("Creating Outbound Campaign %s", name)
	outboundCampaign, _, err := outboundApi.PostOutboundCampaigns(sdkcampaign)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Campaign %s: %s", name, err)
	}

	d.SetId(*outboundCampaign.Id)

	// Campaigns can be enabled after creation
	if campaignStatus != "" && campaignStatus == "on" {
		d.Set("campaign_status", campaignStatus)
		diag := updateOutboundCampaignStatus(d, outboundApi, *outboundCampaign)
		if diag != nil {
			return diag
		}
	}

	log.Printf("Created Outbound Campaign %s %s", name, *outboundCampaign.Id)

	return readOutboundCampaign(ctx, d, meta)
}

// readOutboundCampaign is used by the outbound_campaign resource to read an outbound campaign from genesys cloud
func readOutboundCampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Campaign %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkcampaign, resp, getErr := outboundApi.GetOutboundCampaign(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Campaign %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Campaign %s: %s", d.Id(), getErr))
		}
		//if *sdkcampaign.CampaignStatus == "stopping" {
		//	return retry.RetryableError(fmt.Errorf("Outbound Campaign still stopping %s", d.Id()))
		//}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCampaign())

		if sdkcampaign.Name != nil {
			d.Set("name", *sdkcampaign.Name)
		}
		if sdkcampaign.ContactList != nil && sdkcampaign.ContactList.Id != nil {
			d.Set("contact_list_id", *sdkcampaign.ContactList.Id)
		}
		if sdkcampaign.Queue != nil && sdkcampaign.Queue.Id != nil {
			d.Set("queue_id", *sdkcampaign.Queue.Id)
		}
		if sdkcampaign.DialingMode != nil {
			d.Set("dialing_mode", *sdkcampaign.DialingMode)
		}
		if sdkcampaign.Script != nil && sdkcampaign.Script.Id != nil {
			d.Set("script_id", *sdkcampaign.Script.Id)
		}
		if sdkcampaign.EdgeGroup != nil && sdkcampaign.EdgeGroup.Id != nil {
			d.Set("edge_group_id", *sdkcampaign.EdgeGroup.Id)
		}
		if sdkcampaign.Site != nil && sdkcampaign.Site.Id != nil {
			d.Set("site_id", *sdkcampaign.Site.Id)
		}
		if sdkcampaign.CampaignStatus != nil {
			d.Set("campaign_status", *sdkcampaign.CampaignStatus)
		}
		if sdkcampaign.PhoneColumns != nil {
			d.Set("phone_columns", flattenSdkoutboundcampaignPhonecolumnSlice(*sdkcampaign.PhoneColumns))
		}
		if sdkcampaign.AbandonRate != nil {
			d.Set("abandon_rate", *sdkcampaign.AbandonRate)
		}
		if sdkcampaign.DncLists != nil {
			d.Set("dnc_list_ids", gcloud.SdkDomainEntityRefArrToList(*sdkcampaign.DncLists))
		}
		if sdkcampaign.CallableTimeSet != nil && sdkcampaign.CallableTimeSet.Id != nil {
			d.Set("callable_time_set_id", *sdkcampaign.CallableTimeSet.Id)
		}
		if sdkcampaign.CallAnalysisResponseSet != nil && sdkcampaign.CallAnalysisResponseSet.Id != nil {
			d.Set("call_analysis_response_set_id", *sdkcampaign.CallAnalysisResponseSet.Id)
		}
		if sdkcampaign.CallerName != nil {
			d.Set("caller_name", *sdkcampaign.CallerName)
		}
		if sdkcampaign.CallerAddress != nil {
			d.Set("caller_address", *sdkcampaign.CallerAddress)
		}
		if sdkcampaign.OutboundLineCount != nil {
			d.Set("outbound_line_count", *sdkcampaign.OutboundLineCount)
		}
		if sdkcampaign.RuleSets != nil {
			d.Set("rule_set_ids", gcloud.SdkDomainEntityRefArrToList(*sdkcampaign.RuleSets))
		}
		if sdkcampaign.SkipPreviewDisabled != nil {
			d.Set("skip_preview_disabled", *sdkcampaign.SkipPreviewDisabled)
		}
		if sdkcampaign.PreviewTimeOutSeconds != nil {
			d.Set("preview_time_out_seconds", *sdkcampaign.PreviewTimeOutSeconds)
		}
		if sdkcampaign.AlwaysRunning != nil {
			d.Set("always_running", *sdkcampaign.AlwaysRunning)
		}
		if sdkcampaign.ContactSorts != nil {
			d.Set("contact_sorts", flattenSdkoutboundcampaignContactsortSlice(*sdkcampaign.ContactSorts))
		}
		if sdkcampaign.NoAnswerTimeout != nil {
			d.Set("no_answer_timeout", *sdkcampaign.NoAnswerTimeout)
		}
		if sdkcampaign.CallAnalysisLanguage != nil {
			d.Set("call_analysis_language", *sdkcampaign.CallAnalysisLanguage)
		}
		if sdkcampaign.Priority != nil {
			d.Set("priority", *sdkcampaign.Priority)
		}
		if sdkcampaign.ContactListFilters != nil {
			d.Set("contact_list_filter_ids", gcloud.SdkDomainEntityRefArrToList(*sdkcampaign.ContactListFilters))
		}
		if sdkcampaign.Division != nil && sdkcampaign.Division.Id != nil {
			d.Set("division_id", *sdkcampaign.Division.Id)
		}
		if sdkcampaign.DynamicContactQueueingSettings != nil {
			d.Set("dynamic_contact_queueing_settings", flattenSdkDynamicContactQueueingSettings(*sdkcampaign.DynamicContactQueueingSettings))
		}

		log.Printf("Read Outbound Campaign %s %s", d.Id(), *sdkcampaign.Name)
		return cc.CheckState()
	})
}

// updateOutboundCampaign is used by the outbound_campaign resource to update an outbound campaign in Genesys Cloud
func updateOutboundCampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	dialingMode := d.Get("dialing_mode").(string)
	abandonRate := d.Get("abandon_rate").(float64)
	callerName := d.Get("caller_name").(string)
	callerAddress := d.Get("caller_address").(string)
	outboundLineCount := d.Get("outbound_line_count").(int)
	skipPreviewDisabled := d.Get("skip_preview_disabled").(bool)
	previewTimeOutSeconds := d.Get("preview_time_out_seconds").(int)
	alwaysRunning := d.Get("always_running").(bool)
	noAnswerTimeout := d.Get("no_answer_timeout").(int)
	callAnalysisLanguage := d.Get("call_analysis_language").(string)
	priority := d.Get("priority").(int)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcampaign := platformclientv2.Campaign{
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

	if name != "" {
		sdkcampaign.Name = &name
	}
	if dialingMode != "" {
		sdkcampaign.DialingMode = &dialingMode
	}
	if abandonRate != 0 {
		sdkcampaign.AbandonRate = &abandonRate
	}
	if callerName != "" {
		sdkcampaign.CallerName = &callerName
	}
	if callerAddress != "" {
		sdkcampaign.CallerAddress = &callerAddress
	}
	if outboundLineCount != 0 {
		sdkcampaign.OutboundLineCount = &outboundLineCount
	}
	if previewTimeOutSeconds != 0 {
		sdkcampaign.PreviewTimeOutSeconds = &previewTimeOutSeconds
	}
	if noAnswerTimeout != 0 {
		sdkcampaign.NoAnswerTimeout = &noAnswerTimeout
	}
	if callAnalysisLanguage != "" {
		sdkcampaign.CallAnalysisLanguage = &callAnalysisLanguage
	}
	if priority != 0 {
		sdkcampaign.Priority = &priority
	}

	log.Printf("Updating Outbound Campaign %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Campaign version
		outboundCampaign, resp, getErr := outboundApi.GetOutboundCampaign(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Campaign %s: %s", d.Id(), getErr)
		}
		sdkcampaign.Version = outboundCampaign.Version

		// Campaign Status has to stay the same, and can only be updated independent of any other operations
		sdkcampaign.CampaignStatus = outboundCampaign.CampaignStatus

		_, _, updateErr := outboundApi.PutOutboundCampaign(d.Id(), sdkcampaign)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Campaign %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	// Check if Campaign Status needs updated
	diagErr = updateOutboundCampaignStatus(d, outboundApi, sdkcampaign)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Campaign %s", name)
	return readOutboundCampaign(ctx, d, meta)
}

// deleteOutboundCampaign is used by the outbound_campaign resource to delete an outbound campaign from Genesys cloud
func deleteOutboundCampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	campaignStatus := d.Get("campaign_status").(string)

	// Campaigns have to be turned off before they can be deleted
	if campaignStatus != "off" {
		diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			log.Printf("Turning off Outbound Campaign before deletion")
			d.Set("campaign_status", "off")
			outboundCampaign, resp, getErr := outboundApi.GetOutboundCampaign(d.Id())
			if getErr != nil {
				return resp, diag.Errorf("Failed to read Outbound Campaign %s: %s", d.Id(), getErr)
			}
			// Handles updating the campaign based on what is set in ResourceData.campaign_status
			diagErr := updateOutboundCampaignStatus(d, outboundApi, *outboundCampaign)
			if diagErr != nil {
				return resp, diagErr
			}
			return resp, nil
		})
		if diagErr != nil {
			return diagErr
		}
	}
	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Campaign")
		_, resp, err := outboundApi.DeleteOutboundCampaign(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Campaign: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundCampaign(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Campaign deleted
				log.Printf("Deleted Outbound Campaign %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Outbound Campaign %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Campaign %s still exists", d.Id()))
	})
}
