package knowledge_document_variation

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func buildDocumentContentListBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentlistcontentblock {
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
		blocksOut = append(blocksOut, blockOut)
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

	return &blocksOut
}

func buildDocumentListBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodylistblock {
	blockSlice, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(blockSlice) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodylistblock, 0)

	for _, block := range blockSlice {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockOut := platformclientv2.Documentbodylistblock{
			VarType:    resourcedata.GetNillableValueFromMap[string](blockMap, "type", false),
			Blocks:     buildDocumentContentListBlocks(blockMap),
			Properties: buildDocumentListBlockProperties(blockMap),
		}
		blocksOut = append(blocksOut, blockOut)
	}

	return &blocksOut
}

func buildDocumentListBlockProperties(blocksIn map[string]interface{}) *platformclientv2.Documentbodylistitemproperties {
	propertiesSlice, ok := blocksIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Documentbodylistitemproperties{
		FontSize:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_size", false),
		FontType:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_type", false),
		TextColor:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "text_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
		Indentation:     resourcedata.GetNillableValueFromMap[float32](propertiesMap, "indentation", false),
		OrderedType:     resourcedata.GetNillableValueFromMap[string](propertiesMap, "ordered_type", false),
		UnorderedType:   resourcedata.GetNillableValueFromMap[string](propertiesMap, "unordered_type", false),
	}
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
		Text:      resourcedata.GetNillableValueFromMap[string](textMap, "text", false),
		Hyperlink: resourcedata.GetNillableValueFromMap[string](textMap, "hyperlink", false),
	}

	if marks, ok := textMap["marks"].(*schema.Set); ok && marks != nil {
		markArr := lists.SetToStringList(marks)
		textOut.Marks = markArr
	}

	if properties, ok := textMap["properties"].(map[string]interface{}); ok && properties != nil {
		textOut.Properties = buildTextProperties(properties)
	}

	return &textOut
}

func buildTextProperties(textIn map[string]interface{}) *platformclientv2.Documenttextproperties {
	propertiesSlice, ok := textIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Documenttextproperties{
		FontSize:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_size", false),
		TextColor:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "text_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
	}
}

func buildDocumentParagraph(paragraphIn map[string]any) *platformclientv2.Documentbodyparagraph {
	paragraphSlice, ok := paragraphIn["paragraph"].([]any)
	if !ok || len(paragraphSlice) == 0 {
		return nil
	}

	paragraphMap, ok := paragraphSlice[0].(map[string]any)
	if !ok {
		return nil
	}

	return &platformclientv2.Documentbodyparagraph{
		Blocks:     buildDocumentContentBlocks(paragraphMap),
		Properties: buildParagraphProperties(paragraphMap),
	}
}

func buildParagraphProperties(paragraphIn map[string]interface{}) *platformclientv2.Documentbodyparagraphproperties {
	propertiesSlice, ok := paragraphIn["properties"].([]any)
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]any)
	if !ok {
		return nil
	}

	paragraphPropertiesOut := platformclientv2.Documentbodyparagraphproperties{
		FontSize:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_size", false),
		FontType:        resourcedata.GetNillableValueFromMap[string](propertiesMap, "font_type", false),
		TextColor:       resourcedata.GetNillableValueFromMap[string](propertiesMap, "text_color", false),
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
	}

	if indentation, ok := propertiesMap["indentation"].(float64); ok {
		paragraphPropertiesOut.Indentation = platformclientv2.Float32(float32(indentation))
	}

	return &paragraphPropertiesOut
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

	return &platformclientv2.Documentbodyimage{
		Url:        resourcedata.GetNillableValueFromMap[string](imageMap, "url", false),
		Hyperlink:  resourcedata.GetNillableValueFromMap[string](imageMap, "hyperlink", false),
		Properties: buildDocumentImageProperties(imageMap),
	}
}

func buildDocumentVideoProperties(videoIn map[string]interface{}) *platformclientv2.Documentbodyvideoproperties {
	propertiesSlice, ok := videoIn["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	videoPropertiesOut := platformclientv2.Documentbodyvideoproperties{
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
	}

	if indentation, ok := propertiesMap["indentation"].(float64); ok {
		videoPropertiesOut.Indentation = platformclientv2.Float32(float32(indentation))
	}

	return &videoPropertiesOut
}

func buildDocumentImageProperties(imageMap map[string]interface{}) *platformclientv2.Documentbodyimageproperties {
	if imageMap == nil || len(imageMap) == 0 {
		return nil
	}

	propertiesSlice, ok := imageMap["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	imagePropertiesOut := platformclientv2.Documentbodyimageproperties{
		BackgroundColor: resourcedata.GetNillableValueFromMap[string](propertiesMap, "background_color", false),
		Align:           resourcedata.GetNillableValueFromMap[string](propertiesMap, "align", false),
	}

	if indentation, ok := propertiesMap["indentation"].(float64); ok {
		indentation32 := float32(indentation)
		imagePropertiesOut.Indentation = &indentation32
	}

	return &imagePropertiesOut
}

