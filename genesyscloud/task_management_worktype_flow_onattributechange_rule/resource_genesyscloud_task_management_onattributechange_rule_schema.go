package task_management_worktype_flow_onattributechange_rule

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
resource_genesycloud_task_management_onattributechange_rule_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_worktype_flow_onattributechange_rule resource.
3.  The datasource schema definitions for the task_management_worktype_flow_onattributechange_rule datasource.
4.  The resource exporter configuration for the task_management_worktype_flow_onattributechange_rule exporter.
*/
const ResourceType = "genesyscloud_task_management_worktype_flow_onattributechange_rule"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceTaskManagementOnAttributeChangeRule())
	regInstance.RegisterDataSource(ResourceType, DataSourceTaskManagementOnAttributeChangeRule())
	regInstance.RegisterExporter(ResourceType, TaskManagementOnAttributeChangeRuleExporter())
}

// ResourceTaskManagementOnAttributeChangeRule registers the genesyscloud_task_management_worktype_flow_onattributechange_rule resource with Terraform
func ResourceTaskManagementOnAttributeChangeRule() *schema.Resource {
	condition := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`attribute`: {
				Description: "The name of the workitem attribute whose change will be evaluated as part of the rule.",
				Required:    true,
				Type:        schema.TypeString,
			},
			`new_value`: {
				Description: "The new value of the attribute. If the attribute is updated to this value this part of the condition will be met.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"old_value": {
				Description: "The old value of the attribute. If the attribute was updated from this value this part of the condition will be met.",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud task management onattributeChange Rule`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementOnAttributeChangeRule),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementOnAttributeChangeRule),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementOnAttributeChangeRule),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementOnAttributeChangeRule),
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

// TaskManagementOnAttributeChangeRuleExporter returns the resourceExporter object used to hold the genesyscloud_task_management_worktype_flow_onattributechange_rule exporter's config
func TaskManagementOnAttributeChangeRuleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTaskManagementOnAttributeChangeRule),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"worktype_id": {RefType: "genesyscloud_task_management_worktype"},
			"new_value":   {RefType: "genesyscloud_task_management_worktype_status"},
			"old_value":   {RefType: "genesyscloud_task_management_worktype_status"},
		},
	}
}

// DataSourceTaskManagementOnAttributeChangeRule registers the genesyscloud_task_management_worktype_flow_onattributechange_rule data source
func DataSourceTaskManagementOnAttributeChangeRule() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management onattributechange rule data source. Select a task management onattributechange rule by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementOnAttributeChangeRuleRead),
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
