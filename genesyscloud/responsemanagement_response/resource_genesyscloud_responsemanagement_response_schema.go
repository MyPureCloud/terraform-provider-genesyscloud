package responsemanagement_response

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_responsemanagement_response_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the responsemanagement_response resource.
3.  The datasource schema definitions for the responsemanagement_response datasource.
4.  The resource exporter configuration for the responsemanagement_response exporter.
*/
const resourceName = "genesyscloud_responsemanagement_response"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceResponsemanagementResponse())
	regInstance.RegisterDataSource(resourceName, DataSourceResponsemanagementResponse())
	regInstance.RegisterExporter(resourceName, ResponsemanagementResponseExporter())
}

var (
	responsetextResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`content`: {
				Description: `Response text content.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`content_type`: {
				Description:  `Response text content type.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`text/plain`, `text/html`}, false),
			},
		},
	}

	footerResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description:  `Specifies the type represented by Footer.Valid values: Signature.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Signature`}, false),
			},
			`applicable_resources`: {
				Description: `Specifies the canned response template where the footer can be used.Valid values: Campaign.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	substitutionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`id`: {
				Description: `Response substitution identifier.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `Response substitution description.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_value`: {
				Description: `Response substitution default value.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
	messagingtemplateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`whats_app`: {
				Description: `Defines a messaging template for a WhatsApp messaging channel`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        whatsappDefinitionResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
		},
	}
	whatsappDefinitionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The messaging template name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`namespace`: {
				Description: `The messaging template namespace.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`language`: {
				Description: `The messaging template language configured for this template. This is a WhatsApp specific value. For example, 'en_US'`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

// ResourceResponsemanagementResponse registers the genesyscloud_responsemanagement_response resource with Terraform
func ResourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response`,

		CreateContext: provider.CreateWithPooledClient(createResponsemanagementResponse),
		ReadContext:   provider.ReadWithPooledClient(readResponsemanagementResponse),
		UpdateContext: provider.UpdateWithPooledClient(updateResponsemanagementResponse),
		DeleteContext: provider.DeleteWithPooledClient(deleteResponsemanagementResponse),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the responsemanagement response`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`library_ids`: {
				Description: `One or more libraries response is associated with. Changing the library IDs will result in the resource being recreated`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`texts`: {
				Description: `One or more texts associated with the response.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        responsetextResource,
			},
			`interaction_type`: {
				Description:  `The interaction type for this response.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`chat`, `email`, `twitter`}, false),
			},
			`substitutions`: {
				Description: `Details about any text substitutions used in the texts for this response.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        substitutionResource,
			},
			`substitutions_schema_id`: {
				Description: `Metadata about the text substitutions in json schema format.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`response_type`: {
				Description:  `The response type represented by the response.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`, `Footer`}, false),
			},
			`messaging_template`: {
				Description: `An optional messaging template definition for responseType.MessagingTemplate.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        messagingtemplateResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
			`asset_ids`: {
				Description: `Assets used in the response`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`footer`: {
				Description: `Footer template identifies the Footer type and its footerUsage`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        footerResource,
			},
		},
	}
}

// ResponsemanagementResponseExporter returns the resourceExporter object used to hold the genesyscloud_responsemanagement_response exporter's config
func ResponsemanagementResponseExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthResponsemanagementResponses),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`library_ids`: {
				RefType: "genesyscloud_responsemanagement_library",
			},
			`asset_ids`: {
				RefType: "responsemanagement_responseasset",
			},
		},
		JsonEncodeAttributes: []string{"substitutions_schema_id"},
	}
}

// DataSourceResponsemanagementResponse registers the genesyscloud_responsemanagement_response data source
func DataSourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Responsemanagement Response. Select a Responsemanagement Response by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceResponsemanagementResponseRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Responsemanagement Response name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"library_id": {
				Description: `ID of the library that contains the response.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
