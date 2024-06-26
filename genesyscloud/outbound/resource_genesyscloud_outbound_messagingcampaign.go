package outbound

import (
	"context"
	"errors"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

const (
	resourceName = "genesyscloud_outbound_messagingcampaign"
)

var (
	OutboundmessagingcampaigncontactsortResource = &schema.Resource{
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
				Description: `The Contact List column specifying the message to send to the contact. Either message_column or content_template_id is required.`,
				Optional:    true,
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
				Description: `The content template used to formulate the message to send to the contact. Either message_column or content_template_id is required.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

func ResourceOutboundMessagingCampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Messaging Campaign`,

		CreateContext: provider.CreateWithPooledClient(createOutboundMessagingcampaign),
		ReadContext:   provider.ReadWithPooledClient(readOutboundMessagingcampaign),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundMessagingcampaign),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundMessagingcampaign),
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
				Elem:        OutboundmessagingcampaigncontactsortResource,
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
		sdkMessagingcampaignEntityListing, resp, getErr := outboundApi.GetOutboundMessagingcampaigns(pageSize, pageNum, "", "", "", "", []string{}, "", "", []string{})
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Error requesting page of Outbound Messagingcampaign error: %s", getErr), resp)
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
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOutboundMessagingcampaign),
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
		SmsConfig:          buildSdkoutboundmessagingcampaignSmsconfig(d.Get("sms_config").(*schema.Set)),
	}

	if name != "" {
		sdkmessagingcampaign.Name = &name
	}

	if campaignStatus != "" {
		sdkmessagingcampaign.CampaignStatus = &campaignStatus
	}

	msg, valid := validateSmsconfig(d.Get("sms_config").(*schema.Set))

	if !valid {
		return util.BuildDiagnosticError(resourceName, "Configuration error", errors.New(msg))
	}

	log.Printf("Creating Outbound Messagingcampaign %s", name)
	outboundMessagingcampaign, resp, err := outboundApi.PostOutboundMessagingcampaigns(sdkmessagingcampaign)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create outbound messagingcampaign %s error: %s", name, err), resp)
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
		SmsConfig:          buildSdkoutboundmessagingcampaignSmsconfig(d.Get("sms_config").(*schema.Set)),
	}

	if name != "" {
		sdkmessagingcampaign.Name = &name
	}

	if campaignStatus != "" {
		sdkmessagingcampaign.CampaignStatus = &campaignStatus
	}

	msg, valid := validateSmsconfig(d.Get("sms_config").(*schema.Set))

	if !valid {
		return util.BuildDiagnosticError(resourceName, "Configuration error", errors.New(msg))
	}

	log.Printf("Updating Outbound Messagingcampaign %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Messagingcampaign version
		outboundMessagingcampaign, resp, getErr := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s error: %s", name, getErr), resp)
		}
		sdkmessagingcampaign.Version = outboundMessagingcampaign.Version
		outboundMessagingcampaign, resp, updateErr := outboundApi.PutOutboundMessagingcampaign(d.Id(), sdkmessagingcampaign)
		if updateErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Outbound Messagingcampaign %s error: %s", name, updateErr), resp)
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
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundMessagingCampaign(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading Outbound Messagingcampaign %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkmessagingcampaign, resp, getErr := outboundApi.GetOutboundMessagingcampaign(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s | error: %s", d.Id(), getErr), resp))
		}

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
		return cc.CheckState(d)
	})
}

func deleteOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Messagingcampaign")
		_, resp, err := outboundApi.DeleteOutboundMessagingcampaign(d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete outbound Messagingcampaign %s error: %s", d.Id(), err), resp)
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
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting Outbound Messagingcampaign %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Outbound Messagingcampaign %s still exists", d.Id()), resp))
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
