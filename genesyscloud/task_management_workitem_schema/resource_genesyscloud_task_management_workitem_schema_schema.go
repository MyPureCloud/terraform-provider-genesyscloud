package task_management_workitem_schema

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_task_management_workitem_schema_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_workitem_schema resource.
3.  The datasource schema definitions for the task_management_workitem_schema datasource.
4.  The resource exporter configuration for the task_management_workitem_schema exporter.
*/
const ResourceType = "genesyscloud_task_management_workitem_schema"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceTaskManagementWorkitemSchema())
	regInstance.RegisterDataSource(ResourceType, DataSourceTaskManagementWorkitemSchema())
	regInstance.RegisterExporter(ResourceType, TaskManagementWorkitemSchemaExporter())
}

// ResourceTaskManagementWorkitemSchema registers the genesyscloud_task_management_workitem_schema resource with Terraform
func ResourceTaskManagementWorkitemSchema() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management workitem schema`,

		CreateContext: provider.CreateWithPooledClient(createTaskManagementWorkitemSchema),
		ReadContext:   provider.ReadWithPooledClient(readTaskManagementWorkitemSchema),
		UpdateContext: provider.UpdateWithPooledClient(updateTaskManagementWorkitemSchema),
		DeleteContext: provider.DeleteWithPooledClient(deleteTaskManagementWorkitemSchema),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The name of the Workitem Schema",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"description": {
				Description: "The description of the Workitem Schema",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"properties": {
				Description:      "The properties for the JSON Schema document.",
				Optional:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"enabled": {
				Description: `The schema's enabled/disabled status. A disabled schema cannot be assigned to any other entities, but the data on those entities from the schema still exists.`,
				Optional:    true,
				Default:     true,
				Type:        schema.TypeBool,
			},
			"version": {
				Description: `The version number of the Workitem Schema. The version number is incremented each time the schema is modified.`,
				Computed:    true,
				Type:        schema.TypeFloat,
			},
		},
	}
}

// TaskManagementWorkitemSchemaExporter returns the resourceExporter object used to hold the genesyscloud_task_management_workitem_schema exporter's config
func TaskManagementWorkitemSchemaExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:     provider.GetAllWithPooledClient(getAllTaskManagementWorkitemSchemas),
		RefAttrs:             map[string]*resourceExporter.RefAttrSettings{},
		JsonEncodeAttributes: []string{"properties"},
	}
}

// DataSourceTaskManagementWorkitemSchema registers the genesyscloud_task_management_workitem_schema data source
func DataSourceTaskManagementWorkitemSchema() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management workitem schema data source. Select a workitem schema by its name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTaskManagementWorkitemSchemaRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `task management workitem schema name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
