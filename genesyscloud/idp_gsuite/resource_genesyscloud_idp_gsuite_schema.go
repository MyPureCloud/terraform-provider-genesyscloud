package idp_gsuite

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_idp_gsuite_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the idp_gsuite resource.
3.  The datasource schema definitions for the idp_gsuite datasource.
4.  The resource exporter configuration for the idp_gsuite exporter.
*/
const resourceName = "genesyscloud_idp_gsuite"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceIdpGsuite())
	regInstance.RegisterExporter(resourceName, IdpGsuiteExporter())
}

// ResourceIdpGsuite registers the genesyscloud_idp_gsuite resource with Terraform
func ResourceIdpGsuite() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Single Sign-on GSuite Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-google-g-suite-single-sign-provider/`,

		CreateContext: provider.CreateWithPooledClient(createIdpGsuite),
		ReadContext:   provider.ReadWithPooledClient(readIdpGsuite),
		UpdateContext: provider.UpdateWithPooledClient(updateIdpGsuite),
		DeleteContext: provider.DeleteWithPooledClient(deleteIdpGsuite),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(8 * time.Minute),
			Read:   schema.DefaultTimeout(8 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the provider.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`disabled`: {
				Description: `True if GSuite is disabled.`,
				Optional:    true,
				Type:        schema.TypeBool,
				Default:     false,
			},
			`issuer_uri`: {
				Description: `Issuer URI provided by GSuite.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`target_uri`: {
				Description: `Target URI provided by GSuite.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`slo_uri`: {
				Description: `Provided on app creation.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`slo_binding`: {
				Description:  `Valid values: HTTP Redirect, HTTP Post`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`HTTP Redirect`, `HTTP Post`}, false),
			},
			`relying_party_identifier`: {
				Description: `String used to identify Genesys Cloud to GSuite.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`certificates`: {
				Description: `PEM or DER encoded public X.509 certificates for SAML signature validation.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// IdpGsuiteExporter returns the resourceExporter object used to hold the genesyscloud_idp_gsuite exporter's config
func IdpGsuiteExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthIdpGsuites),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}
