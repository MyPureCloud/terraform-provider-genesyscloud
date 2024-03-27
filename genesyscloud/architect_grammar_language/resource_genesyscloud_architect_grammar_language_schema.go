package architect_grammar_language

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util/architectlanguages"
)

/*
resource_genesyscloud_architect_grammar_language_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the architect_grammar_language resource.
3.  The datasource schema definitions for the architect_grammar_language datasource.
4.  The resource exporter configuration for the architect_grammar_language exporter.
*/
const resourceName = "genesyscloud_architect_grammar_language"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectGrammarLanguage())
	regInstance.RegisterExporter(resourceName, ArchitectGrammarLanguageExporter())
}

// ResourceArchitectGrammarLanguage registers the genesyscloud_architect_grammar_language resource with Terraform
func ResourceArchitectGrammarLanguage() *schema.Resource {
	fileMetadataResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`file_name`: {
				Description: "The name of the file as defined by the user.",
				Required:    true,
				Type:        schema.TypeString,
			},
			`file_type`: {
				Description:  "The extension of the file.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Gram", "Grxml"}, false),
			},
			"file_content_hash": {
				Description: "Hash value of the file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud architect grammar language`,

		CreateContext: provider.CreateWithPooledClient(createArchitectGrammarLanguage),
		ReadContext:   provider.ReadWithPooledClient(readArchitectGrammarLanguage),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectGrammarLanguage),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectGrammarLanguage),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`grammar_id`: {
				Description: "The id of the grammar this language belongs too. If this is changed a new language is created.",
				Required:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
			},
			`language`: {
				Description:  "Language name. (eg. en-us). If this is changed a new language is created.",
				Required:     true,
				Type:         schema.TypeString,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(architectlanguages.Languages, false),
			},
			`voice_file_data`: {
				Description: "Information about the associated voice file.",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        fileMetadataResource,
			},
			`dtmf_file_data`: {
				Description: "Information about the associated dtmf file.",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        fileMetadataResource,
			},
		},
	}
}

// ArchitectGrammarLanguageExporter returns the resourceExporter object used to hold the genesyscloud_architect_grammar_language exporter's config
func ArchitectGrammarLanguageExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthArchitectGrammarLanguage),
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: ArchitectGrammarLanguageResolver,
			SubDirectory:              "language_files",
		},
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"grammar_id": {RefType: "genesyscloud_architect_grammar"},
		},
	}
}
