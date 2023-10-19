package outbound

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	outboundcampaignphonecolumnResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the phone column.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

func ResourceOutboundCampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound campaign`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundCampaign),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundCampaign),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundCampaign),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundCampaign),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Campaign.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_list_id`: {
				Description: `The ContactList for this Campaign to dial.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`queue_id`: {
				Description: `The Queue for this Campaign to route calls to. Required for all dialing modes except agentless.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`dialing_mode`: {
				Description:  `The strategy this Campaign will use for dialing.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`agentless`, `preview`, `power`, `predictive`, `progressive`, `external`}, false),
			},
			`script_id`: {
				Description: `The Script to be displayed to agents that are handling outbound calls. Required for all dialing modes except agentless.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`edge_group_id`: {
				Description: `The EdgeGroup that will place the calls. Required for all dialing modes except preview.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`site_id`: {
				Description: `The identifier of the site to be used for dialing; can be set in place of an edge group.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`campaign_status`: {
				Description:  `The current status of the Campaign. A Campaign may be turned 'on' or 'off' (default). If this value is changed alongside other changes to the resource, a subsequent update will occur immediately afterwards to set the campaign status. This is due to behavioral requirements in the Genesys Cloud API.`,
				Optional:     true,
				Type:         schema.TypeString,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{`on`, `off`}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return (old == `complete` && new == `off`) || (old == `invalid` && new == `off`) || (old == `stopping` && new == `off` || old == `complete` && new == `on`)
				},
			},
			`phone_columns`: {
				Description: `The ContactPhoneNumberColumns on the ContactList that this Campaign should dial.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        outboundcampaignphonecolumnResource,
			},
			`abandon_rate`: {
				Description: `The targeted abandon rate percentage. Required for progressive, power, and predictive campaigns.`,
				Optional:    true,
				Type:        schema.TypeFloat,
			},
			`dnc_list_ids`: {
				Description: `DncLists for this Campaign to check before placing a call.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`callable_time_set_id`: {
				Description: `The callable time set for this campaign to check before placing a call.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`call_analysis_response_set_id`: {
				Description: `The call analysis response set to handle call analysis results from the edge. Required for all dialing modes except preview.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`caller_name`: {
				Description: `The caller id name to be displayed on the outbound call.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`caller_address`: {
				Description: `The caller id phone number to be displayed on the outbound call.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`outbound_line_count`: {
				Description: `The number of outbound lines to be concurrently dialed. Only applicable to non-preview campaigns; only required for agentless.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`rule_set_ids`: {
				Description: `Rule sets to be applied while this campaign is dialing.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`skip_preview_disabled`: {
				Description: `Whether or not agents can skip previews without placing a call. Only applicable for preview campaigns.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`preview_time_out_seconds`: {
				Description: `The number of seconds before a call will be automatically placed on a preview. A value of 0 indicates no automatic placement of calls. Only applicable to preview campaigns.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`always_running`: {
				Description: `Indicates (when true) that the campaign will remain on after contacts are depleted, allowing additional contacts to be appended/added to the contact list and processed by the still-running campaign. The campaign can still be turned off manually.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`contact_sorts`: {
				Description: `The order in which to sort contacts for dialing, based on up to four columns.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundmessagingcampaigncontactsortResource,
			},
			`no_answer_timeout`: {
				Description: `How long to wait before dispositioning a call as 'no-answer'. Default 30 seconds. Only applicable to non-preview campaigns.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`call_analysis_language`: {
				Description: `The language the edge will use to analyze the call.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`priority`: {
				Description: `The priority of this campaign relative to other campaigns that are running on the same queue. 5 is the highest priority, 1 the lowest.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`contact_list_filter_ids`: {
				Description: `Filter to apply to the contact list before dialing. Currently a campaign can only have one filter applied.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`division_id`: {
				Description: `The division this campaign belongs to.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`dynamic_contact_queueing_settings`: {
				Description: `Settings for dynamic queueing of contacts.`,
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sort": {
							Description: "Whether to sort contacts dynamically.",
							Type:        schema.TypeBool,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func getAllOutboundCampaign(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkcampaignentitylisting, _, getErr := outboundApi.GetOutboundCampaigns(pageSize, pageNum, "", "", []string{}, "", "", "", "", "", []string{}, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Outbound Campaign: %s", getErr)
		}

		if sdkcampaignentitylisting.Entities == nil || len(*sdkcampaignentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdkcampaignentitylisting.Entities {
			if *entity.CampaignStatus != "off" && *entity.CampaignStatus != "on" {
				*entity.CampaignStatus = "off"
			}
			resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func OutboundCampaignExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundCampaign),
		AllowZeroValues:  []string{`preview_time_out_seconds`},
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`contact_list_id`: {
				RefType: "genesyscloud_outbound_contact_list",
			},
			`queue_id`: {
				RefType: "genesyscloud_routing_queue",
			},
			`edge_group_id`: {
				RefType: "genesyscloud_telephony_providers_edges_edge_group",
			},
			`site_id`: {
				RefType: "genesyscloud_telephony_providers_edges_site",
			},
			`dnc_list_ids`: {
				RefType: "genesyscloud_outbound_dnclist",
			},
			`call_analysis_response_set_id`: {
				RefType: "genesyscloud_outbound_callanalysisresponseset",
			},
			`contact_list_filter_ids`: {
				RefType: "genesyscloud_outbound_contactlistfilter",
			},
			`division_id`: {
				RefType: "genesyscloud_auth_division",
			},
			`rule_set_ids`: {
				RefType: "genesyscloud_outbound_ruleset",
			},
			`callable_time_set_id`: {
				RefType: "genesyscloud_outbound_callabletimeset",
			},
			`script_id`: {
				RefType: "genesyscloud_script",
			},
		},
	}
}

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
		if *sdkcampaign.CampaignStatus == "stopping" {
			return retry.RetryableError(fmt.Errorf("Outbound Campaign still stopping %s", d.Id()))
		}

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
