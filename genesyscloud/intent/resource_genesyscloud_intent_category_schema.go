package intent_category

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_intent_category_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the intent_category resource.
3.  The datasource schema definitions for the intent_category datasource.
4.  The resource exporter configuration for the intent_category exporter.
*/
const ResourceType = "genesyscloud_intent_category"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceIntentCategory())
	regInstance.RegisterDataSource(ResourceType, DataSourceIntentCategory())
	regInstance.RegisterExporter(ResourceType, IntentCategoryExporter())
}

// ResourceIntentCategory registers the genesyscloud_intent_category resource with Terraform
func ResourceIntentCategory() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud intent category`,
		
		CreateContext: provider.CreateWithPooledClient(createIntentCategory),
		ReadContext:   provider.ReadWithPooledClient(readIntentCategory),
		UpdateContext: provider.UpdateWithPooledClient(updateIntentCategory),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntentCategory),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: { 
			    Description: `Name of the category`,
			    Required: true,
			    Type:   schema.TypeString,
			},
			`description`: { 
			    Description: `Description of the category`,
			    Required: true,
			    Type:   schema.TypeString,
			},
		},
	}
}																																																											

// IntentCategoryExporter returns the resourceExporter object used to hold the genesyscloud_intent_category exporter's config
func IntentCategoryExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthIntentCategories),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceIntentCategory registers the genesyscloud_intent_category data source
func DataSourceIntentCategory() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud intent category data source. Select an intent category by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceIntentCategoryRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `intent category name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
