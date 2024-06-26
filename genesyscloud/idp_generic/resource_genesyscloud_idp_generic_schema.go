package idp_generic

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_idp_generic_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the idp_generic resource.
3.  The datasource schema definitions for the idp_generic datasource.
4.  The resource exporter configuration for the idp_generic exporter.
*/
const resourceName = "genesyscloud_idp_generic"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceIdpGeneric())
	regInstance.RegisterExporter(resourceName, IdpGenericExporter())
}

// ResourceIdpGeneric registers the genesyscloud_idp_generic resource with Terraform
func ResourceIdpGeneric() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Single Sign-on Generic Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-a-generic-single-sign-on-provider/`,

		CreateContext: provider.CreateWithPooledClient(createIdpGeneric),
		ReadContext:   provider.ReadWithPooledClient(readIdpGeneric),
		UpdateContext: provider.UpdateWithPooledClient(updateIdpGeneric),
		DeleteContext: provider.DeleteWithPooledClient(deleteIdpGeneric),
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
				Required:    true,
				Type:        schema.TypeString,
			},
			`disabled`: {
				Description: `True if Generic provider is disabled.`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			`issuer_uri`: {
				Description: `Issuer URI provided by the provider.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`target_uri`: {
				Description: `Target URI provided by the provider.`,
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
				Description: `String used to identify Genesys Cloud to the identity provider.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`certificates`: {
				Description: `PEM or DER encoded public X.509 certificates for SAML signature validation.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`logo_image_data`: {
				Description: `Base64 encoded SVG image.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`endpoint_compression`: {
				Description: `True if the Genesys Cloud authentication request should be compressed.`,
				Optional:    true,
				Type:        schema.TypeBool,
				Default:     false,
			},
			`name_identifier_format`: {
				Description: `SAML name identifier format. (urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified | urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress | urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName | urn:oasis:names:tc:SAML:1.1:nameid-format:WindowsDomainQualifiedName | urn:oasis:names:tc:SAML:2.0:nameid-format:kerberos | urn:oasis:names:tc:SAML:2.0:nameid-format:entity | urn:oasis:names:tc:SAML:2.0:nameid-format:persistent | urn:oasis:names:tc:SAML:2.0:nameid-format:transient)`,
				Type:        schema.TypeString,
				Optional:    true,
				Default:     `urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified`,
				ValidateFunc: validation.StringInSlice([]string{
					`urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified`,
					`urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress`,
					`urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName`,
					`urn:oasis:names:tc:SAML:1.1:nameid-format:WindowsDomainQualifiedName`,
					`urn:oasis:names:tc:SAML:2.0:nameid-format:kerberos`,
					`urn:oasis:names:tc:SAML:2.0:nameid-format:entity`,
					`urn:oasis:names:tc:SAML:2.0:nameid-format:persistent`,
					`urn:oasis:names:tc:SAML:2.0:nameid-format:transient`,
				}, false),
			},
		},
	}
}

// IdpGenericExporter returns the resourceExporter object used to hold the genesyscloud_idp_generic exporter's config
func IdpGenericExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthIdpGenerics),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}
