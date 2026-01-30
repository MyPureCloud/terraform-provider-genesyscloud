package knowledge_document_variation

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

// Since API allows infinite nesting of lists/tables, we are setting a max depth we will allow in terraform
const (
	maxListDepth  = 3
	maxTableDepth = 3
)

// Warn only once when max depth is hit
var (
	warnBuildListTruncateOnce    sync.Once
	warnFlattenListTruncateOnce  sync.Once
	warnBuildTableTruncateOnce   sync.Once
	warnFlattenTableTruncateOnce sync.Once
)

func buildDocumentContentListBlocks(blocksIn map[string]interface{}, listDepth int) *[]platformclientv2.Documentlistcontentblock {
	blocksSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentlistcontentblock, 0)

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockOut := platformclientv2.Documentlistcontentblock{
			VarType: resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Text:    buildDocumentText(blockMap),
			Image:   buildDocumentImage(blockMap),
			Video:   buildDocumentVideo(blockMap),
		}
		if _, ok := blockMap["list"]; ok {
			if listDepth+1 < maxListDepth {
				blockOut.List = buildDocumentList(blockMap, listDepth+1)
			} else {
				warnBuildListTruncateOnce.Do(func() {
					log.Printf("[WARN] knowledge_document_variation: nested list exceeds maxListDepth=%d; truncating at attemptedDepth=%d",
						maxListDepth, listDepth+1)
				})
			}
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}

	return &blocksOut
}

func buildDocumentContentBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentcontentblock {
	blocksSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentcontentblock, 0)

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockOut := platformclientv2.Documentcontentblock{
			VarType: resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Text:    buildDocumentText(blockMap),
			Image:   buildDocumentImage(blockMap),
			Video:   buildDocumentVideo(blockMap),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return &blocksOut
}

func buildDocumentListBlocks(blocksIn map[string]interface{}, listDepth int) *[]platformclientv2.Documentbodylistblock {
	blocksSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodylistblock, 0)

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockOut := platformclientv2.Documentbodylistblock{
			VarType:    resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Blocks:     buildDocumentContentListBlocks(blockMap, listDepth),
			Properties: buildDocumentListBlockProperties(blockMap),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return &blocksOut
}

func buildDocumentListBlockProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodylistitemproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	propertiesOut := platformclientv2.Documentbodylistitemproperties{
		FontSize:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_size", false),
		FontType:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_type", false),
		TextColor:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "text_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
		OrderedType:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "ordered_type", false),
		UnorderedType:   resourcedata.GetNillableValueFromMap[string](propertiesMap, "unordered_type", false),
	}
	// If indentation isn't part of state don't include it in api body
	if indentation, ok := propertiesMap["indentation"].(float64); ok && indentation != 0 {
		propertiesOut.Indentation = platformclientv2.Float32(float32(indentation))
	}
	return &propertiesOut
}

func buildDocumentText(textIn map[string]interface{}) *platformclientv2.Documenttext {
	textSlice, ok := textIn["text"].([]interface{})
	if !ok || len(textSlice) == 0 {
		return nil
	}

	textMap, ok := textSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	textOut := platformclientv2.Documenttext{
		Text:       resourcedata.GetNillableValueFromMap[string](textMap, "text", false),
		Hyperlink:  resourcedata.GetNillableValueFromMap[string](textMap, "hyperlink", false),
		Properties: buildTextProperties(textMap),
	}

	// If marks isn't part of state don't include it in api body
	if marks, ok := textMap["marks"].(*schema.Set); ok && marks != nil && marks.Len() > 0 {
		textOut.Marks = lists.SetToStringList(marks)
	}

	return &textOut
}

func buildTextProperties(propertiesIn map[string]interface{}) *platformclientv2.Documenttextproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	propertiesOut := platformclientv2.Documenttextproperties{
		FontSize:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_size", false),
		TextColor:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "text_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
	}
	return &propertiesOut
}

func buildDocumentParagraph(paragraphIn map[string]interface{}) *platformclientv2.Documentbodyparagraph {
	paragraphSlice, ok := paragraphIn["paragraph"].([]interface{})
	if !ok || len(paragraphSlice) == 0 {
		return nil
	}

	paragraphMap, ok := paragraphSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	paragraphOut := platformclientv2.Documentbodyparagraph{
		Blocks:     buildDocumentContentBlocks(paragraphMap),
		Properties: buildParagraphProperties(paragraphMap),
	}
	return &paragraphOut
}

func buildParagraphProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodyparagraphproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	propertiesOut := platformclientv2.Documentbodyparagraphproperties{
		FontSize:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_size", false),
		FontType:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_type", false),
		TextColor:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "text_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
	}

	// If indentation isn't part of state don't include it in api body
	if indentation, ok := propertiesMap["indentation"].(float64); ok && indentation != 0 {
		propertiesOut.Indentation = platformclientv2.Float32(float32(indentation))
	}

	return &propertiesOut
}