func buildDocumentVideo(videoIn map[string]interface{}) *platformclientv2.Documentbodyvideo {
	videoList, ok := videoIn["video"].([]interface{})
	if !ok || len(videoList) == 0 {
		return nil
	}

	videoMap, ok := videoList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Documentbodyvideo{
		Url:        resourcedata.GetNillableValueFromMap[string](videoMap, "url", false),
		Properties: buildDocumentVideoProperties(videoMap),
	}
}

func buildDocumentList(listIn map[string]any) *platformclientv2.Documentbodylist {
	listSlice, ok := listIn["list"].([]any)
	if !ok || len(listSlice) == 0 {
		return nil
	}

	listMap, ok := listSlice[0].(map[string]any)
	if !ok {
		return nil
	}

	listOut := platformclientv2.Documentbodylist{
		Blocks:     buildDocumentListBlocks(listMap),
		Properties: buildDocumentListProperties(listMap),
	}
	return &listOut
}

func buildDocumentListProperties(list map[string]interface{}) *platformclientv2.Documentbodylistblockproperties {
	propertiesSlice, ok := list["properties"].([]interface{})
	if !ok || len(propertiesSlice) == 0 {
		return nil
	}

	propertiesMap, ok := propertiesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	listPropertiesOut := platformclientv2.Documentbodylistblockproperties{
		OrderedType:   resourcedata.GetNillableValueFromMap[string](propertiesMap, "ordered_type", false),
		UnorderedType: resourcedata.GetNillableValueFromMap[string](propertiesMap, "unordered_type", false),
	}

	return &listPropertiesOut
}

func buildDocumentBodyBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodyblock {
	documentBodyBlocks, ok := blocksIn["blocks"].([]interface{})
	if !ok || len(documentBodyBlocks) == 0 {
		return nil
	}

	blocksOut := make([]platformclientv2.Documentbodyblock, 0)

	for _, block := range documentBodyBlocks {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockOut := platformclientv2.Documentbodyblock{
			VarType:   resourcedata.GetNillableValueFromMap[string](blockMap, "type", true),
			Paragraph: buildDocumentParagraph(blockMap),
			Image:     buildDocumentImage(blockMap),
			Video:     buildDocumentVideo(blockMap),
			List:      buildDocumentList(blockMap),
		}
		blocksOut = append(blocksOut, blockOut)
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

	return &platformclientv2.Documentbodyrequest{
		Blocks: buildDocumentBodyBlocks(bodyMap),
	}
}

func buildKnowledgeDocumentVariation(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	if variationIn == nil {
		return nil
	}

	return &platformclientv2.Documentvariationrequest{
		Name:     resourcedata.GetNillableValueFromMap[string](variationIn, "name", true),
		Body:     buildVariationBody(variationIn),
		Contexts: buildVariationContexts(variationIn),
	}
}

func buildVariationContexts(variationIn map[string]interface{}) *[]platformclientv2.Documentvariationcontext {
	contextsSlice, ok := variationIn["contexts"].([]interface{})
	if !ok || len(contextsSlice) == 0 {
		return nil
	}

	contextsMap, ok := contextsSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &[]platformclientv2.Documentvariationcontext{
		{
			Context: buildVariationContext(contextsMap),
			Values:  buildVariationContextValue(contextsMap),
		},
	}
}

func buildVariationContext(contextsIn map[string]interface{}) *platformclientv2.Knowledgecontextreference {
	contextSlice, ok := contextsIn["context"].([]interface{})
	if !ok || len(contextSlice) == 0 {
		return nil
	}

	contextMap, ok := contextSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Knowledgecontextreference{
		Id: resourcedata.GetNillableValueFromMap[string](contextMap, "context_id", true),
	}
}

func buildVariationContextValue(contextsIn map[string]interface{}) *[]platformclientv2.Knowledgecontextvaluereference {
	valuesSlice, ok := contextsIn["values"].([]interface{})
	if !ok || len(valuesSlice) == 0 {
		return nil
	}

	valueMap, ok := valuesSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &[]platformclientv2.Knowledgecontextvaluereference{
		{
			Id: resourcedata.GetNillableValueFromMap[string](valueMap, "value_id", true),
		},
	}
}

func buildKnowledgeDocumentVariationUpdate(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	return &platformclientv2.Documentvariationrequest{
		Name: resourcedata.GetNillableValueFromMap[string](variationIn, "name", true),
		Body: buildVariationBody(variationIn),
	}
}

// Flatten Functions

func flattenVariationContextValue(valIn []platformclientv2.Knowledgecontextvaluereference) []interface{} {
	if len(valIn) == 0 {
		return nil
	}

	valOut := make(map[string]interface{})
	if valIn[0].Id != nil {
		valId := *valIn[0].Id
		valOut["value_id"] = valId
	}
	return []interface{}{valOut}
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

func flattenVariationContexts(contextIn []platformclientv2.Documentvariationcontext) []interface{} {
	if len(contextIn) == 0 {
		return nil
	}

	contextsOut := make([]interface{}, 0)

	for _, context := range contextIn {
		contextMap := make(map[string]interface{})

		if context.Context != nil {
			contextMap["context"] = flattenVariationContext(*context.Context)
		}
		if context.Values != nil {
			contextMap["values"] = flattenVariationContextValue(*context.Values)
		}
		contextsOut = append(contextsOut, contextMap)
	}
	return contextsOut
}

func flattenDocumentContentListBlocks(blocksIn []platformclientv2.Documentlistcontentblock) []interface{} {
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

func flattenTextProperties(textIn platformclientv2.Documenttextproperties) []interface{} {
	properties := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(properties, "font_size", textIn.FontSize)
	resourcedata.SetMapValueIfNotNil(properties, "text_color", textIn.TextColor)
	resourcedata.SetMapValueIfNotNil(properties, "background_color", textIn.BackgroundColor)

	return []interface{}{properties}
}

func flattenDocumentVideoProperties(property *platformclientv2.Documentbodyvideoproperties) []interface{} {
	if property == nil {
		return nil
	}

	videoProperties := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(videoProperties, "align", property.Align)
	resourcedata.SetMapValueIfNotNil(videoProperties, "background_color", property.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(videoProperties, "indentation", property.Indentation)

	return []interface{}{videoProperties}
}

func flattenDocumentImageProperties(propertiesIn []platformclientv2.Documentbodyimageproperties) []interface{} {
	if len(propertiesIn) == 0 {
		return nil
	}

	propertiesOut := make([]interface{}, 0)
	for _, property := range propertiesIn {
		propertyOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(propertyOutMap, "align", property.Align)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "background_color", property.BackgroundColor)
		resourcedata.SetMapValueIfNotNil(propertyOutMap, "indentation", property.Indentation)

		propertiesOut = append(propertiesOut, propertyOutMap)
	}
	return propertiesOut
}

func flattenDocumentListProperties(blocksIn []platformclientv2.Documentbodylistblockproperties) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)
	for _, property := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "ordered_type", property.OrderedType)
		resourcedata.SetMapValueIfNotNil(blockOutMap, "unordered_type", property.UnorderedType)

		blocksOut = append(blocksOut, blockOutMap)
	}

	return blocksOut
}

