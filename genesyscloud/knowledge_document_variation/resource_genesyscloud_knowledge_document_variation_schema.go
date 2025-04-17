package knowledge_document_variation

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"strconv"

	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_knowledge_document_variation"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceKnowledgeDocumentVariation())
	l.RegisterDataSource(ResourceType, dataSourceKnowledgeDocumentVariation())
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
			"context_id": {
				Description: "The globally unique identifier for the knowledge context",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	valuesBody = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value_id": {
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
			"properties": {
				Description: "The properties for the paragraph",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        paragraphProperties,
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
			"properties": {
				Description: "The properties for the image",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        videoImageProperties,
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
			"properties": {
				Description: "The properties for the video",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        videoImageProperties,
			},
		},
	}

	documentBodyList = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"properties": {
				Description: "Properties for the UnorderedList or OrderedList",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        listProperties,
			},
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
			"properties": {
				Description: "The properties for the list block",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        listBlockProperties,
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
			"video": {
				Description: "Video. It must contain a value if the type of the block is Video.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyVideo,
			},
		},
	}

	documentText = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"text": {
				Description: "Text",
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
			"properties": {
				Description: "The properties for the text",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        textProperties,
			},
		},
	}

	textProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"font_size": {
				Description:  "The font size for the text. The valid values in 'em'.Valid values: XxSmall, XSmall, Small, Medium, Large, XLarge, XxLarge, XxxLarge",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"XxSmall", "XSmall", "Small", "Medium", "Large", "XLarge", "XxLarge", "XxxLarge"}, true),
			},
			"text_color": {
				Description: "The text color for the text. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "The background color for the text. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	videoImageProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"background_color": {
				Description: "The background color for the property. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"align": {
				Description:  "The align type for the property. Valid values: Center, Left, Right, Justify",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right", "Justify"}, true),
			},
			"indentation": {
				Description: "The indentation for the property. The valid values in 'em'",
				Type:        schema.TypeFloat,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					oldFloat, _ := strconv.ParseFloat(old, 32)
					newFloat, _ := strconv.ParseFloat(new, 32)
					return float32(oldFloat) == float32(newFloat)
				},
			},
		},
	}

	listProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"unordered_type": {
				Description:  "The type of icon for the unordered list.Valid values: Normal, Square, Circle, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Normal", "Square", "Circle", "None"}, false),
			},
			"ordered_type": {
				Description:  "The type of icon for the ordered list.Valid values: Number, LowerAlpha, LowerGreek, LowerRoman, UpperAlpha, UpperRoman, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Number", "LowerAlpha", "LowerGreek", "LowerRoman", "UpperAlpha", "UpperRoman", "None"}, false),
			},
		},
	}

	listBlockProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"font_size": {
				Description:  "The font size for the list item. The valid values in 'em'.Valid values: XxSmall, XSmall, Small, Medium, Large, XLarge, XxLarge, XxxLarge",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"XxSmall", "XSmall", "Small", "Medium", "Large", "XLarge", "XxLarge", "XxxLarge"}, true),
			},
			"font_type": {
				Description:  "The font type for the list item. Valid values: Paragraph, Heading1, Heading2, Heading3, Heading4, Heading5, Heading6, Preformatted",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Paragraph", "Heading1", "Heading2", "Heading3", "Heading4", "Heading5", "Heading6", "Preformatted"}, true),
			},
			"text_color": {
				Description: "The text color for the list item. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "The background color for the list item. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"align": {
				Description:  "The align type for the list item.Valid values: Center, Left, Right, Justify",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right", "Justify"}, true),
			},
			"indentation": {
				Description: "The indentation property for the list item. The valid values in 'em'",
				Type:        schema.TypeFloat,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					oldFloat, _ := strconv.ParseFloat(old, 32)
					newFloat, _ := strconv.ParseFloat(new, 32)
					return float32(oldFloat) == float32(newFloat)
				},
			},
			"unordered_type": {
				Description:  "The type of icon for the unordered list.Valid values: Normal, Square, Circle, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Normal", "Square", "Circle", "None"}, false),
			},
			"ordered_type": {
				Description:  "The type of icon for the ordered list.Valid values: Number, LowerAlpha, LowerGreek, LowerRoman, UpperAlpha, UpperRoman, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Number", "LowerAlpha", "LowerGreek", "LowerRoman", "UpperAlpha", "UpperRoman", "None"}, false),
			},
		},
	}

	paragraphProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"font_size": {
				Description:  "The font size for the paragraph. The valid values in 'em'.Valid values: XxSmall, XSmall, Small, Medium, Large, XLarge, XxLarge, XxxLarge",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"XxSmall", "XSmall", "Small", "Medium", "Large", "XLarge", "XxLarge", "XxxLarge"}, true),
			},
			"font_type": {
				Description:  "The font type for the paragraph.Valid values: Paragraph, Heading1, Heading2, Heading3, Heading4, Heading5, Heading6, Preformatted",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Paragraph", "Heading1", "Heading2", "Heading3", "Heading4", "Heading5", "Heading6", "Preformatted"}, true),
			},
			"text_color": {
				Description: "The text color for the paragraph. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "The background color for the paragraph. The valid values in hex color code representation. For example black color - #000000",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"align": {
				Description:  "The align type for the paragraph.Valid values: Center, Left, Right, Justify",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right", "Justify"}, true),
			},
			"indentation": {
				Description: "The indentation color for the paragraph. The valid values in 'em'",
				Type:        schema.TypeFloat,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					oldFloat, _ := strconv.ParseFloat(old, 32)
					newFloat, _ := strconv.ParseFloat(new, 32)
					return float32(oldFloat) == float32(newFloat)
				},
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
				Description: "Knowledge document id of the label",
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

func dataSourceKnowledgeDocumentVariation() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Document Variation. Select a knowledge document variation by knowledge document id and variation id",
		ReadContext: provider.ReadWithPooledClient(dataSourceKnowledgeDocumentVariationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the variation",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"knowledge_base_id": {
				Description: "Knowledge base id of the label",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_document_id": {
				Description: "Knowledge document id",
				Type:        schema.TypeString,
				Required:    true,
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
