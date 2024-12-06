package outbound

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_outbound_messagingcampaign"

var (
	dynamicContactQueueingSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`sort`: {
				Description: `Whether to sort contacts dynamically.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeBool,
			},
			`filter`: {
				Description: `Whether to filter contacts dynamically.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeBool,
			},
		},
	}

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

	smsConfigResource = &schema.Resource{
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

	emailConfigResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`email_columns`: {
				Optional:    true,
				Description: `The contact list columns specifying the email address(es) of the contact.`,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`content_template_id`: {
				Optional:    true,
				Description: `The content template used to formulate the email to send to the contact.`,
				Type:        schema.TypeString,
			},
			`from_address`: {
				Optional:    true,
				Description: `The email address that will be used as the sender of the email.`,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`domain_id`: {
							Description: `The OutboundDomain used for the email address.`,
							Type:        schema.TypeString,
							Optional:    true,
						},
						`friendly_name`: {
							Type:        schema.TypeString,
							Description: `The friendly name of the email address.`,
							Optional:    true,
						},
						`local_part`: {
							Type:        schema.TypeString,
							Description: `The local part of the email address.`,
							Optional:    true,
						},
					},
				},
			},
			`reply_to_address`: {
				Optional:    true,
				Description: `The email address from which any reply will be sent.`,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`domain_id`: {
							Description: `The InboundDomain used for the email address.`,
							Type:        schema.TypeString,
							Optional:    true,
						},
						`route_id`: {
							Type:        schema.TypeString,
							Description: `The InboundRoute used for the email address.`,
							Optional:    true,
						},
					},
				},
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
			`email_config`: {
				Description: `Configuration for this messaging campaign to send Email messages.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        emailConfigResource,
			},
			`sms_config`: {
				Description: `Configuration for this messaging campaign to send SMS messages.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        smsConfigResource,
			},
			`dynamic_contact_queueing_settings`: {
				Description: `Indicates (when true) that the campaign supports dynamic queueing of the contact list at the time of a request for contacts.
				**Warning**: Updating this field will cause the campaign to be destroyed and re-created.`,
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem:     dynamicContactQueueingSettings,
			},
			`rule_set_ids`: {
				Description: `Rule Sets to be applied while this campaign is sending messages`,
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceOutboundMessagingcampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Outbound Messaging Campaign. Select a Outbound Messaging Campaign by name.`,

		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundMessagingcampaignRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Messaging Campaign name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func OutboundMessagingcampaignExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOutboundMessagingcampaign),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`division_id`:                            {RefType: "genesyscloud_auth_division"},
			`contact_list_id`:                        {RefType: "genesyscloud_outbound_contact_list"},
			`contact_list_filter_ids`:                {RefType: "genesyscloud_outbound_contactlistfilter"},
			`dnc_list_ids`:                           {RefType: "genesyscloud_outbound_dnclist"},
			`callable_time_set_id`:                   {RefType: "genesyscloud_outbound_callabletimeset"},
			`rule_set_ids`:                           {RefType: "genesyscloud_outbound_digitalruleset"},
			`sms_config.content_template_id`:         {RefType: "genesyscloud_responsemanagement_response"},
			`email_config.content_template_id`:       {RefType: "genesyscloud_responsemanagement_response"},
			`email_config.reply_to_address.route_id`: {RefType: "genesyscloud_routing_email_route"},

			// TODO: Reference the genesyscloud_routing_email_outbound_domain once it is implemented
			// /api/v2/routing/email/outbound/domains/{domainId}
			`email_config.from_address.domain_id`:     {},
			`email_config.reply_to_address.domain_id`: {},
		},
	}
}
