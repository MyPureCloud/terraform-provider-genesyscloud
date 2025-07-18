package architect_flow

import (
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	ResourceType = "genesyscloud_flow"
)

// SetRegistrar registers all resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceArchitectFlow())
	l.RegisterResource(ResourceType, ResourceArchitectFlow())
	l.RegisterExporter(ResourceType, ArchitectFlowExporter())
}

const ExportSubDirectoryName = "architect_flows"

func ArchitectFlowExporter() *resourceExporter.ResourceExporter {

	legacyExporter := &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": ResourceArchitectFlow().Schema["filepath"],
		},
		CustomFlowResolver: map[string]*resourceExporter.CustomFlowResolver{
			"file_content_hash": {ResolverFunc: resourceExporter.FileContentHashResolver},
		},
	}

	// new feature
	newExporter := &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: architectFlowResolver,
			SubDirectory:              ExportSubDirectoryName,
		},
	}

	resourceExporter.SetNewFlowResourceExporter(newExporter)

	return legacyExporter
}

func ResourceArchitectFlow() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Flow.

Export block label: "{type}_{name}"`,

		CreateContext: provider.CreateWithPooledClient(createFlow),
		UpdateContext: provider.UpdateWithPooledClient(updateFlow),
		ReadContext:   provider.ReadWithPooledClient(readFlow),
		DeleteContext: provider.DeleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Flow Name used for export purposes. Note: The 'substitutions' block should be used to set/change 'name' and any other fields in the yaml file",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description: "Flow Type used for export purposes. Note: The 'substitutions' block should be used to set/change 'type' and any other fields in the yaml file",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
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

var validFlowTypes = []string{
	"bot",
	"commonmodule",
	"digitalbot",
	"inboundcall",
	"inboundchat",
	"inboundemail",
	"inboundshortmessage",
	"outboundcall",
	"inqueuecall",
	"inqueueemail",
	"inqueueshortmessage",
	"speech",
	"securecall",
	"surveyinvite",
	"voice",
	"voicemail",
	"voicesurvey",
	"workflow",
	"workitem",
}

func DataSourceArchitectFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Flows. Select a flow by name and type.",
		ReadContext: provider.ReadWithPooledClient(dataSourceFlowRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Flow name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "Flow type. Valid options: " + strings.Join(validFlowTypes, ", "),
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validFlowTypes, true),
			},
		},
	}
}
