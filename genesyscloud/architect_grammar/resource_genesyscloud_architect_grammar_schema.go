package architect_grammar

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_architect_grammar_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the architect_grammar_language resource.
3.  The datasource schema definitions for the architect_grammar_language datasource.
4.  The resource exporter configuration for the architect_grammar_language exporter.
*/
const resourceName = "genesyscloud_architect_grammar"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectGrammar())
	regInstance.RegisterDataSource(resourceName, DataSourceArchitectGrammar())
	regInstance.RegisterExporter(resourceName, ArchitectGrammarExporter())
}

// ResourceArchitectGrammar registers the genesyscloud_architect_grammar resource with Terraform
func ResourceArchitectGrammar() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud architect grammar`,

		CreateContext: gcloud.CreateWithPooledClient(createArchitectGrammar),
		ReadContext:   gcloud.ReadWithPooledClient(readArchitectGrammar),
		UpdateContext: gcloud.UpdateWithPooledClient(updateArchitectGrammar),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteArchitectGrammar),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: "The name of grammar",
				Required:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: "Description of the grammar",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// ArchitectGrammarExporter returns the resourceExporter object used to hold the genesyscloud_architect_grammar exporter's config
func ArchitectGrammarExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthArchitectGrammar),
	}
}

// DataSourceArchitectGrammar registers the genesyscloud_architect_grammar data source
func DataSourceArchitectGrammar() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Architect Grammar. Select an Architect Grammar by name.`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceArchitectGrammarRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Architect grammar name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
