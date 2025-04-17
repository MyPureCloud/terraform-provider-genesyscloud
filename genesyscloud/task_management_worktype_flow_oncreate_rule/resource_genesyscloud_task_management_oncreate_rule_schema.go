package task_management_worktype_flow_oncreate_rule

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
resource_genesycloud_task_management_oncreate_rule_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_worktype_flow_oncreate_rule resource.
3.  The datasource schema definitions for the task_management_worktype_flow_oncreate_rule datasource.
4.  The resource exporter configuration for the task_management_worktype_flow_oncreate_rule exporter.
*/
const ResourceType = "genesyscloud_task_management_worktype_flow_oncreate_rule"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceTaskManagementOnCreateRule())
	regInstance.RegisterDataSource(ResourceType, DataSourceTaskManagementOnCreateRule())
	regInstance.RegisterExporter(ResourceType, TaskManagementOnCreateRuleExporter())
}

// ResourceTaskManagementOnCreateRule registers the genesyscloud_task_management_worktype_flow_oncreate_rule resource with Terraform
func ResourceTaskManagementOnCreateRule() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management oncreate Rule`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementOnCreateRule),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementOnCreateRule),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementOnCreateRule),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementOnCreateRule),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`worktype_id`: {
				Description: `The Worktype ID of the Rule.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// TaskManagementOnCreateRuleExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype_flow_oncreate_rule exporter's config
func TaskManagementOnCreateRuleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementOnCreateRule),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"worktype_id": {RefType: "genesyscloud_task_management_worktype"},
		},
	}
}

// DataSourceTaskManagementOnCreateRule registers the genesyscloud_task_management_worktype_flow_oncreate_rule data source
func DataSourceTaskManagementOnCreateRule() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management oncreate rule data source. Select a task management oncreate rule by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementOnCreateRuleRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Task management oncreate rule name`,
				Type:        schema.TypeString,
				Required:    true,
			},
			`worktype_id`: {
				Description: `The Worktype ID of the Rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}
