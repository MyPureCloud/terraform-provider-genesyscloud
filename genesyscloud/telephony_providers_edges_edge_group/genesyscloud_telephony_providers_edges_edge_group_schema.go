package telephony_providers_edges_edge_group

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	resourceName = "genesyscloud_telephony_providers_edges_edge_group"
)

func ResourceEdgeGroup() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Edge Group. NOTE: This resource is being kept here for backwards compatibility with older Genesys Cloud Organization. You may get an error if you try to create an edge group with a Genesys Cloud Organization created in 2022 or later.`,

		CreateContext: provider.CreateWithPooledClient(createEdgeGroup),
		ReadContext:   provider.ReadWithPooledClient(readEdgeGroup),
		UpdateContext: provider.UpdateWithPooledClient(updateEdgeGroup),
		DeleteContext: provider.DeleteWithPooledClient(deleteEdgeGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
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
			"state": {
				Description: "Indicates if the resource is active, inactive, or deleted.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"managed": {
				Description: "Is this edge group being managed remotely.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"hybrid": {
				Description: "Is this edge group hybrid.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"phone_trunk_base_ids": {
				Description: "A list of trunk base settings IDs of trunkType \"PHONE\" to inherit to edge logical interface for phone communication.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func DataSourceEdgeGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Edge Group. Select an edge group by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceEdgeGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Edge Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"managed": {
				Description: "Return entities that are managed by Genesys Cloud.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_edge_group", DataSourceEdgeGroup())
	l.RegisterResource("genesyscloud_telephony_providers_edges_edge_group", ResourceEdgeGroup())
	l.RegisterExporter("genesyscloud_telephony_providers_edges_edge_group", EdgeGroupExporter())
}

func EdgeGroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllEdgeGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"phone_trunk_base_ids": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
		},
	}
}
