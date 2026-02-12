package knowledge_document_variation

import (
	"log"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
)

const ResourceType = "genesyscloud_knowledge_document_variation"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceKnowledgeDocumentVariation())
	l.RegisterDataSource(ResourceType, dataSourceKnowledgeDocumentVariation())
	l.RegisterExporter(ResourceType, KnowledgeDocumentVariationExporter())
}

// Since API allows infinite nesting of lists/tables, we are setting max depths we will allow in terraform
// FYI - 9 is the number of nested lists (bullets) that MS Word supports
const maxListDepth = 9

// Tables produce a HUGE schema, which slows plan considerably. Thus we are blocking nested tables by default
var maxTableDepth = func() int {
	if featureToggles.KDVToggleExists() {
		log.Printf("%s is set, enabling nested tables support in %s",
			featureToggles.KDVToggleName(), ResourceType)
		// Change this if you want to allow more table depth
		return 3
	}
	return 1
}()

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
				Optional:    false,
				Required:    false,
				Computed:    true,
				Elem:        addressableEntityRef,
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
			"priority": {
				Description: "The priority of the variation",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
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
				Description: "The content for the body.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentBodyBlock,
			},
		},
	}

	documentElement = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value": {
				Description: "A number",
				Type:        schema.TypeFloat,
				Required:    true,
			},
			"unit": {
				Description:  "The unit of the number. Valid values: Em, Percentage, Px",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Em", "Percentage", "Px"}, false),
			},
		},
	}

	documentBodyTableCellProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cell_type": {
				Description:  "The type of cell. Valid values: Cell, HeaderCell",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Cell", "HeaderCell"}, false),
			},
			"horizontal_align": {
				Description:  "The horizontal alignment of the cell. Valid values: Center, Left, Right",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right"}, true),
			},
			"vertical_align": {
				Description:  "The vertical alignment of the cell. Valid values: Top, Middle, Bottom",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Top", "Middle", "Bottom"}, true),
			},
			"col_span": {
				Description: "The number of columns to span for the cell",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"row_span": {
				Description: "The number of rows to span for the cell",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"height": {
				Description: "The height of the cell",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"scope": {
				Description:  "The scope of the cell. Valid values: Row, Column, RowGroup, ColumnGroup, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Row", "Column", "RowGroup", "ColumnGroup", "None"}, false),
			},
			"border_width": {
				Description: "The width of the cell's border",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"border_style": {
				Description:  "The style of the cell's border. Valid values: Solid, Dotted, Dashed, Double, Groove, Ridge, Inset, Outset, Hidden, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Solid", "Dotted", "Dashed", "Double", "Groove", "Ridge", "Inset", "Outset", "Hidden", "None"}, false),
			},
			"border_color": {
				Description: "The color of the cell's border",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "The background color of the cell",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"width": {
				Description: "The width of the cell (without unit)",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"width_with_unit": {
				Description: "The width of the cell (with unit)",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentElement,
			},
		},
	}

	documentBodyTableRowProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"row_type": {
				Description:  "The type of row. Valid values: Header, Footer, Body",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Header", "Footer", "Body"}, false),
			},
			"background_color": {
				Description: "The background color of the row",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"alignment": {
				Description:  "The alignment of the row. Valid values: Center, Left, Right",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right"}, false),
			},
			"border_style": {
				Description:  "The style of the row's border. Valid values: Solid, Dotted, Dashed, Double, Groove, Ridge, Inset, Outset, Hidden, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Solid", "Dotted", "Dashed", "Double", "Groove", "Ridge", "Inset", "Outset", "Hidden", "None"}, false),
			},
			"border_color": {
				Description: "The color of the row's border",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"height": {
				Description: "The height of the row",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
		},
	}

	documentBodyTableProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"width": {
				Description: "The width of the table (without unit)",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"width_with_unit": {
				Description: "The width of the table (with unit)",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentElement,
			},
			"alignment": {
				Description:  "The alignment of the table. Valid values: Center, Left, Right",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right"}, false),
			},
			"height": {
				Description: "The height of the table",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"cell_spacing": {
				Description: "The spacing of cells in the table",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"caption": {
				Description: "The caption of the table",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentBodyTableCaptionBlock,
			},
			"cell_padding": {
				Description: "The padding of cells in the table",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"border_width": {
				Description: "The width of the table's border",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"border_style": {
				Description:  "The style of the table's border. Valid values: Solid, Dotted, Dashed, Double, Groove, Ridge, Inset, Outset, Hidden, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Solid", "Dotted", "Dashed", "Double", "Groove", "Ridge", "Inset", "Outset", "Hidden", "None"}, false),
			},
			"border_color": {
				Description: "The color of the table's border",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "The background color of the table",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	documentBodyTableCaptionItem = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the table caption. Valid Values: Paragraph, Text, Image, Video, OrderedList, UnorderedList",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Paragraph", "Text", "Image", "Video", "OrderedList", "UnorderedList"}, false),
			},
			"text": {
				Description: "Text. It must contain a value if the type of the caption is Text.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentText,
			},
			"image": {
				Description: "Image. It must contain a value if the type of the caption is Image.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyImage,
			},
			"video": {
				Description: "Video. It must contain a value if the type of the caption is Video.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyVideo,
			},
			"list": {
				Description: "List. It must contain a value if the type of the caption is UnorderedList or OrderedList.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyListSchema(0),
			},
			"paragraph": {
				Description: "Paragraph. It must contain a value if the type of the caption is Paragraph.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyParagraph,
			},
		},
	}

	documentBodyTableCaptionBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"blocks": {
				Description: "The list of captions for a Table.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentBodyTableCaptionItem,
			},
		},
	}

	documentBodyBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the block for the body. This determines which body block object (paragraph, image, video, list or table) would have a value.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Paragraph", "Image", "Video", "OrderedList", "UnorderedList", "Table"}, false),
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
				Elem:        documentBodyListSchema(0),
			},
			"table": {
				Description: "Table. It must contain a value if the type of the block is Table.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentBodyTableSchema(0),
			},
		},
	}

	documentBodyParagraph = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"blocks": {
				Description: "The content for the paragraph.",
				Type:        schema.TypeList,
				Required:    true,
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
				Elem:        documentBodyImageProperties,
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
				Elem:        documentBodyVideoProperties,
			},
		},
	}

	documentContentBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the content block. Valid values: Text, Image, Video",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Text", "Image", "Video"}, false),
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
				Description:  "The font size for the text. The valid values in 'em'. Valid values: XxSmall, XSmall, Small, Medium, Large, XLarge, XxLarge, XxxLarge",
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

	documentBodyImageProperties = &schema.Resource{
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
				Description:      "The indentation for the property. The valid values in 'em'",
				Type:             schema.TypeFloat,
				Optional:         true,
				DiffSuppressFunc: suppressFloat32Equivalent,
			},
			"width": {
				Description: "The width (without unit) for the property",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"width_with_unit": {
				Description: "The width (with unit) for the property",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentElement,
			},
			"alt_text": {
				Description: "The image alt text for the property",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	documentBodyVideoProperties = &schema.Resource{
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
				Description:      "The indentation for the property. The valid values in 'em'",
				Type:             schema.TypeFloat,
				Optional:         true,
				DiffSuppressFunc: suppressFloat32Equivalent,
			},
			"width": {
				Description: "The width (with unit) for the property",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentElement,
			},
			"height": {
				Description: "The height (with unit) for the property",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        documentElement,
			},
		},
	}

	listProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"unordered_type": {
				Description:  "The type of icon for the unordered list. Valid values: Normal, Square, Circle, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Normal", "Square", "Circle", "None"}, false),
			},
			"ordered_type": {
				Description:  "The type of icon for the ordered list. Valid values: Number, LowerAlpha, LowerGreek, LowerRoman, UpperAlpha, UpperRoman, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Number", "LowerAlpha", "LowerGreek", "LowerRoman", "UpperAlpha", "UpperRoman", "None"}, false),
			},
		},
	}

	listBlockProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"font_size": {
				Description:  "The font size for the list item. The valid values in 'em'. Valid values: XxSmall, XSmall, Small, Medium, Large, XLarge, XxLarge, XxxLarge",
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
				Description:  "The align type for the list item. Valid values: Center, Left, Right, Justify",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right", "Justify"}, true),
			},
			"indentation": {
				Description:      "The indentation property for the list item. The valid values in 'em'",
				Type:             schema.TypeFloat,
				Optional:         true,
				DiffSuppressFunc: suppressFloat32Equivalent,
			},
			"unordered_type": {
				Description:  "The type of icon for the unordered list. Valid values: Normal, Square, Circle, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Normal", "Square", "Circle", "None"}, false),
			},
			"ordered_type": {
				Description:  "The type of icon for the ordered list. Valid values: Number, LowerAlpha, LowerGreek, LowerRoman, UpperAlpha, UpperRoman, None",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Number", "LowerAlpha", "LowerGreek", "LowerRoman", "UpperAlpha", "UpperRoman", "None"}, false),
			},
		},
	}

	paragraphProperties = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"font_size": {
				Description:  "The font size for the paragraph. The valid values in 'em'. Valid values: XxSmall, XSmall, Small, Medium, Large, XLarge, XxLarge, XxxLarge",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"XxSmall", "XSmall", "Small", "Medium", "Large", "XLarge", "XxLarge", "XxxLarge"}, true),
			},
			"font_type": {
				Description:  "The font type for the paragraph. Valid values: Paragraph, Heading1, Heading2, Heading3, Heading4, Heading5, Heading6, Preformatted",
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
				Description:  "The align type for the paragraph. Valid values: Center, Left, Right, Justify",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Center", "Left", "Right", "Justify"}, true),
			},
			"indentation": {
				Description:      "The indentation property for the paragraph. The valid values in 'em'",
				Type:             schema.TypeFloat,
				Optional:         true,
				DiffSuppressFunc: suppressFloat32Equivalent,
			},
		},
	}
)

