package task_management_workitem

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_task_management_workitem_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_workitem resource.
3.  The datasource schema definitions for the task_management_workitem datasource.
4.  The resource exporter configuration for the task_management_workitem exporter.
*/
const resourceName = "genesyscloud_task_management_workitem"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceTaskManagementWorkitem())
	regInstance.RegisterDataSource(resourceName, DataSourceTaskManagementWorkitem())
	regInstance.RegisterExporter(resourceName, TaskManagementWorkitemExporter())
}

// ResourceTaskManagementWorkitem registers the genesyscloud_task_management_workitem resource with Terraform
func ResourceTaskManagementWorkitem() *schema.Resource {
	workitemScoredAgentResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`agent_id`: {
				Description: `The agent id`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`score`: {
				Description:  `Agent's score for the workitem, from 0 - 100, higher being better`,
				Required:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(0, 100),
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud task management workitem`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementWorkitem),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementWorkitem),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementWorkitem),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementWorkitem),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Workitem.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`worktype_id`: {
				Description: `The Worktype ID of the Workitem.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `The description of the Workitem.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`language_id`: {
				Description: `The language of the Workitem.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`priority`: {
				Description:  `The priority of the Workitem. The valid range is between -25,000,000 and 25,000,000.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(-25000000, 25000000),
			},
			`date_due`: {
				Description:      `The due date of the Workitem. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z`,
				Optional:         true,
				Computed:         true,
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateLocalDateTimes,
			},
			`date_expires`: {
				Description:      `The expiry date of the Workitem. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z`,
				Optional:         true,
				Computed:         true,
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateLocalDateTimes,
			},
			`duration_seconds`: {
				Description: `The estimated duration in seconds to complete the workitem.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`ttl`: {
				Description: `The time to live of the Workitem in seconds.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`status_id`: {
				Description: `The id of the current status of the Workitem.`,
				Optional:    true,
				Computed:    true,
				StateFunc:   task_management_worktype_status.ModifyStatusIdStateValue,
				Type:        schema.TypeString,
			},
			`workbin_id`: {
				Description: `The id of the Workbin that contains the Workitem.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`assignee_id`: {
				Description: `The id of the assignee of the Workitem.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`external_contact_id`: {
				Description: `The id of the external contact of the Workitem.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`external_tag`: {
				Description: `The external tag of the Workitem.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`queue_id`: {
				Description: `The Workitem's queue id.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`skills_ids`: {
				Description: `The ids of skills of the Workitem.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`preferred_agents_ids`: {
				Description: `Ids of the preferred agents of the Workitem.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`auto_status_transition`: {
				Description: `Set it to false to disable auto status transition. By default, it is enabled.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeBool,
			},
			`scored_agents`: {
				Description: `A list of scored agents for the Workitem.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				MaxItems:    20,
				Elem:        workitemScoredAgentResource,
			},
			`custom_fields`: {
				Description:      `JSON formatted object for custom field values defined in the schema referenced by the worktype of the workitem.`,
				Optional:         true,
				Computed:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
		},
	}
}

// TaskManagementWorkitemExporter returns the resourceExporter object used to hold the genesyscloud_task_management_workitem exporter's config
func TaskManagementWorkitemExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementWorkitems),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"worktype_id":            {RefType: "genesyscloud_task_management_worktype"},
			"language_id":            {RefType: "genesyscloud_routing_language"},
			"workbin_id":             {RefType: "genesyscloud_task_management_workbin"},
			"assignee_id":            {RefType: "genesyscloud_user"},
			"preferred_agents_ids":   {RefType: "genesyscloud_user"},
			"scored_agents.agent_id": {RefType: "genesyscloud_user"},
			"external_contact_id":    {RefType: "genesyscloud_externalcontacts_contact"},
			"queue_id":               {RefType: "genesyscloud_routing_queue"},
			"skills_ids":             {RefType: "genesyscloud_routing_skill"},
			"status_id":              {RefType: "genesyscloud_task_management_worktype_status"},
		},
	}
}

// DataSourceTaskManagementWorkitem registers the genesyscloud_task_management_workitem data source
func DataSourceTaskManagementWorkitem() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management workitem data source. Select an task management workitem by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementWorkitemRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Task management workitem name`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"workbin_id": {
				Description: `Id of the workbin where the desired workitem is.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
			"worktype_id": {
				Description: `Id of the worktype of the desired workitem.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
