package authorization_product

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
genesyscloud_authorization_product_schema holds four functions within it:

1.  The registration code that registers the Datasource for the package.
2.  The datasource schema definitions for the authorization_product datasource.
*/
const resourceName = "genesyscloud_authorization_product"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(resourceName, DataSourceAuthorizationProduct())
}

// DataSourceAuthorizationProduct registers the authorization_product data source
func DataSourceAuthorizationProduct() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Authorisation Products.`,

		ReadContext: provider.ReadWithPooledClient(dataSourceAuthorizationProductRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Authorization Product name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