func buildDocumentImage(imageIn map[string]interface{}) *platformclientv2.Documentbodyimage {
	imageSlice, ok := imageIn["image"].([]interface{})
	if !ok || len(imageSlice) == 0 {
		return nil
	}

	imageMap, ok := imageSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	imageOut := platformclientv2.Documentbodyimage{
		Url:        resourcedata.GetNillableValueFromMap[string](imageMap, "url", false),
		Hyperlink:  resourcedata.GetNillableValueFromMap[string](imageMap, "hyperlink", false),
		Properties: buildDocumentImageProperties(imageMap),
	}
	return &imageOut
}

func buildDocumentVideoProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodyvideoproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	propertiesOut := platformclientv2.Documentbodyvideoproperties{
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
		Width:           buildDocumentElement(propertiesMap, "width"),
		Height:          buildDocumentElement(propertiesMap, "height"),
	}

	// If indentation isn't part of state don't include it in api body
	if indentation, ok := propertiesMap["indentation"].(float64); ok && indentation != 0 {
		propertiesOut.Indentation = platformclientv2.Float32(float32(indentation))
	}

	return &propertiesOut
}

func buildDocumentImageProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodyimageproperties {

	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	propertiesOut := platformclientv2.Documentbodyimageproperties{
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
		Width:           nillableFloat32FromMap(propertiesMap, "width"),
		WidthWithUnit:   buildDocumentElement(propertiesMap, "width_with_unit"),
		AltText:         resourcedata.GetNillableValueFromMap[string](propertiesMap, "alt_text", false),
	}

	// If indentation isn't part of state don't include it in api body
	if indentation, ok := propertiesMap["indentation"].(float64); ok && indentation != 0 {
		propertiesOut.Indentation = platformclientv2.Float32(float32(indentation))
	}

	return &propertiesOut
}

func buildDocumentVideo(videoIn map[string]interface{}) *platformclientv2.Documentbodyvideo {
	videoSlice, ok := videoIn["video"].([]interface{})
	if !ok || len(videoSlice) == 0 {
		return nil
	}

	videoMap, ok := videoSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	videoOut := platformclientv2.Documentbodyvideo{
		Url:        resourcedata.GetNillableValueFromMap[string](videoMap, "url", false),
		Properties: buildDocumentVideoProperties(videoMap),
	}
	return &videoOut
}

func buildDocumentList(listIn map[string]interface{}, listDepth int) *platformclientv2.Documentbodylist {
	listSlice, ok := listIn["list"].([]interface{})
	if !ok || len(listSlice) == 0 {
		return nil
	}

	listMap, ok := listSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	listOut := platformclientv2.Documentbodylist{
		Blocks:     buildDocumentListBlocks(listMap, listDepth),
		Properties: buildDocumentListProperties(listMap),
	}
	if listOut.Blocks == nil && listOut.Properties == nil {
		return nil
	}
	return &listOut
}

func buildDocumentListProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodylistblockproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	propertiesOut := platformclientv2.Documentbodylistblockproperties{
		OrderedType:   resourcedata.GetNillableValueFromMap[string](propertiesMap, "ordered_type", false),
		UnorderedType: resourcedata.GetNillableValueFromMap[string](propertiesMap, "unordered_type", false),
	}

	return &propertiesOut
}

func buildDocumentBodyBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodyblock {
	blocksSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodyblock, 0)

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockOut := platformclientv2.Documentbodyblock{
			VarType:   resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Paragraph: buildDocumentParagraph(blockMap),
			Image:     buildDocumentImage(blockMap),
			Video:     buildDocumentVideo(blockMap),
			List:      buildDocumentList(blockMap, 0),
			Table:     buildDocumentTable(blockMap, 0),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}

	return &blocksOut
}

func buildVariationBody(bodyIn map[string]interface{}) *platformclientv2.Documentbodyrequest {
	bodySlice, ok := bodyIn["body"].([]interface{})
	if !ok || len(bodySlice) == 0 {
		return nil
	}

	bodyMap, ok := bodySlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	bodyOut := platformclientv2.Documentbodyrequest{
		Blocks: buildDocumentBodyBlocks(bodyMap),
	}
	return &bodyOut
}

func buildKnowledgeDocumentVariation(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	if variationIn == nil {
		return nil
	}

	variationOut := platformclientv2.Documentvariationrequest{
		Name:     resourcedata.GetNillableValueFromMap[string](variationIn, "name", true),
		Body:     buildVariationBody(variationIn),
		Contexts: buildVariationContexts(variationIn),
		Priority: resourcedata.GetNillableValueFromMap[int](variationIn, "priority", false),
	}
	return &variationOut
}