func flattenDocumentListBlockProperties(listBlockProperties platformclientv2.Documentbodylistitemproperties) []interface{} {
	properties := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(properties, "font_size", listBlockProperties.FontSize)
	resourcedata.SetMapValueIfNotNil(properties, "font_type", listBlockProperties.FontType)
	resourcedata.SetMapValueIfNotNil(properties, "text_color", listBlockProperties.TextColor)
	resourcedata.SetMapValueIfNotNil(properties, "align", listBlockProperties.Align)
	resourcedata.SetMapValueIfNotNil(properties, "background_color", listBlockProperties.BackgroundColor)
	resourcedata.SetMapValueIfNotNil(properties, "indentation", listBlockProperties.Indentation)
	resourcedata.SetMapValueIfNotNil(properties, "ordered_type", listBlockProperties.OrderedType)
	resourcedata.SetMapValueIfNotNil(properties, "unordered_type", listBlockProperties.UnorderedType)

	return []interface{}{properties}
}

func flattenDocumentListBlocks(blocksIn []platformclientv2.Documentbodylistblock) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(blockOutMap, "type", block.VarType)
		if block.Blocks != nil {
			blockOutMap["blocks"] = flattenDocumentContentListBlocks(*block.Blocks)
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

		propertiesOut = append(propertiesOut, propertyOutMap)
	}
	return propertiesOut
}

func flattenDocumentImage(imageIn platformclientv2.Documentbodyimage) []interface{} {
	imageOut := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(imageOut, "url", imageIn.Url)
	resourcedata.SetMapValueIfNotNil(imageOut, "hyperlink", imageIn.Hyperlink)

	if imageIn.Properties != nil {
		propertiesArray := []platformclientv2.Documentbodyimageproperties{*imageIn.Properties}
		imageOut["properties"] = flattenDocumentImageProperties(propertiesArray)
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

func flattenDocumentList(listIn platformclientv2.Documentbodylist) []interface{} {
	listOut := make(map[string]interface{})

	if listIn.Blocks != nil {
		listOut["blocks"] = flattenDocumentListBlocks(*listIn.Blocks)
	}
	if listIn.Properties != nil {
		propertiesArray := []platformclientv2.Documentbodylistblockproperties{*listIn.Properties}
		listOut["properties"] = flattenDocumentListProperties(propertiesArray)
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
			blockOutMap["list"] = flattenDocumentList(*block.List)
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

	return []interface{}{variationOut}
}

// Utils

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
