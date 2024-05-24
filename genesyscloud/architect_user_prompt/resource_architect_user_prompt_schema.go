package architect_user_prompt

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	architectlanguages "terraform-provider-genesyscloud/genesyscloud/util/architectlanguages"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const resourceName = "genesyscloud_architect_user_prompt"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectUserPrompt())
	regInstance.RegisterDataSource(resourceName, DataSourceUserPrompt())
	regInstance.RegisterExporter(resourceName, ArchitectUserPromptExporter())
}
func ArchitectUserPromptExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllUserPrompts),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: ArchitectPromptAudioResolver,
			SubDirectory:              "audio_prompts",
		},
	}
}

func DataSourceUserPrompt() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud User Prompts. Select a user prompt by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceUserPromptRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "User Prompt name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

var userPromptResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"language": {
			Description:  "Language for the prompt resource. (eg. en-us)",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(architectlanguages.Languages, false),
		},
		"tts_string": {
			Description: "Text to Speech (TTS) value for the prompt.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"text": {
			Description: "Text value for the prompt.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"filename": {
			Description: "Path or URL to the file to be uploaded as prompt.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"file_content_hash": {
			Description: "Hash value of the audio file content. Used to detect changes. Only required when uploading a local audio file.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	},
}

func ResourceArchitectUserPrompt() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud User Audio Prompt",

		CreateContext: provider.CreateWithPooledClient(createUserPrompt),
		ReadContext:   provider.ReadWithPooledClient(readUserPrompt),
		UpdateContext: provider.UpdateWithPooledClient(updateUserPrompt),
		DeleteContext: provider.DeleteWithPooledClient(deleteUserPrompt),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the user audio prompt. Note: If the name of the user prompt is changed, this will cause the Prompt to be dropped and recreated with a new ID. This will generate a new ID for the prompt and will invalidate any Architect flows referencing it. ",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Description of the user audio prompt.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"resources": {
				Description: "Audio of TTS resources for the audio prompt.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        userPromptResource,
			},
		},
	}
}
