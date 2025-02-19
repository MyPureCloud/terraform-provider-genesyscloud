package knowledge_category

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, dataSourceKnowledgeCategory())
	l.RegisterResource(ResourceType, ResourceKnowledgeCategory())
	l.RegisterExporter(ResourceType, KnowledgeCategoryExporter())
}

const ResourceType = "genesyscloud_knowledge_category"

var (
	knowledgeCategory = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name. Changing the name attribute will cause the knowledge_category resource to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Knowledge base description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"parent_id": {
				Description: "Knowledge category parent id",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func KnowledgeCategoryExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeCategories),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id":            {RefType: "genesyscloud_knowledge_knowledgebase"},
			"knowledge_category.parent_id": {RefType: "genesyscloud_knowledge_category"},
		},
	}
}

func ResourceKnowledgeCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Category",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeCategory),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeCategory),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeCategory),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeCategory),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"knowledge_base_id": {
				Description: "Knowledge base id of the category",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_category": {
				Description: "Knowledge category id",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeCategory,
			},
		},
	}
}
