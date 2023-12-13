package members

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	//resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

var (
	responsemanagementresponseresponsetextResource = &schema.Resource{
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

	responsemanagementresponseresponsefooterResource = &schema.Resource{
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

	responsemanagementresponseresponsesubstitutionResource = &schema.Resource{
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
	responsemanagementresponsemessagingtemplateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`whats_app`: {
				Description: `Defines a messaging template for a WhatsApp messaging channel`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        responsemanagementresponsewhatsappdefinitionResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
		},
	}
	responsemanagementresponsewhatsappdefinitionResource = &schema.Resource{
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

/*
resource_genesycloud_members_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the members resource.
3.  The datasource schema definitions for the members datasource.
4.  The resource exporter configuration for the members exporter.
*/
const resourceName = "genesyscloud_members"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceMembers())
}

// ResourceMembers registers the genesyscloud_members resource with Terraform
func ResourceMembers() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud members`,

		CreateContext: gcloud.CreateWithPooledClient(createMembers),
		ReadContext:   gcloud.ReadWithPooledClient(readMembers),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteMembers),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}
