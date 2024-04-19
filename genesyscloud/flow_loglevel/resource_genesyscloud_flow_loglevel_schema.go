package flow_loglevel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_flow_loglevel_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the flow_loglevel resource.
3.  The datasource schema definitions for the flow_loglevel datasource.
4.  The resource exporter configuration for the flow_loglevel exporter.
*/
const resourceName = "genesyscloud_flow_loglevel"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceFlowLoglevel())
}

// FlowMilestoneExporter returns the resourceExporter object used to hold the genesyscloud_flow_milestone exporter's config
func FlowLogLevelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllFlowLogLevels),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

// ResourceFlowLoglevel registers the genesyscloud_flow_loglevel resource with Terraform
func ResourceFlowLoglevel() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow log level`,

		CreateContext: provider.CreateWithPooledClient(createFlowLogLevel),
		ReadContext:   provider.ReadWithPooledClient(readFlowLogLevel),
		UpdateContext: provider.UpdateWithPooledClient(updateFlowLogLevel),
		DeleteContext: provider.DeleteWithPooledClient(deleteFlowLogLevel),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"flow_id": {
				Description: "The flowId for this characteristics set",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"flow_log_level": {
				Description: "The logLevel for this characteristics set",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
