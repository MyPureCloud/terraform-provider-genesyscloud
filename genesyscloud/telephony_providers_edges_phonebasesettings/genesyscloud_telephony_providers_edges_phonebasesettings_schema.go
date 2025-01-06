package telephony_providers_edges_phonebasesettings

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ResourceType = "genesyscloud_telephony_providers_edges_phonebasesettings"
)

func ResourcePhoneBaseSettings() *schema.Resource {
	phoneCapabilities := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"provisions": {
				Description: "Provisions",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"registers": {
				Description: "Registers",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"dual_registers": {
				Description: "Dual Registers",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"hardware_id_type": {
				Description: "HardwareId Type",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allow_reboot": {
				Description: "Allow Reboot",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"no_rebalance": {
				Description: "No Rebalance",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"no_cloud_provisioning": {
				Description: "No Cloud Provisioning",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"media_codecs": {
				Description: "Media Codecs",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"audio/opus", "audio/pcmu", "audio/pcma", "audio/g729", "audio/g722"}, false),
				},
			},
			"cdm": {
				Description: "CDM",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}

	lineBase := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"station_persistent_enabled": {
				Description: "The station_persistent_enabled attribute in the line's property",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"station_persistent_timeout": {
				Description: "The station_persistent_timeout attribute in the line's property",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	return &schema.Resource{
		Description: "Genesys Cloud Phone Base Settings",

		CreateContext: provider.CreateWithPooledClient(createPhoneBaseSettings),
		ReadContext:   provider.ReadWithPooledClient(readPhoneBaseSettings),
		UpdateContext: provider.UpdateWithPooledClient(updatePhoneBaseSettings),
		DeleteContext: provider.DeleteWithPooledClient(deletePhoneBaseSettings),
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
			"phone_meta_base_id": {
				Description: "A phone metabase is essentially a database for storing phone configuration settings, which simplifies the configuration process.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"properties": {
				Description:      "phone base settings properties",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"capabilities": {
				Description: "Phone Capabilities.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        phoneCapabilities,
			},
			"line_base": {
				Description: "Line Base Settings for the phonebasesettings",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        lineBase,
			},
			"line_base_settings_id": {
				Description: "This field is computed when a line base is created.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
		CustomizeDiff: customizePhoneBaseSettingsPropertiesDiff,
	}
}

func DataSourcePhoneBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Phone Base Settings. Select a phone base settings by name",
		ReadContext: provider.ReadWithPooledClient(dataSourcePhoneBaseSettingsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Phone Base Settings name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func PhoneBaseSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:     provider.GetAllWithPooledClient(getAllPhoneBaseSettings),
		RefAttrs:             map[string]*resourceExporter.RefAttrSettings{},
		JsonEncodeAttributes: []string{"properties"},
	}
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_phonebasesettings", DataSourcePhoneBaseSettings())
	l.RegisterResource("genesyscloud_telephony_providers_edges_phonebasesettings", ResourcePhoneBaseSettings())
	l.RegisterExporter("genesyscloud_telephony_providers_edges_phonebasesettings", PhoneBaseSettingsExporter())
}
