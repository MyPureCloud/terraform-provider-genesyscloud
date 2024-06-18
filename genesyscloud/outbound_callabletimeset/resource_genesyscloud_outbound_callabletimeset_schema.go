package outbound_callabletimeset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
)

const resourceName = "genesyscloud_outbound_callabletimeset"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceOutboundCallabletimeset())
	l.RegisterResource(resourceName, ResourceOutboundCallabletimeset())
	l.RegisterExporter(resourceName, OutboundCallableTimesetExporter())
}

var campaignTimeslotResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		`start_time`: {
			Description:      `The start time of the interval as an ISO-8601 string, i.e. HH:mm:ss`,
			Required:         true,
			ValidateDiagFunc: validators.ValidateTime,
			Type:             schema.TypeString,
		},
		`stop_time`: {
			Description:      `The end time of the interval as an ISO-8601 string, i.e. HH:mm:ss`,
			Required:         true,
			ValidateDiagFunc: validators.ValidateTime,
			Type:             schema.TypeString,
		},
		`day`: {
			Description:  `The day of the interval. Valid values: [1-7], representing Monday through Sunday`,
			Required:     true,
			ValidateFunc: validation.IntInSlice([]int{1, 2, 3, 4, 5, 6, 7}),
			Type:         schema.TypeInt,
		},
	},
}

var timeSlotResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		`time_slots`: {
			Description: `The time intervals for which it is acceptable to place outbound calls.`,
			Required:    true,
			Type:        schema.TypeSet,
			Elem:        campaignTimeslotResource,
		},
		`time_zone_id`: {
			Description: `The time zone for the time slots; for example, Africa/Abidjan`,
			Required:    true,
			Type:        schema.TypeString,
		},
	},
}

// ResourceOutboundCallabletimeset registers the genesyscloud_outbound_callabletimeset resource with Terraform
func ResourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound callabletimeset`,

		CreateContext: provider.CreateWithPooledClient(createOutboundCallabletimeset),
		ReadContext:   provider.ReadWithPooledClient(readOutboundCallabletimeset),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundCallabletimeset),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundCallabletimeset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the CallableTimeSet.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`callable_times`: {
				Description: `The list of CallableTimes for which it is acceptable to place outbound calls.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        timeSlotResource,
			},
		},
	}
}

// OutboundCallableTimesetExporter returns the resourceExporter object used to hold the genesyscloud_outbound_callabletimeset exporter's config
func OutboundCallableTimesetExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOutboundCallableTimesets),
	}
}

// dataSourceOutboundCallabletimeset registers the genesyscloud_outbound_callabletimeset data source
func DataSourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Clound Outbound Callable Timesets. Select a callable timeset by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundCallabletimesetRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Callable timeset name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
