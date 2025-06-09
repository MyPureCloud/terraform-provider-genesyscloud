package guides

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_guides"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceGuides())
	l.RegisterResource(ResourceType, ResourceGuides())
	l.RegisterExporter(ResourceType, GuidesExporter())
}

func ResourceGuides() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Guide",
		CreateContext: provider.CreateWithPooledClient(createGuide),
		ReadContext:   provider.ReadWithPooledClient(readGuide),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the guide",
				Type:        schema.TypeString,
				Required:    true,
			},
			"source": {
				Description:  "Indicates how the guide content was generated.Valid values: Manual, Prompt, Document",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Manual", "Prompt", "Document"}, true),
			},
			"status": {
				Description: "The status of the guide.Valid values: Published, Draft",
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"latest_saved_version": {
				Description: "The latest saved version of the guide",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"latest_production_ready_version": {
				Description: "The latest production ready version of the guide",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func DataSourceGuides() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Guides. Select a guide by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceGuideRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the guide",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func GuidesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGuides),
	}
}
