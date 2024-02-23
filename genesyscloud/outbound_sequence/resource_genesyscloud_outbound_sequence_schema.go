package outbound_sequence

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_outbound_sequence_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_sequence resource.
3.  The datasource schema definitions for the outbound_sequence datasource.
4.  The resource exporter configuration for the outbound_sequence exporter.
*/
const resourceName = "genesyscloud_outbound_sequence"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceOutboundSequence())
	regInstance.RegisterDataSource(resourceName, DataSourceOutboundSequence())
	regInstance.RegisterExporter(resourceName, OutboundSequenceExporter())
}

// ResourceOutboundSequence registers the genesyscloud_outbound_sequence resource with Terraform
func ResourceOutboundSequence() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound sequence`,

		CreateContext: provider.CreateWithPooledClient(createOutboundSequence),
		ReadContext:   provider.ReadWithPooledClient(readOutboundSequence),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundSequence),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundSequence),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of outbound sequence`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`campaign_ids`: {
				Description: `The ordered list of Campaigns that this CampaignSequence will run.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`status`: {
				Description:  `The current status of the CampaignSequence. A CampaignSequence can be turned 'on' or 'off' (default). Changing from "on" to "off" will cause the current sequence to drop and be recreated with a new ID.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`on`, `off`}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return (old == `complete` && new == `on`)
				},
			},
			`repeat`: {
				Description: `Indicates if a sequence should repeat from the beginning after the last campaign completes. Default is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
		CustomizeDiff: customdiff.ForceNewIfChange("status", func(ctx context.Context, old, new, meta any) bool {
			return new.(string) == "off" && (old.(string) == "on" || old.(string) == "complete")
		}),
	}
}

// OutboundSequenceExporter returns the resourceExporter object used to hold the genesyscloud_outbound_sequence exporter's config
func OutboundSequenceExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthOutboundSequences),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
		},
	}
}

// DataSourceOutboundSequence registers the genesyscloud_outbound_sequence data source
func DataSourceOutboundSequence() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound sequence data source. Select an outbound sequence by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundSequenceRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Sequence name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
