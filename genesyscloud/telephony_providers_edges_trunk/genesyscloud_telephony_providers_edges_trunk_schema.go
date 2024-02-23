package telephony_providers_edges_trunk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_telephony_providers_edges_trunk"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceTrunk())
	l.RegisterResource(resourceName, ResourceTrunk())
	l.RegisterExporter(resourceName, TrunkExporter())
}

func DataSourceTrunk() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Trunk. Select a trunk by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceTrunkRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Trunk name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ResourceTrunk() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Trunk. Created by assigning a trunk base settings to an edge or edge group",

		CreateContext: provider.CreateWithPooledClient(createTrunk),
		ReadContext:   provider.ReadWithPooledClient(readTrunk),
		UpdateContext: provider.UpdateWithPooledClient(updateTrunk),
		DeleteContext: provider.DeleteWithPooledClient(deleteTrunk),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"trunk_base_settings_id": {
				Description: "The trunk base settings reference",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"edge_group_id": {
				Description: "The edge group associated with this trunk. Either this or \"edge_id\" must be set",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"edge_id": {
				Description: "The edge associated with this trunk. Either this or \"edge_group_id\" must be set",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Description: "The name of the trunk. This property is read only and populated with the auto generated name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}
