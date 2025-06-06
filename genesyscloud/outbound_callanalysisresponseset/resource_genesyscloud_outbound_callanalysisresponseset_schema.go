package outbound_callanalysisresponseset

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_outbound_callanalysisresponseset_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_callanalysisresponseset resource.
3.  The datasource schema definitions for the outbound_callanalysisresponseset datasource.
4.  The resource exporter configuration for the outbound_callanalysisresponseset exporter.
*/
const ResourceType = "genesyscloud_outbound_callanalysisresponseset"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceOutboundCallanalysisresponseset())
	regInstance.RegisterDataSource(ResourceType, DataSourceOutboundCallanalysisresponseset())
	regInstance.RegisterExporter(ResourceType, OutboundCallanalysisresponsesetExporter())
}

var (
	reactionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`data`: {
				Description: `Parameter for this reaction. For transfer_flow, this would be the outbound flow id.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`name`: {
				Description: `Name of the parameter for this reaction. For transfer_flow, this would be the outbound flow name.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`reaction_type`: {
				Description:  `The reaction to take for a given call analysis result.`,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{`hangup`, `transfer`, `transfer_flow`, `play_file`}, false),
			},
		},
	}

	responseResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`callable_lineconnected`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_person`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_busy`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_noanswer`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_fax`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_disconnect`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_machine`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`callable_sit`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`uncallable_sit`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
			`uncallable_notfound`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     reactionResource,
			},
		},
	}
)

// ResourceOutboundCallanalysisresponseset registers the genesyscloud_outbound_callanalysisresponseset resource with Terraform
func ResourceOutboundCallanalysisresponseset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound Call Analysis Response Set`,

		CreateContext: provider.CreateWithPooledClient(createOutboundCallanalysisresponseset),
		ReadContext:   provider.ReadWithPooledClient(readOutboundCallanalysisresponseset),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundCallanalysisresponseset),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundCallanalysisresponseset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Response Set.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`responses`: {
				Description: `List of maps of disposition identifiers to reactions. Required if beep_detection_enabled = true.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        responseResource,
			},
			`beep_detection_enabled`: {
				Description: `Whether to enable answering machine beep detection`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			`amd_speech_distinguish_enabled`: {
				Description: `Whether to enable answering machine detection`,
				Optional:    true,
				Default:     true,
				Type:        schema.TypeBool,
			},
			`live_speaker_detection_mode`: {
				Description:  `Setting level of live speaker detection based on ringbacks. Valid values: Disabled, Low, Medium, High.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Disabled", "Low", "Medium", "High"}, true),
			},
		},
	}
}

// OutboundCallanalysisresponsesetExporter returns the resourceExporter object used to hold the genesyscloud_outbound_callanalysisresponseset exporter's config
func OutboundCallanalysisresponsesetExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthOutboundCallanalysisresponsesets),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"responses.callable_busy.data":          {RefType: "genesyscloud_flow"},
			"responses.callable_disconnect.data":    {RefType: "genesyscloud_flow"},
			"responses.callable_fax.data":           {RefType: "genesyscloud_flow"},
			"responses.callable_lineconnected.data": {RefType: "genesyscloud_flow"},
			"responses.callable_machine.data":       {RefType: "genesyscloud_flow"},
			"responses.callable_noanswer.data":      {RefType: "genesyscloud_flow"},
			"responses.callable_person.data":        {RefType: "genesyscloud_flow"},
			"responses.callable_sit.data":           {RefType: "genesyscloud_flow"},
			"responses.uncallable_notfound.data":    {RefType: "genesyscloud_flow"},
			"responses.uncallable_sit.data":         {RefType: "genesyscloud_flow"},
		},
	}
}

// DataSourceOutboundCallanalysisresponseset registers the genesyscloud_outbound_callanalysisresponseset data source
func DataSourceOutboundCallanalysisresponseset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound callanalysisresponseset data source. Select an outbound callanalysisresponseset by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundCallanalysisresponsesetRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Data source for Genesys Cloud Outbound Call Analysis Response Sets. Select a response set by name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
