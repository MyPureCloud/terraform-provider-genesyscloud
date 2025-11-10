package workforcemanagement_businessunits

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_workforcemanagement_businessunits_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the workforcemanagement_businessunits resource.
3.  The datasource schema definitions for the workforcemanagement_businessunits datasource.
4.  The resource exporter configuration for the workforcemanagement_businessunits exporter.
*/
const resourceName = "genesyscloud_workforcemanagement_businessunits"

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceWorkforcemanagementBusinessunits())
	regInstance.RegisterDataSource(resourceName, DataSourceWorkforcemanagementBusinessunits())
	regInstance.RegisterExporter(resourceName, WorkforcemanagementBusinessunitsExporter())
}

// ResourceWorkforcemanagementBusinessunits registers the genesyscloud_workforcemanagement_businessunits resource with Terraform
func ResourceWorkforcemanagementBusinessunits() *schema.Resource {
	buShortTermForecastingSettingsResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`default_history_weeks`: {
				Description: `The number of historical weeks to consider when creating a forecast. This setting is only used for legacy weighted average forecasts`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
		},
	}

	schedulerMessageTypeSeverityResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description: `The type of the message`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`severity`: {
				Description: `The severity of the message`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	wfmServiceGoalImpactResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`increase_by_percent`: {
				Description: `The maximum allowed percent increase from the configured goal`,
				Required:    true,
				Type:        schema.TypeFloat,
			},
			`decrease_by_percent`: {
				Description: `The maximum allowed percent decrease from the configured goal`,
				Required:    true,
				Type:        schema.TypeFloat,
			},
		},
	}

	wfmServiceGoalImpactSettingsResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`service_level`: {
				Description: `Allowed service level percent increase and decrease`,
				Required:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        wfmServiceGoalImpactResource,
			},
			`average_speed_of_answer`: {
				Description: `Allowed average speed of answer percent increase and decrease`,
				Required:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        wfmServiceGoalImpactResource,
			},
			`abandon_rate`: {
				Description: `Allowed abandon rate percent increase and decrease`,
				Required:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        wfmServiceGoalImpactResource,
			},
		},
	}

	buSchedulingSettingsResponseResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`message_severities`: {
				Description: `Schedule generation message severity configuration`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        schedulerMessageTypeSeverityResource,
			},
			`sync_time_off_properties`: {
				Description: `Synchronize set of time off properties from scheduled activities to time off requests when the schedule is published.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`service_goal_impact`: {
				Description: `Configures the max percent increase and decrease of service goals for this business unit`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        wfmServiceGoalImpactSettingsResource,
			},
			`allow_work_plan_per_minute_granularity`: {
				Description: `Indicates whether or not per minute granularity for scheduling will be enabled for this business unit. Defaults to false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
	}

	wfmVersionedEntityMetadataResource := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}

	businessUnitSettingsResponseResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`start_day_of_week`: {
				Description: `The start day of week for this business unit`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`time_zone`: {
				Description: `The time zone for this business unit, using the Olsen tz database format`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`short_term_forecasting`: {
				Description: `Short term forecasting settings`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        buShortTermForecastingSettingsResource,
			},
			`scheduling`: {
				Description: `Scheduling settings`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        buSchedulingSettingsResponseResource,
			},
			`metadata`: {
				Description: `Version metadata for this business unit`,
				Required:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        wfmVersionedEntityMetadataResource,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud workforcemanagement businessunits`,

		CreateContext: provider.CreateWithPooledClient(createWorkforcemanagementBusinessUnit),
		ReadContext:   provider.ReadWithPooledClient(readWorkforcemanagementBusinessunits),
		UpdateContext: provider.UpdateWithPooledClient(updateWorkforcemanagementBusinessunits),
		DeleteContext: provider.DeleteWithPooledClient(deleteWorkforcemanagementBusinessunits),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`settings`: {
				Description: `Settings for this business unit`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        businessUnitSettingsResponseResource,
			},
			`division_id`: {
				Description: `The division to which this entity belongs.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// WorkforcemanagementBusinessunitsExporter returns the resourceExporter object used to hold the genesyscloud_workforcemanagement_businessunits exporter's config
func WorkforcemanagementBusinessunitsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthWorkforcemanagementBusinessunits),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceWorkforcemanagementBusinessunits registers the genesyscloud_workforcemanagement_businessunits data source
func DataSourceWorkforcemanagementBusinessunits() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud workforcemanagement businessunits data source. Select an workforcemanagement businessunits by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceWorkforcemanagementBusinessunitsRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `workforcemanagement businessunits name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
