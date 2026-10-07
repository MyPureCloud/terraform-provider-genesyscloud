package agentic_virtual_agent

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
   resource_genesyscloud_agentic_virtual_agent_schema.go holds:

   1. The registration code that registers the Resource, DataSource, and Exporter.
   2. The resource schema definition for the agentic_virtual_agent resource.
   3. The data source schema definition for the agentic_virtual_agent data source.
   4. The exporter configuration for the agentic_virtual_agent exporter.
*/

const ResourceType = "genesyscloud_agentic_virtual_agent"

// SetRegistrar registers the resource, data source, and exporter for this package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceAgenticVirtualAgent())
	regInstance.RegisterDataSource(ResourceType, DataSourceAgenticVirtualAgent())
	regInstance.RegisterExporter(ResourceType, AgenticVirtualAgentExporter())
}

// ResourceAgenticVirtualAgent registers the genesyscloud_agentic_virtual_agent resource with Terraform.
func ResourceAgenticVirtualAgent() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Agentic Virtual Agent",
		CreateContext: provider.CreateWithPooledClient(createAgenticVirtualAgent),
		ReadContext:   provider.ReadWithPooledClient(readAgenticVirtualAgent),
		UpdateContext: provider.UpdateWithPooledClient(updateAgenticVirtualAgent),
		DeleteContext: provider.DeleteWithPooledClient(deleteAgenticVirtualAgentResource),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the agentic virtual agent. Must be unique per organization (case-insensitive).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"image_uri": {
				Description: "URI for the agent's avatar image. Must be a valid HTTPS URL.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"status": {
				Description: "Current status of the agent. 'Draft' when no version is published for production, 'Published' when one is.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"latest_saved_version": {
				Description: "Version number of the latest created/updated version.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"latest_production_ready_version": {
				Description: "Version number of the latest production-published version.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// DataSourceAgenticVirtualAgent registers the data source for looking up agents by name.
func DataSourceAgenticVirtualAgent() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Agentic Virtual Agent. Select an agent by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceAgenticVirtualAgentRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the agentic virtual agent to look up.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// AgenticVirtualAgentExporter returns the exporter configuration for this resource.
func AgenticVirtualAgentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAgenticVirtualAgents),
	}
}
