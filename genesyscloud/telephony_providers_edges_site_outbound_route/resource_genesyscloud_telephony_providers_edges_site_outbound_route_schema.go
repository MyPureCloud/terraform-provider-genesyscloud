package telephony_providers_edges_site_outbound_route

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_telephony_providers_edges_site_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the telephony_providers_edges_site_outbound_route resource.
3.  The resource exporter configuration for the telephony_providers_edges_site exporter.
*/

const ResourceType = "genesyscloud_telephony_providers_edges_site_outbound_route"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceSiteOutboundRoute())
	l.RegisterExporter(ResourceType, SiteExporterOutboundRoute())
	l.RegisterDataSource(ResourceType, DataSourceSiteOutboundRoute())
}

// ResourceSiteOutboundRoute registers the genesyscloud_telephony_providers_edges_site_outbound_route resource with Terraform
func ResourceSiteOutboundRoute() *schema.Resource {
	return &schema.Resource{
		Description: "Outbound Routes for a Genesys Cloud Site",

		CreateContext: provider.CreateWithPooledClient(createSiteOutboundRoute),
		ReadContext:   provider.ReadWithPooledClient(readSiteOutboundRoute),
		UpdateContext: provider.UpdateWithPooledClient(updateSiteOutboundRoute),
		DeleteContext: provider.DeleteWithPooledClient(deleteSiteOutboundRoute),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"site_id": {
				Description: "The Id of the site to which the outbound routes belong.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"route_id": {
				Description: "The Id of the outbound route. This is distinct from the \"id\" field. The \"id\" field is a combination of the site_id and route_id",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"classification_types": {
				Description: "Used to classify this outbound route.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"enabled": {
				Description: "Enable or disable the outbound route",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"distribution": {
				Description:  "Valid values: SEQUENTIAL, RANDOM.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "SEQUENTIAL",
				ValidateFunc: validation.StringInSlice([]string{"SEQUENTIAL", "RANDOM"}, false),
			},
			"external_trunk_base_ids": {
				Description: "Trunk base settings of trunkType \"EXTERNAL\". This base must also be set on an edge logical interface for correct routing. The order of the IDs determines the distribution if \"distribution\" is set to \"SEQUENTIAL\"",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// SiteExporterOutboundRoute returns the resourceExporter object used to hold the genesyscloud_telephony_providers_edges_site_outbound_route exporter's config
func SiteExporterOutboundRoute() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllSitesAndOutboundRoutes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"site_id":                 {RefType: "genesyscloud_telephony_providers_edges_site"},
			"external_trunk_base_ids": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
		},
	}
}

// DataSourceSite registers the genesyscloud_telephony_providers_edges_site_outbound_route data source
func DataSourceSiteOutboundRoute() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Site Outbound Routes. Select a Site Outbound Route by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceSiteOutboundRouteRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Outbound Route name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"site_id": {
				Description: "Site Id",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"route_id": {
				Description: "Route Id",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
