package knowledgedocumentvariation

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_knowledge_document_variation"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceKnowledgeDocumentVariation())
	l.RegisterExporter(ResourceType, KnowledgeDocumentVariationExporter())
}

var (
	knowledgeDocumentVariation = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"body": {
				Description: "The content for the variation.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        documentBody,
			},
			"document_version": {
				Description: "The version of the document.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        addressableEntityRef,
				Computed:    true,
			},
			"name": {
				Description: "The name of the variation",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"contexts": {
				Description: "The context values associated with the variation",
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        documentVariationContexts,
			},
		},
	}

	documentVariationContexts = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"context": {
				Description: "The knowledge context associated with the variation",
				Required:    true,
				Type:        schema.TypeList,
				Elem:        contextBody,
			},
			"values": {
				Description: "The list of knowledge context values associated with the variation",
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        valuesBody,
			},
		},
	}

	contextBody = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The globally unique identifier for the knowledge context",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	valuesBody = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The globally unique identifier for the knowledge context value",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	documentBody = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"blocks": {
				Description: "The content for the variation.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentBodyBlock,
			},
		},
	}

	documentBodyBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the block for the body. This determines which body block object (paragraph, list, video or image) would have a value.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Paragraph", "Image", "Video", "OrderedList", "UnorderedList"}, false),
			},
			"paragraph": {
				Description: "Paragraph. It must contain a value if the type of the block is Paragraph.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyParagraph,
			},
			"image": {
				Description: "Image. It must contain a value if the type of the block is Image.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyImage,
			},
			"video": {
				Description: "Video. It must contain a value if the type of the block is Video.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyVideo,
			},
			"list": {
				Description: "List. It must contain a value if the type of the block is UnorderedList or OrderedList.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyList,
			},
		},
	}

	documentBodyParagraph = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"blocks": {
				Description: "The content for the variation.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentContentBlock,
			},
		},
	}

	addressableEntityRef = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Id",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	documentBodyImage = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"url": {
				Description: "The URL for the image.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hyperlink": {
				Description: "The URL of the page that the hyperlink goes to.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	documentBodyVideo = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"url": {
				Description: "The URL for the video.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	documentBodyList = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"blocks": {
				Description: "The list of items for an OrderedList or an UnorderedList.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentBodyListBlock,
			},
		},
	}

	documentBodyListBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the list block.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ListItem"}, false),
			},
			"blocks": {
				Description: "The list of items for an OrderedList or an UnorderedList.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentContentBlock,
			},
		},
	}

	documentContentBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the content block.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Text", "Image"}, false),
			},
			"text": {
				Description: "Text. It must contain a value if the type of the block is Text.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        documentText,
			},
			"image": {
				Description: "Image. It must contain a value if the type of the block is Image.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        documentBodyImage,
			},
		},
	}

	documentText = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"text": {
				Description: "Text.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"marks": {
				Description: "The unique list of marks (whether it is bold and/or underlined etc.) for the text. Valid values: Bold | Italic | Underline",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"hyperlink": {
				Description: "The URL of the page that the hyperlink goes to.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func ResourceKnowledgeDocumentVariation() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Document Variation",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeDocumentVariation),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeDocumentVariation),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeDocumentVariation),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeDocumentVariation),
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
			"knowledge_document_id": {
				Description: "Knowledge base id of the label",
				Type:        schema.TypeString,
				Required:    true,
			},
			"published": {
				Description: "If true, the document will be published with the new variation. If false, the updated document will be in a draft state.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"knowledge_document_variation": {
				Description: "Knowledge document variation",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeDocumentVariation,
			},
		},
	}
}

func KnowledgeDocumentVariationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeDocumentVariations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id":     {RefType: "genesyscloud_knowledge_knowledgebase"},
			"knowledge_document_id": {RefType: "genesyscloud_knowledge_document"},
		},
	}
}
