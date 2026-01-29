package greeting_user

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceGreeting())
	l.RegisterExporter(ResourceType, GreetingExporter())
}

const ResourceType = "genesyscloud_greeting_user"

func ResourceGreeting() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Greetings (User)",

		CreateContext: provider.CreateWithPooledClient(createUserGreeting),
		ReadContext:   provider.ReadWithPooledClient(readUserGreeting),
		UpdateContext: provider.UpdateWithPooledClient(updateUserGreeting),
		DeleteContext: provider.DeleteWithPooledClient(deleteUserGreeting),
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
				Description:  "Greeting type. NAME is only supported type for user greetings.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"NAME"}, false),
			},
			"owner_type": {
				Description:  "Greeting owner type. USER is the only supported owner type for user greetings.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"USER"}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// API may override owner_type. Suppress diffs when both values exist.
					return oldValue != "" && newValue != ""
				},
			},
			"user_id": {
				Description: "The ID of the user owner of the greeting.",
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
			"user_id": {RefType: "genesyscloud_user"},
		},
	}
}
