package authorization_product

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
genesyscloud_authorization_product_schema holds four functions within it:

1.  The registration code that registers the Datasource for the package.
2.  The datasource schema definitions for the authorization_product datasource.
*/
const ResourceType = "genesyscloud_authorization_product"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(ResourceType, DataSourceAuthorizationProduct())
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
