package outbound_callabletimeset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_outbound_callabletimeset_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_callabletimeset resource.
3.  The datasource schema definitions for the outbound_callabletimeset datasource.
4.  The resource exporter configuration for the outbound_callabletimeset exporter.
*/
const resourceName = "genesyscloud_outbound_callabletimeset"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceOutboundCallabletimeset())
	regInstance.RegisterDataSource(resourceName, DataSourceOutboundCallabletimeset())
	regInstance.RegisterExporter(resourceName, OutboundCallabletimesetExporter())
}

// ResourceOutboundCallabletimeset registers the genesyscloud_outbound_callabletimeset resource with Terraform
func ResourceOutboundCallabletimeset() *schema.Resource {
	CampaignTimeSlotResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"start_time": {
				Description: "The start time of the interval as an ISO-8601 string, i.e. HH:mm:ss",
				Required:    true,
				Type:        schema.TypeString,
			},
			"stop_time": {
				Description: "The end time of the interval as an ISO-8601 string, i.e. HH:mm:ss",
				Required:    true,
				Type:        schema.TypeString,
			},
			"day": {
				Description: "The day of the interval. Valid values: [1-7], representing Monday through Sunday",
				Required:    true,
				Type:        schema.TypeInt,
			},
		},
	}
	CallableTimeResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"time_slots": {
				Description: "The time intervals for which it is acceptable to place outbound calls.",
				Required:    true,
				Type:        schema.TypeList,
				Elem:        CampaignTimeSlotResource,
			},
			"time_zone_id": {
				Description: "The time zone for the time slots; for example, Africa/Abidjan",
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud Outbound Callabletimeset`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundCallabletimeset),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundCallabletimeset),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundCallabletimeset),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundCallabletimeset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the CallableTimeSet.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"callable_times": {
				Description: "The list of CallableTimes for which it is acceptable to place outbound calls.",
				Required:    true,
				Type:        schema.TypeList,
				Elem:        CallableTimeResource,
			},
		},
	}
}

// OutboundCallabletimesetExporter returns the resourceExporter object used to hold the genesyscloud_outbound_callabletimeset exporter's config
func OutboundCallabletimesetExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthOutboundCallabletimesets),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceOutboundCallabletimeset registers the genesyscloud_outbound_callabletimeset data source
func DataSourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Callabletimeset data source. Select an Outbound Callabletimeset by name`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceOutboundCallabletimesetRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Callabletimeset name`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
