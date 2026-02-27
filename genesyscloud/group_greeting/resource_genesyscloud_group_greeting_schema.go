package group_greeting

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceGroupGreeting())
	l.RegisterExporter(ResourceType, GreetingExporter())
}

const ResourceType = "genesyscloud_group_greeting"

func ResourceGroupGreeting() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Greetings (Group)",

		CreateContext: provider.CreateWithPooledClient(createGroupGreeting),
		ReadContext:   provider.ReadWithPooledClient(readGroupGreeting),
		UpdateContext: provider.UpdateWithPooledClient(updateGroupGreeting),
		DeleteContext: provider.DeleteWithPooledClient(deleteGroupGreeting),
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
				Description:  "Greeting type. VOICEMAIL is the only supported type for group greetings.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"VOICEMAIL"}, false),
			},
			"owner_type": {
				Description:  "Greeting owner type. GROUP is the only supported owner type for group greetings.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"GROUP"}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// API may override owner_type. Suppress diffs when both values exist.
					return oldValue != "" && newValue != ""
				},
			},
			"group_id": {
				Description: "The ID of the group owner of the greeting.",
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// Suppress diffs when both values exist - API may override or user reference resolves differently
					return oldValue != "" && newValue != ""
				},
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
		},
	}
}

func GreetingExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGreetings),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"group_id": {RefType: "genesyscloud_group"},
		},
	}
}
