package outbound_campaign

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/outbound"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_outbound_campaign_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_campaign resource.
3.  The datasource schema definitions for the outbound_campaign datasource.
4.  The resource exporter configuration for the outbound_campaign exporter.
*/
const resourceName = "genesyscloud_outbound_campaign"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceOutboundCampaign())
	regInstance.RegisterDataSource(resourceName, DataSourceOutboundCampaign())
	regInstance.RegisterExporter(resourceName, OutboundCampaignExporter())
}

// ResourceOutboundCampaign registers the genesyscloud_outbound_campaign resource with Terraform
func ResourceOutboundCampaign() *schema.Resource {
	outboundcampaignphonecolumnResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the phone column.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud outbound campaign`,

		CreateContext: provider.CreateWithPooledClient(createOutboundCampaign),
		ReadContext:   provider.ReadWithPooledClient(readOutboundCampaign),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundCampaign),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundCampaign),
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
			`max_calls_per_agent`: {
				Description:  `The maximum number of calls that can be placed per agent on this campaign.`,
				Optional:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(1),
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
				Elem:        outbound.OutboundmessagingcampaigncontactsortResource,
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

// OutboundCampaignExporter returns the resourceExporter object used to hold the genesyscloud_outbound_campaign exporter's config
func OutboundCampaignExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthOutboundCampaign),
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
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"campaign_status": {ResolverFunc: resourceExporter.CampaignStatusResolver},
			"script_id": {
				ResolveToDataSourceFunc: resourceExporter.OutboundCampaignAgentScriptResolver,
			},
		},
	}
}

// DataSourceOutboundCampaign registers the genesyscloud_outbound_campaign data source
func DataSourceOutboundCampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound campaign data source. Select an outbound campaign by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundCampaignRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `outbound campaign name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
