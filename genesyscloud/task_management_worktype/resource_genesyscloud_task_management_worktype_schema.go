package task_management_worktype

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_task_management_worktype_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_worktype resource.
3.  The datasource schema definitions for the task_management_worktype datasource.
4.  The resource exporter configuration for the task_management_worktype exporter.
*/
const resourceName = "genesyscloud_task_management_worktype"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceTaskManagementWorktype())
	regInstance.RegisterDataSource(resourceName, DataSourceTaskManagementWorktype())
	regInstance.RegisterExporter(resourceName, TaskManagementWorktypeExporter())
}

// ResourceTaskManagementWorktype registers the genesyscloud_task_management_worktype resource with Terraform
func ResourceTaskManagementWorktype() *schema.Resource {

	localTimeResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`hour`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`minute`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`second`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`nano`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeInt,
			},
		},
	}

	workitemStatusResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the status`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`category`: {
				Description:  `The Category of the Status.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Open", "InProgress", "Waiting", "Closed"}, false),
			},
			`destination_statuses`: {
				Description: `The Statuses the Status can transition to.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`description`: {
				Description: `The description of the Status.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_destination_status`: {
				Description: `Default destination status to which this Status will transition to if auto status transition enabled.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`status_transition_delay_seconds`: {
				Description: `Delay in seconds for auto status transition`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`status_transition_time`: {
				Description: `Time in HH:MM:SS format at which auto status transition will occur after statusTransitionDelaySeconds delay. To set Time, the statusTransitionDelaySeconds must be equal to or greater than 86400 i.e. a day`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        localTimeResource,
			},
		},
	}

	workitemSchemaResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`json_schema`: {
				Description:      `The JSON Schema document`,
				Required:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: gcloud.SuppressEquivalentJsonDiffs,
			},
			`name`: {
				Description: `The name of the schema`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`enabled`: {
				Description: `The schema's enabled/disabled status. A disabled schema cannot be assigned to any other entities, but the data on those entities from the schema still exists.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud task management worktype`,

		CreateContext: gcloud.CreateWithPooledClient(createTaskManagementWorktype),
		ReadContext:   gcloud.ReadWithPooledClient(readTaskManagementWorktype),
		UpdateContext: gcloud.UpdateWithPooledClient(updateTaskManagementWorktype),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteTaskManagementWorktype),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Worktype.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`default_workbin_id`: {
				Description: `The default Workbin for Workitems created from the Worktype.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division to which this entity belongs.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `The description of the Worktype.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_status`: {
				Description: `The default status for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`statuses`: {
				Description: `The list of possible statuses for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        workitemStatusResource,
			},
			`default_duration_seconds`: {
				Description: `The default duration in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`default_expiration_seconds`: {
				Description: `The default expiration time in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`default_due_duration_seconds`: {
				Description: `The default due duration in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`default_priority`: {
				Description:  `The default priority for Workitems created from the Worktype. The valid range is between -25,000,000 and 25,000,000.`,
				Optional:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(-25000000, 25000000),
			},
			`default_language_id`: {
				Description: `The default routing language for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_ttl_seconds`: {
				Description: `The default time to time to live in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`default_queue_id`: {
				Description: `The default queue for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_skills_ids`: {
				Description: `The default skills for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`assignment_enabled`: {
				Description: `When set to true, Workitems will be sent to the queue of the Worktype as they are created. Default value is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`schema`: {
				Description: `The schema defining the custom attributes for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        workitemSchemaResource,
			},
		},
	}
}

// TaskManagementWorktypeExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype exporter's config
func TaskManagementWorktypeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthTaskManagementWorktypes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":         {RefType: "genesyscloud_auth_division"},
			"default_workbin_id":  {RefType: "genesyscloud_task_management_workbin"},
			"default_language_id": {RefType: "genesyscloud_routing_language"},
			"default_queue_id":    {RefType: "genesyscloud_routing_queue"},
			"default_skills_ids":  {RefType: "genesyscloud_routing_skill"},
		},
	}
}

// DataSourceTaskManagementWorktype registers the genesyscloud_task_management_worktype data source
func DataSourceTaskManagementWorktype() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype data source. Select an task management worktype by name`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceTaskManagementWorktypeRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Task management worktype name`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"default_workbin_id": {
				Description: `The default workbin id assigned to the worktype`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
