package scripts

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Defines the resource schema, the datasource, and the exporters for the scripts package
*/
const resourceName = "genesyscloud_script"

// SetRegistrar registers all of the resources, datasources and exporters in the packagee
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceScript())
	l.RegisterResource(resourceName, ResourceScript())
	l.RegisterExporter(resourceName, ExporterScript())
}

// DataSourceScript() returns the data source schema definition
func DataSourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Scripts. Select a script by name.  This will only search on published scripts.  Unpublished scripts will not be returned",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceScriptRead),
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

		CreateContext: gcloud.CreateWithPooledClient(createScript),
		ReadContext:   gcloud.ReadWithPooledClient(readScript),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteScript),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"script_name": {
				Description: "Display name for the script. A reliably unique name is recommended.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"filepath": {
				Description:  "Path to the script file to upload.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: gcloud.ValidatePath,
				ForceNew:     true,
			},
			"file_content_hash": {
				Description: "Hash value of the script file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

// ExporterScript returns all of the exporter configuration for this resource
func ExporterScript() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllScripts),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: ScriptResolver,
			SubDirectory:              "scripts",
		},
	}
}
