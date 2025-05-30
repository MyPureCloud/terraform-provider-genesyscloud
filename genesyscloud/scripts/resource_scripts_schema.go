package scripts

import (
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Defines the resource schema, the datasource, and the exporters for the scripts package
*/
const ResourceType = "genesyscloud_script"

// SetRegistrar registers all the resources, data sources and exporters in the packages
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceScript())
	l.RegisterResource(ResourceType, ResourceScript())
	l.RegisterExporter(ResourceType, ExporterScript())
}

// DataSourceScript returns the data source schema definition
func DataSourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Scripts. Select a script by name.  This will only search on published scripts.  Unpublished scripts will not be returned",
		ReadContext: provider.ReadWithPooledClient(dataSourceScriptRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Script name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// ResourceScript returns the resource script definitions
func ResourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Script",

		CreateContext: provider.CreateWithPooledClient(createScript),
		ReadContext:   provider.ReadWithPooledClient(readScript),
		UpdateContext: provider.UpdateWithPooledClient(updateScript),
		DeleteContext: provider.DeleteWithPooledClient(deleteScript),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"script_name": {
				Description: "Display name for the script. A reliably unique name is recommended. Updating this field will result in the script being dropped and recreated with a new GUID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"filepath": {
				Description:  "Path to the script file to upload.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the script file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"division_id": {
				Description: "Specify division id",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// ExporterScript returns all the exporter configuration for this resource
func ExporterScript() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllScripts),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: authDivision.ResourceType},
		},
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: ScriptResolver,
			SubDirectory:              "scripts",
		},
		DataSourceResolver: map[*resourceExporter.DataAttr]*resourceExporter.ResourceAttr{
			{Attr: "name"}: {Attr: "script_name"},
		},
	}
}