func documentBodyListSchema(depth int) *schema.Resource {
	// Beyond max depth: do not allow nested lists inside list items.
	allowNested := (depth + 1) < maxListDepth
	allowedTypes := []string{"Text", "Image", "Video"}
	if allowNested {
		allowedTypes = append(allowedTypes, "OrderedList", "UnorderedList")
	}

	schemaMap := map[string]*schema.Schema{
		"type": {
			Description:  "The type of the block for the list. This determines which list block object (text, video, image or list) would have a value.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(allowedTypes, false),
		},
		"text": {
			Description: "Text. It must contain a value if the type of the block is Text.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        documentText,
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
	}

	if allowNested {
		schemaMap["list"] = &schema.Schema{
			Description: "List. It must contain a value if the type of the block is UnorderedList or OrderedList.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        documentBodyListSchema(depth + 1),
		}
	}

	documentListContentBlock := &schema.Resource{Schema: schemaMap}

	documentBodyListBlock := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the list. Valid values: ListItem",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ListItem"}, false),
			},
			"blocks": {
				Description: "The list of items for an OrderedList or an UnorderedList.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentListContentBlock,
			},
			"properties": {
				Description: "The properties for the list block",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        listBlockProperties,
			},
		},
	}

	documentBodyList := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"properties": {
				Description: "Properties for the list",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        listProperties,
			},
			"blocks": {
				Description: "The items in the list",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentBodyListBlock,
			},
		},
	}

	return documentBodyList
}

