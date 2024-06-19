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

const resourceName = "genesyscloud_telephony_providers_edges_site_outbound_route"

var (
	// This is outside the ResourceSite because it is used in a utility function.
	outboundRouteSchema = &schema.Resource{
		Schema: map[string]*schema.Schema{
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
)

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(resourceName, ResourceSiteOutboundRoute())
	l.RegisterExporter(resourceName, SiteExporterOutboundRoute())
}

// ResourceSiteOutboundRoute registers the genesyscloud_telephony_providers_edges_site_outbound_route resource with Terraform
func ResourceSiteOutboundRoute() *schema.Resource {
	return &schema.Resource{
		Description: "Outbound Routes for a Genesys Cloud Site",

		CreateContext: provider.CreateWithPooledClient(createSiteOutboundRoutes),
		ReadContext:   provider.ReadWithPooledClient(readSiteOutboundRoutes),
		UpdateContext: provider.UpdateWithPooledClient(updateSiteOutboundRoutes),
		DeleteContext: provider.DeleteWithPooledClient(deleteSiteOutboundRoutes),
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
			"outbound_routes": {
				Description: "Outbound Routes for the site. The default outbound route for the site will be deleted if routes are specified",
				Type:        schema.TypeSet,
				Required:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        outboundRouteSchema,
			},
		},
	}
}

// SiteExporterOutboundRoute returns the resourceExporter object used to hold the genesyscloud_telephony_providers_edges_site_outbound_route exporter's config
func SiteExporterOutboundRoute() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllSitesOutboundRoutes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"site_id": {RefType: "genesyscloud_telephony_providers_edges_site"},
			"outbound_routes.external_trunk_base_ids": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
		},
	}
}
