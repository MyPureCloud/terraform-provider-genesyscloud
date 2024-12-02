package knowledge_document

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// import other necessary packages here

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceKnowledgeDocument())
	l.RegisterExporter(ResourceType, KnowledgeDocumentExporter())
}

func KnowledgeDocumentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeDocuments),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

const ResourceType = "genesyscloud_knowledge_document"

var (
	knowledgeDocument = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"title": {
				Description: "Document title",
				Type:        schema.TypeString,
				Required:    true,
			},
			"visible": {
				Description: "Indicates if the knowledge document should be included in search results.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"alternatives": {
				Description: "List of alternate phrases related to the title which improves search results.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentAlternative,
			},
			"category_name": {
				Description: "The name of the category associated with the document.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"label_names": {
				Description: "The names of labels associated with the document.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	documentAlternative = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"phrase": {
				Description: "Alternate phrasing to the document title.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"autocomplete": {
				Description: "Autocomplete enabled for the alternate phrase.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
)

func ResourceKnowledgeDocument() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge document",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeDocument),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeDocument),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeDocument),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeDocument),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"knowledge_base_id": {
				Description: "Knowledge base id",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_document": {
				Description: "Knowledge document request body",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeDocument,
			},
			"published": {
				Description: "If true, the knowledge document will be published. If false, it will be a draft. The document can only be published if it has document variations.",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}
