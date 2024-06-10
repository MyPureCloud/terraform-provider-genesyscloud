package idp_okta

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_idp_okta_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the idp_okta resource.
3.  The datasource schema definitions for the idp_okta datasource.
4.  The resource exporter configuration for the idp_okta exporter.
*/
const resourceName = "genesyscloud_idp_okta"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceIdpOkta())
	regInstance.RegisterExporter(resourceName, IdpOktaExporter())
}

// ResourceIdpOkta registers the genesyscloud_idp_okta resource with Terraform
func ResourceIdpOkta() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Okta Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-okta-as-a-single-sign-on-provider/",

		CreateContext: provider.CreateWithPooledClient(createIdpOkta),
		ReadContext:   provider.ReadWithPooledClient(readIdpOkta),
		UpdateContext: provider.UpdateWithPooledClient(updateIdpOkta),
		DeleteContext: provider.DeleteWithPooledClient(deleteIdpOkta),
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
				Description: `IDP Okta name`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`disabled`: {
				Description: `True if Okta is disabled.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`issuer_uri`: {
				Description: `Issuer URI provided by Okta.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`target_uri`: {
				Description: `Target URI provided by Okta.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`slo_uri`: {
				Description: `Provided by Okta on app creation.`,
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
				Description: `String used to identify Genesys Cloud to Okta.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`certificates`: {
				Description: `PEM or DER encoded public X.509 certificates for SAML signature validation.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
			},
		},
	}
}

// IdpOktaExporter returns the resourceExporter object used to hold the genesyscloud_idp_okta exporter's config
func IdpOktaExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthIdpOktas),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}
