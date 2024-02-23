package architect_flow

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	resourceName = "genesyscloud_flow"
)

// SetRegistrar registers all resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceArchitectFlow())
	l.RegisterResource(resourceName, ResourceArchitectFlow())
	l.RegisterExporter(resourceName, ArchitectFlowExporter())
}

func ArchitectFlowExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": ResourceArchitectFlow().Schema["filepath"],
		},
		CustomFlowResolver: map[string]*resourceExporter.CustomFlowResolver{
			"file_content_hash": {ResolverFunc: resourceExporter.FileContentHashResolver},
		},
	}
}

func ResourceArchitectFlow() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Flow`,

		CreateContext: provider.CreateWithPooledClient(createFlow),
		UpdateContext: provider.UpdateWithPooledClient(updateFlow),
		ReadContext:   provider.ReadWithPooledClient(readFlow),
		DeleteContext: provider.DeleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"filepath": {
				Description:  "YAML file path for flow configuration. Note: Changing the flow name will result in the creation of a new flow with a new GUID, while the original flow will persist in your org.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the YAML file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"force_unlock": {
				Description: `Will perform a force unlock on an architect flow before beginning the publication process.  NOTE: The force unlock publishes the 'draft'
				              architect flow and then publishes the flow named in this resource. This mirrors the behavior found in the archy CLI tool.`,
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func DataSourceArchitectFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Flows. Select a flow by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceFlowRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Flow name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