func documentBodyTableSchema(depth int) *schema.Resource {
	// Beyond max depth: do not allow nested tables inside tables.
	allowNested := (depth + 1) < maxTableDepth
	allowedTypes := []string{"Text", "Image", "Video", "OrderedList", "UnorderedList", "Paragraph"}
	if allowNested {
		allowedTypes = append(allowedTypes, "Table")
	}

	schemaMap := map[string]*schema.Schema{
		"type": {
			Description:  "The type of the block for the table. This determines which table block object (text, image, video, list, paragraph, or table) would have a value.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(allowedTypes, false),
		},
		"text": {
			Description: "Text. It must contain a value if the type of the block is Text.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        documentText,
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
			Elem:        documentBodyListSchema(0),
		},
		"paragraph": {
			Description: "Paragraph. It must contain a value if the type of the block is Paragraph.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        documentBodyParagraph,
		},
	}

	if allowNested {
		schemaMap["table"] = &schema.Schema{
			Description: "Table. It must contain a value if the type of the block is Table.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        documentBodyTableSchema(depth + 1),
		}
	}

	documentTableContentBlock := &schema.Resource{Schema: schemaMap}

	documentBodyTableCellBlock := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"properties": {
				Description: "The properties for a row cell",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentBodyTableCellProperties,
			},
			"blocks": {
				Description: "The list of items in a row cell",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentTableContentBlock,
			},
		},
	}

	documentBodyTableRowBlock := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"properties": {
				Description: "The properties for a table row",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentBodyTableRowProperties,
			},
			"cells": {
				Description: "The cells in a table row",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentBodyTableCellBlock,
			},
		},
	}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"properties": {
				Description: "The properties for the table",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentBodyTableProperties,
			},
			"rows": {
				Description: "The rows in the table",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        documentBodyTableRowBlock,
			},
		},
	}
}

func suppressFloat32Equivalent(k, old, new string, d *schema.ResourceData) bool {
	oldFloat, _ := strconv.ParseFloat(old, 32)
	newFloat, _ := strconv.ParseFloat(new, 32)
	return float32(oldFloat) == float32(newFloat)
}

func ResourceKnowledgeDocumentVariation() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Knowledge Document Variation.

Export block label: "{parent knowledge base name}_{parent document title}_{knowledge_document_variation.name}`,

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
				Description: "The knowledge base id of the variation",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_document_id": {
				Description: "The knowledge document id of the variation",
				Type:        schema.TypeString,
				Required:    true,
			},
			"published": {
				Description: "If true, the document will be published with the new variation. If false, the updated document will be in a draft state.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"knowledge_document_variation": {
				Description: "The knowledge document variation",
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
		ExcludedAttributes: []string{
			"knowledge_document_variation.document_version",
		},
	}
}
