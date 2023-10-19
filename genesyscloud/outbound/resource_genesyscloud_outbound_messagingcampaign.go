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
	outboundmessagingcampaigncontactsortResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`field_name`: {
				Description: `The field name by which to sort contacts.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`direction`: {
				Description:  `The direction in which to sort contacts.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`ASC`, `DESC`}, false),
				Default:      `ASC`,
			},
			`numeric`: {
				Description: `Whether or not the column contains numeric data.`,
				Optional:    true,
				Type:        schema.TypeBool,
				Default:     false,
			},
		},
	}
	outboundmessagingcampaignsmsconfigResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`message_column`: {
				Description: `The Contact List column specifying the message to send to the contact.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`phone_column`: {
				Description: `The Contact List column specifying the phone number to send a message to.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`sender_sms_phone_number`: {
				Description: `A phone number provisioned for SMS communications in E.164 format. E.g. +13175555555 or +34234234234`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`content_template_id`: {
				Description: `The content template used to formulate the message to send to the contact.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

func ResourceOutboundMessagingCampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Messaging Campaign`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundMessagingcampaign),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundMessagingcampaign),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundMessagingcampaign),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundMessagingcampaign),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The campaign name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division this entity belongs to.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`callable_time_set_id`: {
				Description: `The callable time set for this messaging campaign.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`contact_list_id`: {
				Description: `The contact list that this messaging campaign will send messages for.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`dnc_list_ids`: {
				Description: `The dnc lists to check before sending a message for this messaging campaign.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`campaign_status`: {
				Description:  `The current status of the messaging campaign. A messaging campaign may be turned 'on' or 'off'.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`on`, `off`}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == `complete` && new == `on` {
						return true
					}
					if old == `invalid` && new == `on` {
						return true
					}
					if old == `stopping` && new == `off` {
						return true
					}
					return false
				},
			},
			`always_running`: {
				Description: `Whether this messaging campaign is always running`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			`contact_sorts`: {
				Description: `The order in which to sort contacts for dialing, based on up to four columns.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundmessagingcampaigncontactsortResource,
			},
			`messages_per_minute`: {
				Description: `How many messages this messaging campaign will send per minute.`,
				Required:    true,
				Type:        schema.TypeInt,
			},
			`contact_list_filter_ids`: {
				Description: `The contact list filter to check before sending a message for this messaging campaign.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`sms_config`: {
				Description: `Configuration for this messaging campaign to send SMS messages.`,
				Required:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        outboundmessagingcampaignsmsconfigResource,
			},
		},
	}
}

func getAllOutboundMessagingcampaign(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkMessagingcampaignEntityListing, _, getErr := outboundApi.GetOutboundMessagingcampaigns(pageSize, pageNum, "", "", "", "", []string{}, "", "", []string{})
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Outbound Messagingcampaign: %s", getErr)
		}

		if sdkMessagingcampaignEntityListing.Entities == nil || len(*sdkMessagingcampaignEntityListing.Entities) == 0 {
			break
		}

		for _, entity := range *sdkMessagingcampaignEntityListing.Entities {
			resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func OutboundMessagingcampaignExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundMessagingcampaign),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`division_id`:             {RefType: "genesyscloud_auth_division"},
			`contact_list_id`:         {RefType: "genesyscloud_outbound_contact_list"},
			`contact_list_filter_ids`: {RefType: "genesyscloud_outbound_contactlistfilter"},
			`dnc_list_ids`:            {RefType: "genesyscloud_outbound_dnclist"},
			`callable_time_set_id`:    {RefType: "genesyscloud_outbound_callabletimeset"},
			// /api/v2/responsemanagement/responses/{responseId}
			`sms_config.content_template_id`: {},
		},
	}
}

func createOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	alwaysRunning := d.Get("always_running").(bool)
	messagesPerMinute := d.Get("messages_per_minute").(int)
	campaignStatus := d.Get("campaign_status").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkmessagingcampaign := platformclientv2.Messagingcampaign{
		Division:           gcloud.BuildSdkDomainEntityRef(d, "division_id"),
		CallableTimeSet:    gcloud.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		ContactList:        gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		DncLists:           gcloud.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		AlwaysRunning:      &alwaysRunning,
		ContactSorts:       buildSdkoutboundmessagingcampaignContactsortSlice(d.Get("contact_sorts").([]interface{})),
		MessagesPerMinute:  &messagesPerMinute,
		ContactListFilters: gcloud.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		SmsConfig:          buildSdkoutboundmessagingcampaignSmsconfig(d.Get("sms_config").(*schema.Set)),
	}

	if name != "" {
		sdkmessagingcampaign.Name = &name
	}

	if campaignStatus != "" {
		sdkmessagingcampaign.CampaignStatus = &campaignStatus
	}

	log.Printf("Creating Outbound Messagingcampaign %s", name)
	outboundMessagingcampaign, _, err := outboundApi.PostOutboundMessagingcampaigns(sdkmessagingcampaign)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Messagingcampaign %s: %s", name, err)
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

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkmessagingcampaign := platformclientv2.Messagingcampaign{
		Division:           gcloud.BuildSdkDomainEntityRef(d, "division_id"),
		CallableTimeSet:    gcloud.BuildSdkDomainEntityRef(d, "callable_time_set_id"),
		ContactList:        gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		DncLists:           gcloud.BuildSdkDomainEntityRefArr(d, "dnc_list_ids"),
		AlwaysRunning:      &alwaysRunning,
		ContactSorts:       buildSdkoutboundmessagingcampaignContactsortSlice(d.Get("contact_sorts").([]interface{})),
		MessagesPerMinute:  &messagesPerMinute,
		ContactListFilters: gcloud.BuildSdkDomainEntityRefArr(d, "contact_list_filter_ids"),
		SmsConfig:          buildSdkoutboundmessagingcampaignSmsconfig(d.Get("sms_config").(*schema.Set)),
	}

	if name != "" {
		sdkmessagingcampaign.Name = &name
	}

	if campaignStatus != "" {
		sdkmessagingcampaign.CampaignStatus = &campaignStatus
	}

	log.Printf("Updating Outbound Messagingcampaign %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Messagingcampaign version
		outboundMessagingcampaign, resp, getErr := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Messagingcampaign %s: %s", d.Id(), getErr)
		}
		sdkmessagingcampaign.Version = outboundMessagingcampaign.Version
		outboundMessagingcampaign, _, updateErr := outboundApi.PutOutboundMessagingcampaign(d.Id(), sdkmessagingcampaign)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Messagingcampaign %s: %s", name, updateErr)
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Messagingcampaign %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkmessagingcampaign, resp, getErr := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Messagingcampaign %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Messagingcampaign %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundMessagingCampaign())

		if sdkmessagingcampaign.Name != nil {
			d.Set("name", *sdkmessagingcampaign.Name)
		}
		if sdkmessagingcampaign.Division != nil && sdkmessagingcampaign.Division.Id != nil {
			d.Set("division_id", *sdkmessagingcampaign.Division.Id)
		}
		if sdkmessagingcampaign.CallableTimeSet != nil && sdkmessagingcampaign.CallableTimeSet.Id != nil {
			d.Set("callable_time_set_id", *sdkmessagingcampaign.CallableTimeSet.Id)
		}
		if sdkmessagingcampaign.ContactList != nil && sdkmessagingcampaign.ContactList.Id != nil {
			d.Set("contact_list_id", *sdkmessagingcampaign.ContactList.Id)
		}
		if sdkmessagingcampaign.CampaignStatus != nil {
			d.Set("campaign_status", *sdkmessagingcampaign.CampaignStatus)
		}
		if sdkmessagingcampaign.DncLists != nil {
			var dncListIds []string
			for _, dnc := range *sdkmessagingcampaign.DncLists {
				dncListIds = append(dncListIds, *dnc.Id)
			}
			d.Set("dnc_list_ids", dncListIds)
		}
		if sdkmessagingcampaign.AlwaysRunning != nil {
			d.Set("always_running", *sdkmessagingcampaign.AlwaysRunning)
		}
		if sdkmessagingcampaign.ContactSorts != nil {
			d.Set("contact_sorts", flattenSdkOutboundMessagingCampaignContactsortSlice(*sdkmessagingcampaign.ContactSorts))
		}
		if sdkmessagingcampaign.MessagesPerMinute != nil {
			d.Set("messages_per_minute", *sdkmessagingcampaign.MessagesPerMinute)
		}
		if sdkmessagingcampaign.ContactListFilters != nil {
			var contactListFilterIds []string
			for _, clf := range *sdkmessagingcampaign.ContactListFilters {
				contactListFilterIds = append(contactListFilterIds, *clf.Id)
			}
			d.Set("contact_list_filter_ids", contactListFilterIds)
		}
		if sdkmessagingcampaign.SmsConfig != nil {
			d.Set("sms_config", flattenSdkOutboundMessagingCampaignSmsconfig(sdkmessagingcampaign.SmsConfig))
		}

		log.Printf("Read Outbound Messagingcampaign %s %s", d.Id(), *sdkmessagingcampaign.Name)
		return cc.CheckState()
	})
}

func deleteOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Messagingcampaign")
		_, resp, err := outboundApi.DeleteOutboundMessagingcampaign(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Messagingcampaign: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Messagingcampaign deleted
				log.Printf("Deleted Outbound Messagingcampaign %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Outbound Messagingcampaign %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Messagingcampaign %s still exists", d.Id()))
	})
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

func buildSdkoutboundmessagingcampaignSmsconfig(smsconfig *schema.Set) *platformclientv2.Smsconfig {
	if smsconfig == nil {
		return nil
	}
	var sdkSmsconfig platformclientv2.Smsconfig
	smsconfigList := smsconfig.List()
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
	}

	return &sdkSmsconfig
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

func flattenSdkOutboundMessagingCampaignSmsconfig(smsconfig *platformclientv2.Smsconfig) *schema.Set {
	if smsconfig == nil {
		return nil
	}

	smsconfigSet := schema.NewSet(schema.HashResource(outboundmessagingcampaignsmsconfigResource), []interface{}{})
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