func buildVariationContexts(contextsIn map[string]interface{}) *[]platformclientv2.Documentvariationcontext {
	contextsSlice, ok := contextsIn["contexts"].([]interface{})
	if !ok || len(contextsSlice) == 0 {
		return nil
	}

	contextsOut := make([]platformclientv2.Documentvariationcontext, 0, len(contextsSlice))

	for _, contextAny := range contextsSlice {
		contextMap, ok := contextAny.(map[string]interface{})
		if !ok {
			continue
		}

		contextsOut = append(contextsOut, platformclientv2.Documentvariationcontext{
			Context: buildVariationContext(contextMap),
			Values:  buildVariationContextValue(contextMap),
		})
	}

	if len(contextsOut) == 0 {
		return nil
	}
	return &contextsOut
}

func buildVariationContext(contextIn map[string]interface{}) *platformclientv2.Knowledgecontextreference {
	contextSlice, ok := contextIn["context"].([]interface{})
	if !ok || len(contextSlice) == 0 {
		return nil
	}

	contextMap, ok := contextSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	contextOut := platformclientv2.Knowledgecontextreference{
		Id: resourcedata.GetNillableValueFromMap[string](contextMap, "context_id", true),
	}
	return &contextOut
}

func buildVariationContextValue(valuesIn map[string]interface{}) *[]platformclientv2.Knowledgecontextvaluereference {
	valuesSlice, ok := valuesIn["values"].([]interface{})
	if !ok || len(valuesSlice) == 0 {
		return nil
	}

	valuesOut := make([]platformclientv2.Knowledgecontextvaluereference, 0, len(valuesSlice))

	for _, valueAny := range valuesSlice {
		valueMap, ok := valueAny.(map[string]interface{})
		if !ok {
			continue
		}
		valuesOut = append(valuesOut, platformclientv2.Knowledgecontextvaluereference{
			Id: resourcedata.GetNillableValueFromMap[string](valueMap, "value_id", true),
		})
	}

	if len(valuesOut) == 0 {
		return nil
	}
	return &valuesOut
}

func buildKnowledgeDocumentVariationUpdate(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	variationOut := platformclientv2.Documentvariationrequest{
		Name:     resourcedata.GetNillableValueFromMap[string](variationIn, "name", true),
		Body:     buildVariationBody(variationIn),
		Priority: resourcedata.GetNillableValueFromMap[int](variationIn, "priority", false),
	}
	return &variationOut
}

func buildDocumentTable(tableIn map[string]interface{}, tableDepth int) *platformclientv2.Documentbodytable {
	tableSlice, ok := tableIn["table"].([]interface{})
	if !ok || len(tableSlice) == 0 {
		return nil
	}

	tableMap, ok := tableSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}
	tableOut := platformclientv2.Documentbodytable{
		Properties: buildDocumentTableProperties(tableMap),
		Rows:       buildDocumentTableRowBlocks(tableMap, tableDepth),
	}
	if tableOut.Rows == nil && tableOut.Properties == nil {
		return nil
	}
	return &tableOut
}

func buildDocumentTableProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodytableproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}
	propertiesOut := platformclientv2.Documentbodytableproperties{
		Width:           nillableFloat32FromMap(propertiesMap, "width"),
		WidthWithUnit:   buildDocumentElement(propertiesMap, "width_with_unit"),
		Alignment:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "alignment", false),
		Height:          nillableFloat32FromMap(propertiesMap, "height"),
		CellSpacing:     nillableFloat32FromMap(propertiesMap, "cell_spacing"),
		Caption:         buildDocumentTableCaption(propertiesMap),
		CellPadding:     nillableFloat32FromMap(propertiesMap, "cell_padding"),
		BorderWidth:     nillableFloat32FromMap(propertiesMap, "border_width"),
		BorderStyle:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "border_style", false),
		BorderColor:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "border_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
	}
	return &propertiesOut
}

func buildDocumentTableCaption(captionIn map[string]interface{}) *platformclientv2.Documentbodytablecaptionblock {
	captionSlice, ok := captionIn["caption"].([]interface{})
	if !ok || len(captionSlice) == 0 {
		return nil
	}

	captionMap, ok := captionSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}
	captionOut := platformclientv2.Documentbodytablecaptionblock{
		Blocks: buildDocumentTableCaptionBlocks(captionMap),
	}
	return &captionOut
}

func buildDocumentTableCaptionBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodytablecaptionitem {
	blocksSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodytablecaptionitem, 0)

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		blockOut := platformclientv2.Documentbodytablecaptionitem{
			VarType:   resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Text:      buildDocumentText(blockMap),
			Image:     buildDocumentImage(blockMap),
			Video:     buildDocumentVideo(blockMap),
			List:      buildDocumentList(blockMap, 0),
			Paragraph: buildDocumentParagraph(blockMap),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return &blocksOut
}

