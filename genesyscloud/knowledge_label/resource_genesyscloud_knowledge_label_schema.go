package knowledge_label

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// import other necessary packages here

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceKnowledgeLabel())
	l.RegisterDataSource(ResourceType, dataSourceKnowledgeLabel())
	l.RegisterExporter(ResourceType, KnowledgeLabelExporter())
}

func KnowledgeLabelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeLabels),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

const ResourceType = "genesyscloud_knowledge_label"

var (
	knowledgeLabel = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the label.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"color": {
				Description: "The color for the label.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
)

func ResourceKnowledgeLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Label",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeLabel),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeLabel),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeLabel),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeLabel),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"knowledge_base_id": {
				Description: "Knowledge base id of the label",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_label": {
				Description: "Knowledge label id",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeLabel,
			},
		},
	}
}

func dataSourceKnowledgeLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base Label. Select a label by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceKnowledgeLabelRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base label name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_base_name": {
				Description: "Knowledge base name",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
