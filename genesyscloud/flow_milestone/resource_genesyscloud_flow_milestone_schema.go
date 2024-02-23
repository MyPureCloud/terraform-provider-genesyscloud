package flow_milestone

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_flow_milestone_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the flow_milestone resource.
3.  The datasource schema definitions for the flow_milestone datasource.
4.  The resource exporter configuration for the flow_milestone exporter.
*/
const resourceName = "genesyscloud_flow_milestone"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceFlowMilestone())
	regInstance.RegisterDataSource(resourceName, DataSourceFlowMilestone())
	regInstance.RegisterExporter(resourceName, FlowMilestoneExporter())
}

// ResourceFlowMilestone registers the genesyscloud_flow_milestone resource with Terraform
func ResourceFlowMilestone() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow milestone`,

		CreateContext: provider.CreateWithPooledClient(createFlowMilestone),
		ReadContext:   provider.ReadWithPooledClient(readFlowMilestone),
		UpdateContext: provider.UpdateWithPooledClient(updateFlowMilestone),
		DeleteContext: provider.DeleteWithPooledClient(deleteFlowMilestone),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The flow milestone name.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"division_id": {
				Description: "The division to which this entity belongs.",
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: "The flow milestone description.",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// FlowMilestoneExporter returns the resourceExporter object used to hold the genesyscloud_flow_milestone exporter's config
func FlowMilestoneExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthFlowMilestones),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

// DataSourceFlowMilestone registers the genesyscloud_flow_milestone data source
func DataSourceFlowMilestone() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow milestone data source. Select a flow milestone by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceFlowMilestoneRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `flow milestone name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
