package task_management_worktype_flow_datebased_rule

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
resource_genesycloud_task_management_datebased_rule_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_worktype_flow_datebased_rule resource.
3.  The datasource schema definitions for the task_management_worktype_flow_datebased_rule datasource.
4.  The resource exporter configuration for the task_management_worktype_flow_datebased_rule exporter.
*/
const ResourceType = "genesyscloud_task_management_worktype_flow_datebased_rule"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceTaskManagementDateBasedRule())
	regInstance.RegisterDataSource(ResourceType, DataSourceTaskManagementDateBasedRule())
	regInstance.RegisterExporter(ResourceType, TaskManagementDateBasedRuleExporter())
}

// ResourceTaskManagementDateBasedRule registers the genesyscloud_task_management_worktype_flow_datebased_rule resource with Terraform
func ResourceTaskManagementDateBasedRule() *schema.Resource {
	condition := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`attribute`: {
				Description: "The name of the workitem date attribute.",
				Required:    true,
				Type:        schema.TypeString,
			},
			`relative_minutes_to_invocation`: {
				Description: "The time in minutes before or after the date attribute.",
				Required:    true,
				Type:        schema.TypeInt,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud task management onattributeChange Rule`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementDateBasedRule),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementDateBasedRule),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementDateBasedRule),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementDateBasedRule),
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
			"condition": {
				Description: "Condition for this Rule.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        condition,
			},
		},
	}
}

// TaskManagementDateBasedRuleExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype_flow_datebased_rule exporter's config
func TaskManagementDateBasedRuleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementDateBasedRule),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"worktype_id": {RefType: "genesyscloud_task_management_worktype"},
		},
	}
}

// DataSourceTaskManagementDateBasedRule registers the genesyscloud_task_management_worktype_flow_datebased_rule data source
func DataSourceTaskManagementDateBasedRule() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management datebased rule data source. Select a task management datebased rule by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementDateBasedRuleRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`worktype_id`: {
				Description: `The Worktype ID of the Rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}
