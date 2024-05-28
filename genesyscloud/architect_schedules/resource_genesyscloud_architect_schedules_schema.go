package architect_schedules

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceName = "genesyscloud_architect_schedules"

// SetRegistrar registers all of the resources, datasources and exporters in the pakage
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectSchedules())
	regInstance.RegisterDataSource(resourceName, DataSourceArchitectSchedules())
	regInstance.RegisterExporter(resourceName, ArchitectSchedulesExporter())
}

// ResourceArchitectSchedules registers the genesyscloud_architect_schedules resource with Terraform
func ResourceArchitectSchedules() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Schedules",

		CreateContext: provider.CreateWithPooledClient(createArchitectSchedules),
		ReadContext:   provider.ReadWithPooledClient(readArchitectSchedules),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectSchedules),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectSchedules),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the schedule.",
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
				Description: "Description of the schedule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"start": {
				Description:      "Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLocalDateTimes,
			},
			"end": {
				Description:      "Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLocalDateTimes,
			},
			"rrule": {
				Description:      "An iCal Recurrence Rule (RRULE) string. It is required to be set for schedules determining when upgrades to the Edge software can be applied.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateRrule,
			},
		},
	}
}

// ArchitectSchedulesExporter returns the resourceExporter object used to hold the genesyscloud_architect_schedules exporter's config
func ArchitectSchedulesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllArchitectSchedules),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
		CustomValidateExports: map[string][]string{
			"rrule": {"rrule"},
		},
	}
}

// DataSourceArchitectSchedules registers the genesyscloud_architect_schedules datat source
func DataSourceArchitectSchedules() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Schedule. Select a schedule by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceArchitectSchedulesRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Schedule name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
