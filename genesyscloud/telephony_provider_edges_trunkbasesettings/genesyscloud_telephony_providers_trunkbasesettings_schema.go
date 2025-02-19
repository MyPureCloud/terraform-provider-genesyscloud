package telephony_provider_edges_trunkbasesettings

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ResourceType = "genesyscloud_telephony_providers_edges_trunkbasesettings"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_trunkbasesettings", DataSourceTrunkBaseSettings())
	l.RegisterResource("genesyscloud_telephony_providers_edges_trunkbasesettings", ResourceTrunkBaseSettings())
	l.RegisterExporter("genesyscloud_telephony_providers_edges_trunkbasesettings", TrunkBaseSettingsExporter())

}

func ResourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Trunk Base Settings",

		CreateContext: provider.CreateWithPooledClient(createTrunkBaseSettings),
		ReadContext:   provider.ReadWithPooledClient(readTrunkBaseSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateTrunkBaseSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteTrunkBaseSettings),
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
			"state": {
				Description: "The resource's state.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"trunk_meta_base_id": {
				Description: "The meta-base this trunk is based on.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"properties": {
				Description:      "trunk base settings properties",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"trunk_type": {
				Description:  "The type of this trunk base.Valid values: EXTERNAL, PHONE, EDGE.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EXTERNAL", "PHONE", "EDGE"}, false),
			},
			"managed": {
				Description: "Is this trunk being managed remotely. This property is synchronized with the managed property of the Edge Group to which it is assigned.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"inbound_site_id": {
				Description: "The site to which inbound calls will be routed. Only valid for External BYOC Trunks.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"site_id": {
				Description: "Used to determine the media regions for inbound and outbound calls through a trunk. Also determines the dial plan to use for calls that came in on a trunk and have to be sent out on it as well.  While this is called the site on the API, in the UI it is referred to as the media site.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, //This needs to be computed as the field is prepopulated at time if the field is not set
			},
		},
		CustomizeDiff: util.CustomizeTrunkBaseSettingsPropertiesDiff,
	}
}

func DataSourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Trunk Base Settings. Select a trunk base settings by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceTrunkBaseSettingsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Trunk Base Settings name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func TrunkBaseSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllTrunkBaseSettings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			//"inbound_site_id": {RefType: "genesyscloud_telephony_providers_edges_site"}, TODO: decide how/if this will be included after DEVTOOLING-676 is resolved
		},
		JsonEncodeAttributes: []string{"properties"},
		ExportAsDataFunc:     shouldExportTrunkBaseSettingsAsDataSource,
	}
}
