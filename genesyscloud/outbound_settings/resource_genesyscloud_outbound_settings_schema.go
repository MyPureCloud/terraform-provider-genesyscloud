package outbound_settings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
)

/*
resource_genesycloud_outbound_settings_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_settings resource.
3.  The datasource schema definitions for the outbound_settings datasource.
4.  The resource exporter configuration for the outbound_settings exporter.
*/
const resourceName = "genesyscloud_outbound_settings"

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(resourceName, ResourceOutboundSettings())
	l.RegisterExporter(resourceName, OutboundSettingsExporter())
}

var (
	automaticTimeZoneMappingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`callable_windows`: {
				Description: "The time intervals to use for automatic time zone mapping.",
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        callableWindowsResource,
			},
			`supported_countries`: {
				Description: "The countries that are supported for automatic time zone mapping.",
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	callableWindowsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`mapped`: {
				Description: "The time interval to place outbound calls, for contacts that can be mapped to a time zone.",
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        mappedResource,
			},
			`unmapped`: {
				Description: "The time interval and time zone to place outbound calls, for contacts that cannot be mapped to a time zone.",
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        UnmappedResource,
			},
		},
	}
	mappedResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`earliest_callable_time`: {
				Description:      "The earliest time to dial a contact. Valid format is HH:mm",
				Optional:         true,
				ValidateDiagFunc: validators.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
			`latest_callable_time`: {
				Description:      "The latest time to dial a contact. Valid format is HH:mm.",
				Optional:         true,
				ValidateDiagFunc: validators.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
		},
	}
	UnmappedResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`earliest_callable_time`: {
				Description:      "The earliest time to dial a contact. Valid format is HH:mm.",
				Optional:         true,
				ValidateDiagFunc: validators.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
			`latest_callable_time`: {
				Description:      "The latest time to dial a contact. Valid format is HH:mm.",
				Optional:         true,
				ValidateDiagFunc: validators.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
			`time_zone_id`: {
				Description: "The time zone to use for contacts that cannot be mapped.",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

// ResourceOutboundSettings registers the genesyscloud_outbound_settings resource with Terraform
func ResourceOutboundSettings() *schema.Resource {
	return &schema.Resource{
		Description: "An organization's outbound settings",

		CreateContext: provider.CreateWithPooledClient(createOutboundSettings),
		ReadContext:   provider.ReadWithPooledClient(readOutboundSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`max_calls_per_agent`: {
				Description: "The maximum number of calls that can be placed per agent on any campaign.",
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max_line_utilization`: {
				Description:  "The maximum percentage of lines that should be used for Outbound, expressed as a decimal in the range [0.0, 1.0].",
				Optional:     true,
				ValidateFunc: validation.FloatBetween(0.0, 1.0),
				Type:         schema.TypeFloat,
			},
			`abandon_seconds`: {
				Description: "The number of seconds used to determine if a call is abandoned.",
				Optional:    true,
				Type:        schema.TypeFloat,
			},
			`compliance_abandon_rate_denominator`: {
				Description:  "The denominator to be used in determining the compliance abandon rate.Valid values: ALL_CALLS, CALLS_THAT_REACHED_QUEUE.",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ALL_CALLS", "CALLS_THAT_REACHED_QUEUE", ""}, false),
				Type:         schema.TypeString,
			},
			`automatic_time_zone_mapping`: {
				Description: "The settings for automatic time zone mapping. Note that changing these settings will change them for both voice and messaging campaigns.",
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        automaticTimeZoneMappingResource,
			},
			`reschedule_time_zone_skipped_contacts`: {
				Description: "Whether or not to reschedule time-zone blocked contacts.",
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
	}
}

func OutboundSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOutboundSettings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}