func buildDocumentElement(elementIn map[string]interface{}, key string) *platformclientv2.Documentelementlength {
	elementSlice, ok := elementIn[key].([]interface{})
	if !ok || len(elementSlice) == 0 {
		return nil
	}

	elementMap, ok := elementSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}
	val := nillableFloat32FromMap(elementMap, "value")
	unit := resourcedata.GetNillableValueFromMap[string](elementMap, "unit", false)
	if val == nil || unit == nil {
		return nil
	}
	elementOut := platformclientv2.Documentelementlength{
		Value: val,
		Unit:  unit,
	}
	return &elementOut
}

func buildDocumentTableRowBlocks(blocksIn map[string]interface{}, tableDepth int) *[]platformclientv2.Documentbodytablerowblock {
	blocksSlice, ok := blocksIn["rows"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodytablerowblock, 0, len(blocksSlice))

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		blockOut := platformclientv2.Documentbodytablerowblock{
			Properties: buildDocumentTableRowProperties(blockMap),
			Cells:      buildDocumentTableCellBlocks(blockMap, tableDepth),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return &blocksOut
}

func buildDocumentTableRowProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodytablerowblockproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}
	propertiesOut := platformclientv2.Documentbodytablerowblockproperties{
		RowType:         resourcedata.GetNillableValueFromMap[string](propertiesMap, "row_type", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Alignment:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "alignment", false),
		BorderStyle:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "border_style", false),
		BorderColor:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "border_color", false),
		Height:          nillableFloat32FromMap(propertiesMap, "height"),
	}
	return &propertiesOut
}

func buildDocumentTableCellBlocks(blocksIn map[string]interface{}, tableDepth int) *[]platformclientv2.Documentbodytablecellblock {
	blocksSlice, ok := blocksIn["cells"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodytablecellblock, 0, len(blocksSlice))

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		blockOut := platformclientv2.Documentbodytablecellblock{
			Properties: buildDocumentTableCellProperties(blockMap),
			Blocks:     buildDocumentTableContentBlocks(blockMap, tableDepth),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return &blocksOut
}

func buildDocumentTableCellProperties(propertiesIn map[string]interface{}) *platformclientv2.Documentbodytablecellblockproperties {
	propertiesSlice, ok := propertiesIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}
	propertiesOut := platformclientv2.Documentbodytablecellblockproperties{
		CellType:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "cell_type", false),
		HorizontalAlign: resourcedata.GetNillableValueFromMap[string](propertiesMap, "horizontal_align", false),
		VerticalAlign:   resourcedata.GetNillableValueFromMap[string](propertiesMap, "vertical_align", false),
		ColSpan:         nillableIntFromMap(propertiesMap, "col_span"),
		RowSpan:         nillableIntFromMap(propertiesMap, "row_span"),
		Height:          nillableFloat32FromMap(propertiesMap, "height"),
		Scope:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "scope", false),
		BorderWidth:     nillableFloat32FromMap(propertiesMap, "border_width"),
		BorderStyle:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "border_style", false),
		BorderColor:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "border_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Width:           nillableFloat32FromMap(propertiesMap, "width"),
		WidthWithUnit:   buildDocumentElement(propertiesMap, "width_with_unit"),
	}
	return &propertiesOut
}

func buildDocumentTableContentBlocks(blocksIn map[string]interface{}, tableDepth int) *[]platformclientv2.Documenttablecontentblock {
	blocksSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blocksSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documenttablecontentblock, 0, len(blocksSlice))

	for _, block := range blocksSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		blockOut := platformclientv2.Documenttablecontentblock{
			VarType:   resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Text:      buildDocumentText(blockMap),
			Image:     buildDocumentImage(blockMap),
			Video:     buildDocumentVideo(blockMap),
			List:      buildDocumentList(blockMap, 0),
			Paragraph: buildDocumentParagraph(blockMap),
		}
		if _, ok := blockMap["table"]; ok {
			if tableDepth+1 < maxTableDepth {
				blockOut.Table = buildDocumentTable(blockMap, tableDepth+1)
			} else {
				warnBuildTableTruncateOnce.Do(func() {
					log.Printf("[WARN] knowledge_document_variation: nested table exceeds maxTableDepth=%d; truncating at attemptedDepth=%d",
						maxTableDepth, tableDepth+1)
				})
			}
		}
		blocksOut = append(blocksOut, blockOut)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return &blocksOut
}

// Flatten Functions

func flattenVariationContextValue(valuesIn []platformclientv2.Knowledgecontextvaluereference) []interface{} {
	if len(valuesIn) == 0 {
		return nil
	}

	valuesOut := make([]interface{}, 0, len(valuesIn))

	for _, valueAny := range valuesIn {
		valueOut := make(map[string]interface{})

		if valueAny.Id != nil {
			valueOut["value_id"] = *valueAny.Id
		}

		if len(valueOut) == 0 {
			continue
		}
		valuesOut = append(valuesOut, valueOut)
	}

	if len(valuesOut) == 0 {
		return nil
	}
	return valuesOut
}

