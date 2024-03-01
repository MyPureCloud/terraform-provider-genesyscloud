package architect_schedulegroups

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
resource_genesycloud_architect_schedulegroups_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the architect_schedulegroups resource.
3.  The datasource schema definitions for the architect_schedulegroups datasource.
4.  The resource exporter configuration for the architect_schedulegroups exporter.
*/
const resourceName = "genesyscloud_architect_schedulegroups"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectSchedulegroups())
	regInstance.RegisterDataSource(resourceName, DataSourceArchitectSchedulegroups())
	regInstance.RegisterExporter(resourceName, ArchitectSchedulegroupsExporter())
}

// ResourceArchitectSchedulegroups registers the genesyscloud_architect_schedulegroups resource with Terraform
func ResourceArchitectSchedulegroups() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedule Groups",

		CreateContext: provider.CreateWithPooledClient(createArchitectSchedulegroups),
		ReadContext:   provider.ReadWithPooledClient(readArchitectSchedulegroups),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectSchedulegroups),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectSchedulegroups),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the schedule group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_id": {
				Description: "The division to which this schedule group will belong. If not set, the home division will be used. If set, you must have all divisions and future divisions selected in your OAuth client role",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Description of the schedule group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"time_zone": {
				Description: "The timezone the schedules are a part of.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"open_schedules_id": {
				Description: "The schedules defining the hours an organization is open.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"closed_schedules_id": {
				Description: "The schedules defining the hours an organization is closed.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"holiday_schedules_id": {
				Description: "The schedules defining the hours an organization is closed for the holidays.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// ArchitectSchedulegroupsExporter returns the resourceExporter object used to hold the genesyscloud_architect_schedulegroups exporter's config
func ArchitectSchedulegroupsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthArchitectSchedulegroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":          {RefType: "genesyscloud_auth_division"},
			"open_schedules_id":    {RefType: "genesyscloud_architect_schedules"},
			"closed_schedules_id":  {RefType: "genesyscloud_architect_schedules"},
			"holiday_schedules_id": {RefType: "genesyscloud_architect_schedules"},
		},
	}
}

// DataSourceArchitectSchedulegroups registers the genesyscloud_architect_schedulegroups data source
func DataSourceArchitectSchedulegroups() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Schedule Groups. Select a schedule group by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceArchitectSchedulegroupsRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Schedule Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
