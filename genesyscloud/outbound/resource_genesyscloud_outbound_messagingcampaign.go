package outbound

import (
	"context"
	"errors"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllOutboundMessagingcampaign(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkMessagingcampaignEntityListing, resp, getErr := outboundApi.GetOutboundMessagingcampaigns(pageSize, pageNum, "", "", "", "", []string{}, "", "", []string{})
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error requesting page of Outbound Messagingcampaign error: %s", getErr), resp)
		}

		if sdkMessagingcampaignEntityListing.Entities == nil || len(*sdkMessagingcampaignEntityListing.Entities) == 0 {
			break
		}

		for _, entity := range *sdkMessagingcampaignEntityListing.Entities {
			resources[*entity.Id] = &resourceExporter.ResourceMeta{BlockLabel: *entity.Name}
		}
	}

	return resources, nil
}

func createOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	alwaysRunning := d.Get("always_running").(bool)
	messagesPerMinute := d.Get("messages_per_minute").(int)
	campaignStatus := d.Get("campaign_status").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkmessagingcampaign := platformclientv2.Messagingcampaign{
		Division:           util.BuildSdkDomainEntityRef(d, "division_id"),
		CallableTimeSet:    util.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		ContactList:        util.BuildSdkDomainEntityRef(d, "contact_list_id"),
		DncLists:           util.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		AlwaysRunning:      &alwaysRunning,
		ContactSorts:       buildSdkoutboundmessagingcampaignContactsortSlice(d.Get("contact_sorts").([]interface{})),
		MessagesPerMinute:  &messagesPerMinute,
		ContactListFilters: util.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		RuleSets:           util.BuildSdkDomainEntityRefArr(d, "rule_set_ids"),
	}

	if smsConfig := buildSmsconfig(d); smsConfig != nil {
		sdkmessagingcampaign.SmsConfig = smsConfig
	}

	if emailConfig := buildEmailConfig(d); emailConfig != nil {
		sdkmessagingcampaign.EmailConfig = emailConfig
	}

	if dcqSettings := buildDynamicContactQueueingSettings(d); dcqSettings != nil {
		sdkmessagingcampaign.DynamicContactQueueingSettings = dcqSettings
	}

	if name != "" {
		sdkmessagingcampaign.Name = &name
	}

	if campaignStatus != "" {
		sdkmessagingcampaign.CampaignStatus = &campaignStatus
	}

	msg, valid := validateSmsconfig(d.Get("sms_config").(*schema.Set))

	if !valid {
		return util.BuildDiagnosticError(ResourceType, "Configuration error", errors.New(msg))
	}

	log.Printf("Creating Outbound Messagingcampaign %s", name)
	outboundMessagingcampaign, resp, err := outboundApi.PostOutboundMessagingcampaigns(sdkmessagingcampaign)
	if err != nil {
		extraDetails, edErr := gatherExtraErrorMessagesFromResponseBody(resp)
		if edErr != nil {
			log.Println(edErr.Error())
		} else {
			extraDetails = fmt.Sprintf("Extra error details: %s", extraDetails)
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create outbound messagingcampaign %s error: %s. %s", name, err, extraDetails), resp)
	}

	if outboundMessagingcampaign.Id == nil {
		msg := "response body from POST /api/v2/outbound/messagingcampaigns did not contain an ID"
		return util.BuildDiagnosticError(ResourceType, msg, errors.New(msg))
	}

	d.SetId(*outboundMessagingcampaign.Id)

	log.Printf("Created Outbound Messagingcampaign %s %s", name, *outboundMessagingcampaign.Id)
	return readOutboundMessagingcampaign(ctx, d, meta)
}

func updateOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	alwaysRunning := d.Get("always_running").(bool)
	messagesPerMinute := d.Get("messages_per_minute").(int)
	campaignStatus := d.Get("campaign_status").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkmessagingcampaign := platformclientv2.Messagingcampaign{
		Division:           util.BuildSdkDomainEntityRef(d, "division_id"),
		CallableTimeSet:    util.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		ContactList:        util.BuildSdkDomainEntityRef(d, "contact_list_id"),
		DncLists:           util.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		AlwaysRunning:      &alwaysRunning,
		ContactSorts:       buildSdkoutboundmessagingcampaignContactsortSlice(d.Get("contact_sorts").([]interface{})),
		MessagesPerMinute:  &messagesPerMinute,
		ContactListFilters: util.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		RuleSets:           util.BuildSdkDomainEntityRefArr(d, "rule_set_ids"),
	}

	if smsConfig := buildSmsconfig(d); smsConfig != nil {
		sdkmessagingcampaign.SmsConfig = smsConfig
	}

	if emailConfig := buildEmailConfig(d); emailConfig != nil {
		sdkmessagingcampaign.EmailConfig = emailConfig
	}

	if name != "" {
		sdkmessagingcampaign.Name = &name
	}

	if campaignStatus != "" {
		sdkmessagingcampaign.CampaignStatus = &campaignStatus
	}

	msg, valid := validateSmsconfig(d.Get("sms_config").(*schema.Set))
	if !valid {
		return util.BuildDiagnosticError(ResourceType, "Configuration error", errors.New(msg))
	}

	log.Printf("Updating Outbound Messaging Campaign %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Messagingcampaign version
		outboundMessagingcampaign, resp, getErr := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s error: %s", name, getErr), resp)
		}
		sdkmessagingcampaign.Version = outboundMessagingcampaign.Version
		_, resp, updateErr := outboundApi.PutOutboundMessagingcampaign(d.Id(), sdkmessagingcampaign)
		if updateErr != nil {
			extraDetails, edErr := gatherExtraErrorMessagesFromResponseBody(resp)
			if edErr != nil {
				log.Println(edErr.Error())
			} else {
				extraDetails = fmt.Sprintf("Extra error details: %s", extraDetails)
			}
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound Messagingcampaign %s error: %s. %s", name, updateErr, extraDetails), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Messagingcampaign %s", name)
	return readOutboundMessagingcampaign(ctx, d, meta)
}

func readOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	//cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundMessagingCampaign(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Messaging Campaign %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkMessagingCampaign, resp, getErr := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", sdkMessagingCampaign.Name)
		resourcedata.SetNillableValue(d, "campaign_status", sdkMessagingCampaign.CampaignStatus)
		resourcedata.SetNillableValue(d, "always_running", sdkMessagingCampaign.AlwaysRunning)
		resourcedata.SetNillableValue(d, "messages_per_minute", sdkMessagingCampaign.MessagesPerMinute)
		resourcedata.SetNillableReference(d, "division_id", sdkMessagingCampaign.Division)
		resourcedata.SetNillableReference(d, "callable_time_set_id", sdkMessagingCampaign.CallableTimeSet)
		resourcedata.SetNillableReference(d, "contact_list_id", sdkMessagingCampaign.ContactList)

		if sdkMessagingCampaign.EmailConfig != nil {
			_ = d.Set("email_config", flattenEmailConfig(*sdkMessagingCampaign.EmailConfig))
		}

		if sdkMessagingCampaign.ContactSorts != nil {
			_ = d.Set("contact_sorts", flattenSdkOutboundMessagingCampaignContactsortSlice(*sdkMessagingCampaign.ContactSorts))
		}

		if sdkMessagingCampaign.SmsConfig != nil {
			_ = d.Set("sms_config", flattenSmsconfig(*sdkMessagingCampaign.SmsConfig))
		}

		if dcqSettings := sdkMessagingCampaign.DynamicContactQueueingSettings; dcqSettings != nil {
			_ = d.Set("dynamic_contact_queueing_settings", flattenDynamicContactQueueingSettings(*dcqSettings))
		}

		if sdkMessagingCampaign.RuleSets != nil {
			_ = d.Set("rule_set_ids", util.SdkDomainEntityRefArrToList(*sdkMessagingCampaign.RuleSets))
		}

		if sdkMessagingCampaign.ContactListFilters != nil {
			_ = d.Set("contact_list_filter_ids", util.SdkDomainEntityRefArrToList(*sdkMessagingCampaign.ContactListFilters))
		}

		if sdkMessagingCampaign.DncLists != nil {
			_ = d.Set("dnc_list_ids", util.SdkDomainEntityRefArrToList(*sdkMessagingCampaign.DncLists))
		}

		log.Printf("Read Outbound Messaging Campaign %s", d.Id())
		//return cc.CheckState(d)
		return nil
	})
}

func deleteOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Messagingcampaign")
		_, resp, err := outboundApi.DeleteOutboundMessagingcampaign(d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete outbound Messagingcampaign %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound Messagingcampaign deleted
				log.Printf("Deleted Outbound Messagingcampaign %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Outbound Messagingcampaign %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Messagingcampaign %s still exists", d.Id()), resp))
	})
}