func flattenDocumentText(textIn platformclientv2.Documenttext) []interface{} {
	textOut := make(map[string]interface{})

	if textIn.Text != nil && *textIn.Text != "" {
		textOut["text"] = *textIn.Text
	}
	if textIn.Marks != nil {
		markSet := lists.StringListToSet(*textIn.Marks)
		textOut["marks"] = markSet
	}
	if textIn.Hyperlink != nil && *textIn.Hyperlink != "" {
		textOut["hyperlink"] = *textIn.Hyperlink
	}
	if textIn.Properties != nil {
		textOut["properties"] = flattenTextProperties(*textIn.Properties)
	}

	return []interface{}{textOut}
}

func flattenVariationContexts(contextsIn []platformclientv2.Documentvariationcontext) []interface{} {
	if len(contextsIn) == 0 {
		return nil
	}

	contextsOut := make([]interface{}, 0)

	for _, contextAny := range contextsIn {
		contextMap := make(map[string]interface{})

		if contextAny.Context != nil {
			contextMap["context"] = flattenVariationContext(*contextAny.Context)
		}
		if contextAny.Values != nil {
			contextMap["values"] = flattenVariationContextValue(*contextAny.Values)
		}

		if len(contextMap) == 0 {
			continue
		}
		contextsOut = append(contextsOut, contextMap)
	}

	if len(contextsOut) == 0 {
		return nil
	}
	return contextsOut
}

func flattenDocumentContentListBlocks(blocksIn []platformclientv2.Documentlistcontentblock, listDepth int) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "type", block.VarType)

		if block.Text != nil {
			blockOutMap["text"] = flattenDocumentText(*block.Text)
		}
		if block.Image != nil {
			blockOutMap["image"] = flattenDocumentImage(*block.Image)
		}
		if block.Video != nil {
			blockOutMap["video"] = flattenDocumentVideo(*block.Video)
		}
		if block.List != nil {
			if listDepth+1 < maxListDepth {
				blockOutMap["list"] = flattenDocumentList(*block.List, listDepth+1)
			} else {
				warnFlattenListTruncateOnce.Do(func() {
					log.Printf("[WARN] knowledge_document_variation: flatten nested list exceeds maxListDepth=%d; truncating at attemptedDepth=%d",
						maxListDepth, listDepth+1)
				})
			}
		}

		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

func flattenDocumentContentBlocks(blocksIn []platformclientv2.Documentcontentblock) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "type", block.VarType)

		if block.Text != nil {
			blockOutMap["text"] = flattenDocumentText(*block.Text)
		}
		if block.Image != nil {
			blockOutMap["image"] = flattenDocumentImage(*block.Image)
		}
		if block.Video != nil {
			blockOutMap["video"] = flattenDocumentVideo(*block.Video)
		}

		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

func flattenVariationContext(contextIn platformclientv2.Knowledgecontextreference) []interface{} {
	contextOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(contextOut, "context_id", contextIn.Id)
	return []interface{}{contextOut}
}

func flattenTextProperties(propertiesIn platformclientv2.Documenttextproperties) []interface{} {
	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "font_size", propertiesIn.FontSize)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "text_color", propertiesIn.TextColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentVideoProperties(propertiesIn *platformclientv2.Documentbodyvideoproperties) []interface{} {
	if propertiesIn == nil {
		return nil
	}

	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "align", propertiesIn.Align)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "indentation", propertiesIn.Indentation)

	if propertiesIn.Width != nil {
		propertiesOut["width"] = flattenDocumentElement(propertiesIn.Width)
	}
	if propertiesIn.Height != nil {
		propertiesOut["height"] = flattenDocumentElement(propertiesIn.Height)
	}

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentImageProperties(propertiesIn *platformclientv2.Documentbodyimageproperties) []interface{} {
	if propertiesIn == nil {
		return nil
	}
	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "align", propertiesIn.Align)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "indentation", propertiesIn.Indentation)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "width", propertiesIn.Width)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "alt_text", propertiesIn.AltText)
	if propertiesIn.WidthWithUnit != nil {
		propertiesOut["width_with_unit"] = flattenDocumentElement(propertiesIn.WidthWithUnit)
	}

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentListProperties(propertiesIn []platformclientv2.Documentbodylistblockproperties) []interface{} {
	if len(propertiesIn) == 0 {
		return nil
	}

	propertiesOut := make([]interface{}, 0)

	for _, property := range propertiesIn {
		propertyOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(propertyOutMap, "ordered_type", property.OrderedType)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "unordered_type", property.UnorderedType)

		if len(propertyOutMap) == 0 {
			continue
		}

		propertiesOut = append(propertiesOut, propertyOutMap)
	}

	if len(propertiesOut) == 0 {
		return nil
	}

	return propertiesOut
}

