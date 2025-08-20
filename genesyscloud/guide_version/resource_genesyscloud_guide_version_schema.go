package guide_version

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_guide_version"

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
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceGuideVersion())
	l.RegisterExporter(ResourceType, GuideVersionExporter())
}

func ResourceGuideVersion() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Guide Version",
		CreateContext: provider.CreateWithPooledClient(createGuideVersion),
		ReadContext:   provider.ReadWithPooledClient(readGuideVersion),
		UpdateContext: provider.UpdateWithPooledClient(updateGuideVersion),
		DeleteContext: provider.DeleteWithPooledClient(deleteGuideVersion),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"guide_id": {
				Description: "The ID of the guide this version belongs to.",
				Type:        schema.TypeString,
				Required:    true,
			},
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
}

func GuideVersionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGuideVersions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"guide_id":                             {RefType: "genesyscloud_guide"},
			"resources.data_action.data_action_id": {RefType: "genesyscloud_integration_action"},
		},
		ExcludedAttributes: []string{"generate_content"},
	}
}
