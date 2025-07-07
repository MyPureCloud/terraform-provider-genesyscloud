package guide_jobs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_guide_jobs"

var (
	variableElem = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the variable.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "The data type of the variable.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"String", "Integer", "Number", "Boolean"}, false),
			},
			"scope": {
				Description:  "The scope that determines the variable's usage context within Guides runtime.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Input", "Output", "InputAndOutput"}, false),
			},
			"description": {
				Description: "The description of the variable used by Guides runtime for input/output handling.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	resourcesElem = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data_action": {
				Description: "The data actions associated with this version of the guide.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        dataActionResource,
			},
		},
	}

	dataActionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data_action_id": {
				Description: "The id of the data action.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"label": {
				Description: "The label of the GC data action as referenced in the guide instruction.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The optional description of the data action.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	guideContentElem = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"instruction": {
				Description: "The instruction given to this version of the guide, for how it should behave when interacting with a User.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"variables": {
				Description: "The variables associated with this version of the guide. Includes input variables (provided) and output variables (captured during execution).",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        variableElem,
			},
			"resources": {
				Description: "The resources associated with this version of the guide.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        resourcesElem,
			},
		},
	}
)

type GenerateGuideContentRequest struct {
	Id          *string `json:"$id,omitempty"`
	Url         *string `json:"url,omitempty"`
	Description *string `json:"description,omitempty"`
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceGuideJobs())
}

func ResourceGuideJobs() *schema.Resource {
	return &schema.Resource{
		Description: "Guide Jobs",

		CreateContext: provider.CreateWithPooledClient(createGuideJob),
		ReadContext:   provider.ReadWithPooledClient(readGuideJob),
		DeleteContext: provider.DeleteWithPooledClient(deleteGuideJob),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "The description that you wish to use to generate the guide content from",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"url": {
				Description: "The URL of the file you wish to use to generate the guide content from",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"status": {
				Description: "The status of the guide job",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"guide_content": {
				Description: "The content of the guide",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    false,
				Required:    false,
				Elem:        guideContentElem,
			},
			"guide_id": {
				Description: "The id of the guide",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
		},
	}
}
