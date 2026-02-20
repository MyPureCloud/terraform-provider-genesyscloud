package customer_intent

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_customer_intent_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the customer_intent resource.
3.  The datasource schema definitions for the customer_intent datasource.
4.  The resource exporter configuration for the customer_intent exporter.
*/
const resourceName = "genesyscloud_customer_intent"

// ResourceType is the resource type for customer intent
const ResourceType = resourceName

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceCustomerIntent())
	regInstance.RegisterDataSource(ResourceType, DataSourceCustomerIntent())
	regInstance.RegisterExporter(ResourceType, CustomerIntentExporter())
}

// ResourceCustomerIntent registers the genesyscloud_customer_intent resource with Terraform
func ResourceCustomerIntent() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud customer intent`,

		CreateContext: provider.CreateWithPooledClient(createCustomerIntent),
		ReadContext:   provider.ReadWithPooledClient(readCustomerIntent),
		UpdateContext: provider.UpdateWithPooledClient(updateCustomerIntent),
		DeleteContext: provider.DeleteWithPooledClient(deleteCustomerIntent),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the customer intent`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `Description of the customer intent`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`expiry_time`: {
				Description: `Expiry time in hours of the customer intent`,
				Required:    true,
				Type:        schema.TypeInt,
			},
			`category_id`: {
				Description: `ID of the intent category`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// CustomerIntentExporter returns the resourceExporter object used to hold the genesyscloud_customer_intent exporter's config
func CustomerIntentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthCustomerIntents),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceCustomerIntent registers the genesyscloud_customer_intent data source
func DataSourceCustomerIntent() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud customer intent data source. Select an customer intent by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceCustomerIntentRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `customer intent name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