func flattenDocumentListBlockProperties(propertiesIn platformclientv2.Documentbodylistitemproperties) []interface{} {
	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "font_size", propertiesIn.FontSize)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "font_type", propertiesIn.FontType)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "text_color", propertiesIn.TextColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "align", propertiesIn.Align)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "indentation", propertiesIn.Indentation)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "ordered_type", propertiesIn.OrderedType)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "unordered_type", propertiesIn.UnorderedType)

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentListBlocks(blocksIn []platformclientv2.Documentbodylistblock, listDepth int) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "type", block.VarType)
		if block.Blocks != nil {
			blockOutMap["blocks"] = flattenDocumentContentListBlocks(*block.Blocks, listDepth)
		}
		if block.Properties != nil {
			blockOutMap["properties"] = flattenDocumentListBlockProperties(*block.Properties)
		}

		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

func flattenDocumentParagraph(paragraphIn platformclientv2.Documentbodyparagraph) []interface{} {
	paragraphOut := make(map[string]interface{})

	if paragraphIn.Blocks != nil {
		paragraphOut["blocks"] = flattenDocumentContentBlocks(*paragraphIn.Blocks)
	}
	if paragraphIn.Properties != nil {
		propertiesArray := []platformclientv2.Documentbodyparagraphproperties{*paragraphIn.Properties}
		paragraphOut["properties"] = flattenParagraphProperties(propertiesArray)
	}

	return []interface{}{paragraphOut}
}

func flattenParagraphProperties(propertiesIn []platformclientv2.Documentbodyparagraphproperties) []interface{} {
	if len(propertiesIn) == 0 {
		return nil
	}

	propertiesOut := make([]interface{}, 0)

	for _, property := range propertiesIn {
		propertyOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(propertyOutMap, "font_size", property.FontSize)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "font_type", property.FontType)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "text_color", property.TextColor)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "align", property.Align)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "background_color", property.BackgroundColor)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "indentation", property.Indentation)

		if len(propertyOutMap) == 0 {
			continue
		}
		propertiesOut = append(propertiesOut, propertyOutMap)
	}

	if len(propertiesOut) == 0 {
		return nil
	}
	return propertiesOut
}

func flattenDocumentImage(imageIn platformclientv2.Documentbodyimage) []interface{} {
	imageOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(imageOut, "url", imageIn.Url)
	resourcedata.SetMapValueIfNotNil(imageOut, "hyperlink", imageIn.Hyperlink)

	if imageIn.Properties != nil {
		imageOut["properties"] = flattenDocumentImageProperties(imageIn.Properties)
	}

	return []interface{}{imageOut}
}

func flattenDocumentVideo(videoIn platformclientv2.Documentbodyvideo) []interface{} {
	videoOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(videoOut, "url", videoIn.Url)
	if videoIn.Properties != nil {
		videoOut["properties"] = flattenDocumentVideoProperties(videoIn.Properties)
	}

	return []interface{}{videoOut}
}

func flattenDocumentList(listIn platformclientv2.Documentbodylist, listDepth int) []interface{} {
	listOut := make(map[string]interface{})

	if listIn.Blocks != nil {
		listOut["blocks"] = flattenDocumentListBlocks(*listIn.Blocks, listDepth)
	}
	if listIn.Properties != nil {
		listOut["properties"] = flattenDocumentListProperties([]platformclientv2.Documentbodylistblockproperties{*listIn.Properties})
	}

	return []interface{}{listOut}
}

func flattenDocumentBodyBlocks(blocksIn []platformclientv2.Documentbodyblock) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "type", block.VarType)
		if block.Paragraph != nil {
			blockOutMap["paragraph"] = flattenDocumentParagraph(*block.Paragraph)
		}
		if block.Image != nil {
			blockOutMap["image"] = flattenDocumentImage(*block.Image)
		}
		if block.Video != nil {
			blockOutMap["video"] = flattenDocumentVideo(*block.Video)
		}
		if block.List != nil {
			blockOutMap["list"] = flattenDocumentList(*block.List, 0)
		}
		if block.Table != nil {
			blockOutMap["table"] = flattenDocumentTable(*block.Table, 0)
		}
		blocksOut = append(blocksOut, blockOutMap)
	}

	return blocksOut
}

func flattenVariationBody(bodyIn platformclientv2.Documentbodyresponse) []interface{} {
	bodyOut := make(map[string]interface{})

	if bodyIn.Blocks != nil {
		bodyOut["blocks"] = flattenDocumentBodyBlocks(*bodyIn.Blocks)
	}

	return []interface{}{bodyOut}
}

