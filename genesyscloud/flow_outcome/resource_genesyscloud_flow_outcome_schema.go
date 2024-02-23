package flow_outcome

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_flow_outcome_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the flow_outcome resource.
3.  The datasource schema definitions for the flow_outcome datasource.
4.  The resource exporter configuration for the flow_outcome exporter.
*/
const resourceName = "genesyscloud_flow_outcome"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceFlowOutcome())
	regInstance.RegisterDataSource(resourceName, DataSourceFlowOutcome())
	regInstance.RegisterExporter(resourceName, FlowOutcomeExporter())
}

// ResourceFlowOutcome registers the genesyscloud_flow_outcome resource with Terraform
func ResourceFlowOutcome() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow outcome`,

		CreateContext: provider.CreateWithPooledClient(createFlowOutcome),
		ReadContext:   provider.ReadWithPooledClient(readFlowOutcome),
		UpdateContext: provider.UpdateWithPooledClient(updateFlowOutcome),
		DeleteContext: provider.DeleteWithPooledClient(deleteFlowOutcome),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The flow outcome name.",
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
				Description: "This is a description for the flow outcome.",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// FlowOutcomeExporter returns the resourceExporter object used to hold the genesyscloud_flow_outcome exporter's config
func FlowOutcomeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthFlowOutcomes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

// DataSourceFlowOutcome registers the genesyscloud_flow_outcome data source
func DataSourceFlowOutcome() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow outcome data source. Select a flow outcome by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceFlowOutcomeRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `flow outcome name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
