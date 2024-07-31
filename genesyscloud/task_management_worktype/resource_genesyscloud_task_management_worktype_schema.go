package task_management_worktype

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
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
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementWorktype),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementWorktype),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementWorktype),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementWorktype),
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
			`description`: {
				Description: `The description of the Worktype.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division to which this entity belongs.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`default_workbin_id`: {
				Description: `The default Workbin for Workitems created from the Worktype.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`default_duration_seconds`: {
				Description: `The default duration in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`default_expiration_seconds`: {
				Description: `The default expiration time in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`default_due_duration_seconds`: {
				Description: `The default due duration in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`default_priority`: {
				Description:  `The default priority for Workitems created from the Worktype. The valid range is between -25,000,000 and 25,000,000.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(-25000000, 25000000),
			},
			`default_ttl_seconds`: {
				Description: `The default time to time to live in seconds for Workitems created from the Worktype.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`default_language_id`: {
				Description: `The default routing language for Workitems created from the Worktype.`,
				Optional:    true,
				Type:        schema.TypeString,
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
				MaxItems:    20,
			},
			`assignment_enabled`: {
				Description: `When set to true, Workitems will be sent to the queue of the Worktype as they are created. Default value is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`schema_id`: {
				Description: `Id of the workitem schema.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`schema_version`: {
				Description: `Version of the workitem schema to use. If not provided, the worktype will use the latest version.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
		},
	}
}

// TaskManagementWorktypeExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype exporter's config
func TaskManagementWorktypeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementWorktypes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":         {RefType: "genesyscloud_auth_division"},
			"default_workbin_id":  {RefType: "genesyscloud_task_management_workbin"},
			"default_language_id": {RefType: "genesyscloud_routing_language"},
			"default_queue_id":    {RefType: "genesyscloud_routing_queue"},
			"default_skills_ids":  {RefType: "genesyscloud_routing_skill"},
			"schema_id":           {RefType: "genesyscloud_task_management_workitem_schema"},
		},
	}
}

// DataSourceTaskManagementWorktype registers the genesyscloud_task_management_worktype data source
func DataSourceTaskManagementWorktype() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management worktype data source. Select a task management worktype by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementWorktypeRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Task management worktype name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
