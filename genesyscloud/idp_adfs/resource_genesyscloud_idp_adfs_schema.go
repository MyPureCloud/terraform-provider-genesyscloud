package idp_adfs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_idp_adfs_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the idp_adfs resource.
3.  The datasource schema definitions for the idp_adfs datasource.
4.  The resource exporter configuration for the idp_adfs exporter.
*/
const resourceName = "genesyscloud_idp_adfs"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceIdpAdfs())
	regInstance.RegisterExporter(resourceName, IdpAdfsExporter())
}

// ResourceIdpAdfs registers the genesyscloud_idp_adfs resource with Terraform
func ResourceIdpAdfs() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud idp adfs`,

		CreateContext: provider.CreateWithPooledClient(createIdpAdfs),
		ReadContext:   provider.ReadWithPooledClient(readIdpAdfs),
		UpdateContext: provider.UpdateWithPooledClient(updateIdpAdfs),
		DeleteContext: provider.DeleteWithPooledClient(deleteIdpAdfs),
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
			`disabled`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`issuer_u_r_i`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`sso_target_u_r_i`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`slo_u_r_i`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`slo_binding`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`relying_party_identifier`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`certificate`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`certificates`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// IdpAdfsExporter returns the resourceExporter object used to hold the genesyscloud_idp_adfs exporter's config
func IdpAdfsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthIdpAdfss),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// // DataSourceIdpAdfs registers the genesyscloud_idp_adfs data source
// func DataSourceIdpAdfs() *schema.Resource {
// 	return &schema.Resource{
// 		Description: `Genesys Cloud idp adfs data source. Select an idp adfs by name`,
// 		ReadContext: provider.ReadWithPooledClient(readIdpAdfs),
// 		Importer: &schema.ResourceImporter{
// 			StateContext: schema.ImportStatePassthroughContext,
// 		},
// 		Schema: map[string]*schema.Schema{
// 			"name": {
// 				Description: `idp adfs name`,
// 				Type:        schema.TypeString,
// 				Required:    true,
// 			},
// 		},
// 	}
// }
