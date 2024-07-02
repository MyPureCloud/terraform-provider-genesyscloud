package telephony_providers_edges_phone

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_telephony_providers_edges_phone_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the telephony_providers_edges_phone resource.
3.  The datasource schema definitions for the telephony_providers_edges_phone datasource.
4.  The resource exporter configuration for the telephony_providers_edges_phone exporter.
*/
const resourceName = "genesyscloud_telephony_providers_edges_phone"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourcePhone())
	l.RegisterResource(resourceName, ResourcePhone())
	l.RegisterExporter(resourceName, PhoneExporter())
}

// ResourcePhone registers the genesyscloud_telephony_providers_edges_phone resource with Terraform
func ResourcePhone() *schema.Resource {
	lineProperties := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`line_address`: {
				Description: `DID for standalone phones. Each phone number must be in an E.164 phone number format.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: validators.ValidatePhoneNumber},
			},
			`remote_address`: {
				Description: `Station remote property for phones. No validation is provided`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
			},
		},
	}

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

	return &schema.Resource{
		Description: "Genesys Cloud Phone",

		CreateContext: provider.CreateWithPooledClient(createPhone),
		ReadContext:   provider.ReadWithPooledClient(readPhone),
		UpdateContext: provider.UpdateWithPooledClient(updatePhone),
		DeleteContext: provider.DeleteWithPooledClient(deletePhone),
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
				Description:  "Indicates if the resource is active, inactive, or deleted. Valid values: active, inactive, deleted.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "deleted"}, false),
			},
			"site_id": {
				Description: "The site ID associated to the phone.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"phone_base_settings_id": {
				Description: "Phone Base Settings ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"line_base_settings_id": {
				Description: "Line Base Settings ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"phone_meta_base_id": {
				Description: "Phone Meta Base ID.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"web_rtc_user_id": {
				Description: "Web RTC User ID. This is necessary when creating a Web RTC phone. This user will be assigned to the phone after it is created.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"properties": {
				Description:      "phone properties",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"line_properties": {
				Description: "line properties",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        lineProperties,
			},
			"capabilities": {
				Description: "Phone Capabilities.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        phoneCapabilities,
			},
		},
		CustomizeDiff: util.CustomizePhonePropertiesDiff,
	}
}

// PhoneExporter returns the resourceExporter object used to hold the genesyscloud_telephony_providers_edges_phone exporter's config
func PhoneExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllPhones),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"web_rtc_user_id":        {RefType: "genesyscloud_user"},
			"site_id":                {RefType: "genesyscloud_telephony_providers_edges_site"},
			"phone_base_settings_id": {RefType: "genesyscloud_telephony_providers_edges_phonebasesettings"},
		},
	}
}

// DataSourcePhone registers the genesyscloud_telephony_providers_edges_phone data source
func DataSourcePhone() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Phone. Select a phone by name",
		ReadContext: provider.ReadWithPooledClient(dataSourcePhoneRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Phone name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
