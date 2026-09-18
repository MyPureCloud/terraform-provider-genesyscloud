package speechandtextanalytics_category

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_speechandtextanalytics_category_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the speechandtextanalytics_category resource.
3.  The datasource schema definitions for the speechandtextanalytics_category datasource.
4.  The resource exporter configuration for the speechandtextanalytics_category exporter.
*/

const ResourceType = "genesyscloud_speechandtextanalytics_category"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceCategory())
	regInstance.RegisterDataSource(ResourceType, DataSourceCategory())
	regInstance.RegisterExporter(ResourceType, CategoryExporter())
}

// ResourceCategory registers the genesyscloud_speechandtextanalytics_category resource with Terraform
func ResourceCategory() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Speech and Text Analytics Category. Categorizes conversations based on a set of criteria.`,

		CreateContext: provider.CreateWithPooledClient(createCategory),
		ReadContext:   provider.ReadWithPooledClient(readCategory),
		UpdateContext: provider.UpdateWithPooledClient(updateCategory),
		DeleteContext: provider.DeleteWithPooledClient(deleteCategory),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the category.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the category.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"interaction_type": {
				Description:  "The type of interaction the category will apply to. Valid values: Voice, Digital, All.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Voice", "Digital", "All"}, false),
			},
			"criteria": {
				Description:      "The category criteria as a JSON string. A collection of conditions joined together by logical operations to provide refined filtering of conversations. Use jsonencode() to construct this value.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
		},
	}
}

// CategoryExporter returns the resourceExporter object used to hold the genesyscloud_speechandtextanalytics_category exporter's config
func CategoryExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllCategories),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		JsonEncodeAttributes: []string{
			"criteria",
		},
	}
}

// DataSourceCategory registers the genesyscloud_speechandtextanalytics_category data source
func DataSourceCategory() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Speech and Text Analytics Categories. Select a category by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceCategoryRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the category.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
