package task_management_worktype_status

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_task_management_worktype_status_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_worktype_status resource.
3.  The datasource schema definitions for the task_management_worktype_status datasource.
4.  The resource exporter configuration for the task_management_worktype_status exporter.
*/
const resourceName = "genesyscloud_task_management_worktype_status"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceTaskManagementWorktypeStatus())
	regInstance.RegisterDataSource(resourceName, DataSourceTaskManagementWorktypeStatus())
	regInstance.RegisterExporter(resourceName, TaskManagementWorktypeStatusExporter())
}

// ResourceTaskManagementWorktypeStatus registers the genesyscloud_task_management_worktype_status resource with Terraform
func ResourceTaskManagementWorktypeStatus() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype status`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementWorktypeStatus),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementWorktypeStatus),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementWorktypeStatus),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementWorktypeStatus),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"worktype_id": {
				Description: `The id of the worktype this status belongs to. Changing this attribute will cause the status to be dropped and recreated.`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			`name`: {
				Description:  `Name of the status.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(3, 256),
			},
			`category`: {
				Description:  `The Category of the Status. Changing the category will cause the resource to be dropped and recreated with a new id.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Open", "Waiting", "Closed", "Unknown"}, false),
				ForceNew:     true,
			},
			`description`: {
				Description:  `The description of the Status.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 4096),
			},
			`destination_status_ids`: {
				Description: `A list of destination Statuses where a Workitem with this Status can transition to. If the list is empty Workitems with this Status can transition to all other Statuses defined on the Worktype. A Status can have a maximum of 24 destinations.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:      schema.TypeString,
					StateFunc: ModifyStatusIdStateValue,
				},
				MaxItems: 24,
			},
			`default_destination_status_id`: {
				Description: `Default destination status to which this Status will transition to if auto status transition enabled.`,
				Optional:    true,
				StateFunc:   ModifyStatusIdStateValue,
				Type:        schema.TypeString,
			},
			`status_transition_delay_seconds`: {
				Description:  `Delay in seconds for auto status transition. Required if default_destination_status_id is provided.`,
				Optional:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(60),
			},
			`status_transition_time`: {
				Description: `Time is represented as an ISO-8601 string without a timezone. For example: HH:mm:ss`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default`: {
				Description: `This status is the default status for Workitems created from this Worktype. Only one status can be set as the default status at a time.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
	}
}

// TaskManagementWorktypeStatusExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype_status exporter's config
func TaskManagementWorktypeStatusExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementWorktypeStatuss),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"worktype_id": {RefType: "genesyscloud_task_management_worktype"},
		},
	}
}

// DataSourceTaskManagementWorktypeStatus registers the genesyscloud_task_management_worktype_status data source
func DataSourceTaskManagementWorktypeStatus() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype status data source. Select an task management worktype status by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementWorktypeStatusRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"worktype_id": {
				Description: `The id of the worktype the status belongs to`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: `Task management worktype status name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
