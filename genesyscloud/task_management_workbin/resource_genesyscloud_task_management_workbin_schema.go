package task_management_workbin

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_task_management_workbin_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the task_management_workbin resource.
3.  The datasource schema definitions for the task_management_workbin datasource.
4.  The resource exporter configuration for the task_management_workbin exporter.
*/
const resourceName = "genesyscloud_task_management_workbin"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceTaskManagementWorkbin())
	regInstance.RegisterDataSource(resourceName, DataSourceTaskManagementWorkbin())
	regInstance.RegisterExporter(resourceName, TaskManagementWorkbinExporter())
}

// ResourceTaskManagementWorkbin registers the genesyscloud_task_management_workbin resource with Terraform
func ResourceTaskManagementWorkbin() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management workbin`,

		CreateContext: gcloud.CreateWithPooledClient(createTaskManagementWorkbin),
		ReadContext:   gcloud.ReadWithPooledClient(readTaskManagementWorkbin),
		UpdateContext: gcloud.UpdateWithPooledClient(updateTaskManagementWorkbin),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteTaskManagementWorkbin),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Workbin name",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"division_id": {
				Description: "The division to which this entity belongs.",
				Optional:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
			},
			"description": {
				Description: "Workbin description",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// TaskManagementWorkbinExporter returns the resourceExporter object used to hold the genesyscloud_task_management_workbin exporter's config
func TaskManagementWorkbinExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthTaskManagementWorkbins),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceTaskManagementWorkbin registers the genesyscloud_task_management_workbin data source
func DataSourceTaskManagementWorkbin() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud task management workbin data source. Select an task management workbin by name`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceTaskManagementWorkbinRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `task management workbin name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
