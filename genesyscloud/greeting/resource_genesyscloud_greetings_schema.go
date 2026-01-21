package greeting

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

const ResourceType = "genesyscloud_greeting"

func ResourceGreeting() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Greetings",

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
				Description:  "Greeting type (STATION | VOICEMAIL | NAME).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"STATION", "VOICEMAIL", "NAME"}, false),
			},
			"owner_type": {
				Description:  "Greeting owner type (USER | ORGANIZATION | GROUP).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"USER", "ORGANIZATION", "GROUP"}, false),
			},
			"owner": {
				Description: "Greeting owner",
				Required:    true,
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
			"created_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"created_by": {
				Description: "Greeting created by.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"modified_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"modified_by": {
				Description: "Greeting modified by.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func GreetingExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGreetings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}
