package guide_jobs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_guide_jobs"

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
		},
	}
}