func flattenDocumentVersion(versionIn platformclientv2.Addressableentityref) []interface{} {
	versionOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(versionOut, "id", versionIn.Id)
	return []interface{}{versionOut}
}

func flattenKnowledgeDocumentVariation(variationIn platformclientv2.Documentvariationresponse) []interface{} {
	variationOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(variationOut, "name", variationIn.Name)

	if variationIn.Body != nil {
		variationOut["body"] = flattenVariationBody(*variationIn.Body)
	}
	if variationIn.DocumentVersion != nil {
		variationOut["document_version"] = flattenDocumentVersion(*variationIn.DocumentVersion)
	}
	if variationIn.Contexts != nil {
		variationOut["contexts"] = flattenVariationContexts(*variationIn.Contexts)
	}
	if variationIn.Priority != nil {
		variationOut["priority"] = *variationIn.Priority
	}

	return []interface{}{variationOut}
}

func flattenDocumentTable(tableIn platformclientv2.Documentbodytable, tableDepth int) []interface{} {
	tableOut := make(map[string]interface{})

	if tableIn.Properties != nil {
		tableOut["properties"] = flattenDocumentTableProperties(*tableIn.Properties)
	}
	if tableIn.Rows != nil {
		tableOut["rows"] = flattenDocumentTableRowBlocks(*tableIn.Rows, tableDepth)
	}
	return []interface{}{tableOut}
}

func flattenDocumentTableProperties(propertiesIn platformclientv2.Documentbodytableproperties) []interface{} {
	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "width", propertiesIn.Width)
	if propertiesIn.WidthWithUnit != nil {
		propertiesOut["width_with_unit"] = flattenDocumentElement(propertiesIn.WidthWithUnit)
	}
	resourcedata.SetMapValueIfNotNil(propertiesOut, "cell_padding", propertiesIn.CellPadding)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_width", propertiesIn.BorderWidth)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_style", propertiesIn.BorderStyle)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_color", propertiesIn.BorderColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "alignment", propertiesIn.Alignment)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "height", propertiesIn.Height)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "cell_spacing", propertiesIn.CellSpacing)
	if propertiesIn.Caption != nil {
		propertiesOut["caption"] = flattenDocumentTableCaption(*propertiesIn.Caption)
	}

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentTableCaption(captionIn platformclientv2.Documentbodytablecaptionblock) []interface{} {
	captionOut := make(map[string]interface{})

	if captionIn.Blocks != nil {
		captionOut["blocks"] = flattenDocumentTableCaptionBlocks(*captionIn.Blocks)
	}
	return []interface{}{captionOut}
}

func flattenDocumentTableCaptionBlocks(blocksIn []platformclientv2.Documentbodytablecaptionitem) []interface{} {
	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "type", block.VarType)
		if block.Text != nil {
			blockOutMap["text"] = flattenDocumentText(*block.Text)
		}
		if block.Image != nil {
			blockOutMap["image"] = flattenDocumentImage(*block.Image)
		}
		if block.Video != nil {
			blockOutMap["video"] = flattenDocumentVideo(*block.Video)
		}
		if block.List != nil {
			blockOutMap["list"] = flattenDocumentList(*block.List, 0)
		}
		if block.Paragraph != nil {
			blockOutMap["paragraph"] = flattenDocumentParagraph(*block.Paragraph)
		}
		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

func flattenDocumentElement(elementIn *platformclientv2.Documentelementlength) []interface{} {
	if elementIn == nil || elementIn.Value == nil || elementIn.Unit == nil {
		return nil
	}
	elementOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(elementOut, "value", elementIn.Value)
	resourcedata.SetMapValueIfNotNil(elementOut, "unit", elementIn.Unit)
	return []interface{}{elementOut}
}

func flattenDocumentTableRowBlocks(blocksIn []platformclientv2.Documentbodytablerowblock, tableDepth int) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}
	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		if block.Properties != nil {
			blockOutMap["properties"] = flattenDocumentTableRowProperties(*block.Properties)
		}
		if block.Cells != nil {
			blockOutMap["cells"] = flattenDocumentTableCellBlocks(*block.Cells, tableDepth)
		}

		if len(blockOutMap) == 0 {
			continue
		}
		blocksOut = append(blocksOut, blockOutMap)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return blocksOut
}

func flattenDocumentTableRowProperties(propertiesIn platformclientv2.Documentbodytablerowblockproperties) []interface{} {
	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "row_type", propertiesIn.RowType)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "alignment", propertiesIn.Alignment)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_style", propertiesIn.BorderStyle)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_color", propertiesIn.BorderColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "height", propertiesIn.Height)

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentTableCellBlocks(blocksIn []platformclientv2.Documentbodytablecellblock, tableDepth int) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}
	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		if block.Properties != nil {
			blockOutMap["properties"] = flattenDocumentTableCellProperties(*block.Properties)
		}
		if block.Blocks != nil {
			blockOutMap["blocks"] = flattenDocumentTableContentBlocks(*block.Blocks, tableDepth)
		}

		if len(blockOutMap) == 0 {
			continue
		}
		blocksOut = append(blocksOut, blockOutMap)
	}

	if len(blocksOut) == 0 {
		return nil
	}
	return blocksOut
}

