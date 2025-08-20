package guide

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_guide"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceGuide())
	l.RegisterResource(ResourceType, ResourceGuide())
	l.RegisterExporter(ResourceType, GuideExporter())
}

func ResourceGuide() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Guide",
		CreateContext: provider.CreateWithPooledClient(createGuide),
		ReadContext:   provider.ReadWithPooledClient(readGuide),
		DeleteContext: provider.DeleteWithPooledClient(deleteGuide),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the guide",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func DataSourceGuide() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Guide. Select a guide by name.",
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

func GuideExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGuides),
	}
}
