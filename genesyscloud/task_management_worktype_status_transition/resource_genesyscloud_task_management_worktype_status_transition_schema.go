package task_management_worktype_status_transition

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_task_management_worktype_status_transition_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_worktype_status_transition resource.
3.  The datasource schema definitions for the task_management_worktype_status_transition datasource.
4.  The resource exporter configuration for the task_management_worktype_status_transition exporter.
*/
const ResourceType = "genesyscloud_task_management_worktype_status_transition"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceTaskManagementWorktypeStatusTransition())
	regInstance.RegisterDataSource(ResourceType, DataSourceTaskManagementWorktypeStatusTransition())
	regInstance.RegisterExporter(ResourceType, TaskManagementWorktypeStatusTransitionExporter())
}

// ResourceTaskManagementWorktypeStatus registers the genesyscloud_task_management_worktype_status_transition resource with Terraform
func ResourceTaskManagementWorktypeStatusTransition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype status Transition`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementWorkTypeStatusTransition),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementWorkTypeStatusTransition),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementWorkTypeStatusTransition),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementWorkTypeStatusTransition),
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
			`status_id`: {
				Description:  `Name of the status.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(3, 256),
			},
			`destination_status_ids`: {
				Description: `A list of destination Statuses where a Workitem with this Status can transition to. If the list is empty Workitems with this Status can transition to all other Statuses defined on the Worktype. A Status can have a maximum of 24 destinations.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:      schema.TypeString,
					StateFunc: modifyStatusIdStateValue,
				},
				MaxItems: 24,
			},
			`default_destination_status_id`: {
				Description: `Default destination status to which this Status will transition to if auto status transition enabled.`,
				Optional:    true,
				StateFunc:   modifyStatusIdStateValue,
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
		},
	}
}

// TaskManagementWorktypeStatusTransitionExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype_status exporter's config
func TaskManagementWorktypeStatusTransitionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementWorkTypeStatusTransition),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"worktype_id":                   {RefType: "genesyscloud_task_management_worktype"},
			"destination_status_ids.*":      {RefType: "genesyscloud_task_management_worktype_status"},
			"default_destination_status_id": {RefType: "genesyscloud_task_management_worktype_status"},
		},
	}
}

// DataSourceTaskManagementWorktypeStatusTransition registers the genesyscloud_task_management_worktype_status data source
func DataSourceTaskManagementWorktypeStatusTransition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype status transition data source. Select an task management worktype status by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementWorktypeStatusTransitionRead),
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