func flattenDocumentTableCellProperties(propertiesIn platformclientv2.Documentbodytablecellblockproperties) []interface{} {
	propertiesOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(propertiesOut, "cell_type", propertiesIn.CellType)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "horizontal_align", propertiesIn.HorizontalAlign)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "vertical_align", propertiesIn.VerticalAlign)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "col_span", propertiesIn.ColSpan)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "row_span", propertiesIn.RowSpan)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "height", propertiesIn.Height)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "scope", propertiesIn.Scope)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_width", propertiesIn.BorderWidth)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_style", propertiesIn.BorderStyle)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "border_color", propertiesIn.BorderColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "background_color", propertiesIn.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(propertiesOut, "width", propertiesIn.Width)
	if propertiesIn.WidthWithUnit != nil {
		propertiesOut["width_with_unit"] = flattenDocumentElement(propertiesIn.WidthWithUnit)
	}

	if len(propertiesOut) == 0 {
		return nil
	}
	return []interface{}{propertiesOut}
}

func flattenDocumentTableContentBlocks(blocksIn []platformclientv2.Documenttablecontentblock, tableDepth int) []interface{} {
	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		if block.VarType != nil {
			blockOutMap["type"] = *block.VarType
		}
		if block.Text != nil {
			blockOutMap["text"] = flattenDocumentText(*block.Text)
		}
		if block.Image != nil {
			blockOutMap["image"] = flattenDocumentImage(*block.Image)
		}
		if block.Video != nil {
			blockOutMap["video"] = flattenDocumentVideo(*block.Video)
		}
		if block.List != nil {
			blockOutMap["list"] = flattenDocumentList(*block.List, 0)
		}
		if block.Paragraph != nil {
			blockOutMap["paragraph"] = flattenDocumentParagraph(*block.Paragraph)
		}
		if block.Table != nil {
			if tableDepth+1 < maxTableDepth {
				blockOutMap["table"] = flattenDocumentTable(*block.Table, tableDepth+1)
			} else {
				warnFlattenTableTruncateOnce.Do(func() {
					log.Printf("[WARN] knowledge_document_variation: flatten nested table exceeds maxTableDepth=%d; truncating at attemptedDepth=%d",
						maxTableDepth, tableDepth+1)
				})
			}
		}
		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

// Utils

func nillableIntFromMap(m map[string]interface{}, key string) *int {
	if v := resourcedata.GetNillableValueFromMap[int](m, key, false); v != nil {
		return v
	}
	if vf := resourcedata.GetNillableValueFromMap[float64](m, key, false); vf != nil {
		i := int(*vf)
		return &i
	}
	if vs := resourcedata.GetNillableValueFromMap[string](m, key, false); vs != nil {
		if i, err := strconv.Atoi(*vs); err == nil {
			return &i
		}
	}
	return nil
}

func nillableFloat32FromMap(m map[string]interface{}, key string) *float32 {
	if v := resourcedata.GetNillableValueFromMap[float32](m, key, false); v != nil {
		return v
	}
	if vf := resourcedata.GetNillableValueFromMap[float64](m, key, false); vf != nil {
		f := float32(*vf)
		return &f
	}
	if vi := resourcedata.GetNillableValueFromMap[int](m, key, false); vi != nil {
		f := float32(*vi)
		return &f
	}
	return nil
}

func buildVariationId(baseID, documentID, variationID string) string {
	return baseID + variationIdSeparator + documentID + variationIdSeparator + variationID
}

func parseResourceIDs(id string) (*resourceIDs, error) {
	parts := strings.Split(id, variationIdSeparator)

	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid resource ID: %s", id)
	}

	return &resourceIDs{
		knowledgeDocumentVariationID:    parts[2],
		knowledgeBaseID:                 parts[0],
		knowledgeDocumentResourceDataID: parts[1],
		knowledgeDocumentID:             strings.Split(parts[1], ",")[0],
	}, nil
}

func getKnowledgeIdsFromResourceData(d *schema.ResourceData) resourceIDs {
	knowledgeBaseID, _ := d.Get("knowledge_base_id").(string)
	documentResourceId, _ := d.Get("knowledge_document_id").(string)
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]

	return resourceIDs{
		knowledgeBaseID:                 knowledgeBaseID,
		knowledgeDocumentResourceDataID: documentResourceId,
		knowledgeDocumentID:             knowledgeDocumentId,
	}
}
