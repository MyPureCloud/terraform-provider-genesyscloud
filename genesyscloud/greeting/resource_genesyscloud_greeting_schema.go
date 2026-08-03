package greeting

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/greetingmedia"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceGreeting())
	l.RegisterExporter(ResourceType, GreetingExporter())
}

const ResourceType = "genesyscloud_greeting"

func ResourceGreeting() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Greeting",

		CreateContext: provider.CreateWithPooledClient(createGreeting),
		ReadContext:   provider.ReadWithPooledClient(readGreeting),
		UpdateContext: provider.UpdateWithPooledClient(updateGreeting),
		DeleteContext: provider.DeleteWithPooledClient(deleteGreeting),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Greeting name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description: "Greeting type.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"owner_type": {
				Description:  "Greeting owner type. ORGANIZATION is the only supported owner type for organization greetings.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ORGANIZATION"}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// API may override owner_type. Suppress diffs when both values exist.
					return oldValue != "" && newValue != ""
				},
			},
			"owner_id": {
				Description: "The ID of the owner (organization) of the greeting.",
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"audio_file": {
				Description: "Greeting audio file.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration_milliseconds": {
							Description: "Greeting audio file duration in milliseconds.",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"size_bytes": {
							Description: "Greeting audio file size in bytes.",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"self_uri": {
							Description: "Greeting audio file self URI.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"audio_tts": {
				Description: "Greeting audio TTS.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"audio_filename": {
				Description: "Path to the greeting audio file used during export and import.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"audio_file_content_hash": {
				Description: "Hash value of the greeting audio file content. Used to detect changes.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func GreetingExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGreetings),
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: greetingmedia.OrganizationGreetingAudioResolver,
			SubDirectory:              greetingmedia.SubDirectory,
		},
		ThirdPartyRefAttrs: []string{
			"audio_filename",
			"audio_file_content_hash",
		},
	}
}
