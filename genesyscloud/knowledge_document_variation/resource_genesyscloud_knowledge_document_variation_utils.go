package knowledge_document_variation

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func buildDocumentContentListBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentlistcontentblock {
	if documentContentBlocks, _ := blocksIn["blocks"].([]interface{}); len(documentContentBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentlistcontentblock, 0)

		for _, block := range documentContentBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)

			blockOut := platformclientv2.Documentlistcontentblock{
				VarType: &varType,
				Text:    buildDocumentText(blockMap),
				Image:   buildDocumentImage(blockMap),
				Video:   buildDocumentVideo(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildDocumentContentBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentcontentblock {
	if documentContentBlocks, _ := blocksIn["blocks"].([]interface{}); len(documentContentBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentcontentblock, 0)

		for _, block := range documentContentBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)

			blockOut := platformclientv2.Documentcontentblock{
				VarType: &varType,
				Text:    buildDocumentText(blockMap),
				Image:   buildDocumentImage(blockMap),
				Video:   buildDocumentVideo(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildDocumentListBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodylistblock {
	if documentListBlocks, _ := blocksIn["blocks"].([]interface{}); len(documentListBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentbodylistblock, 0)

		for _, block := range documentListBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)

			blockOut := platformclientv2.Documentbodylistblock{
				VarType:    &varType,
				Blocks:     buildDocumentContentListBlocks(blockMap),
				Properties: buildDocumentListBlockProperties(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildDocumentListBlockProperties(blocksIn map[string]interface{}) *platformclientv2.Documentbodylistitemproperties {
	if listBlockProperties, _ := blocksIn["properties"].([]interface{}); len(listBlockProperties) > 0 {
		properties := listBlockProperties[0].(map[string]interface{})
		listBlockPropertiesOut := platformclientv2.Documentbodylistitemproperties{}

		if fontSize, ok := properties["font_size"].(string); ok && fontSize != "" {
			listBlockPropertiesOut.FontSize = &fontSize
		}
		if fontType, ok := properties["font_type"].(string); ok && fontType != "" {
			listBlockPropertiesOut.FontType = &fontType
		}
		if textColor, ok := properties["text_color"].(string); ok && textColor != "" {
			listBlockPropertiesOut.TextColor = &textColor
		}
		if backgroundColor, ok := properties["background_color"].(string); ok && backgroundColor != "" {
			listBlockPropertiesOut.BackgroundColor = &backgroundColor
		}
		if align, ok := properties["align"].(string); ok && align != "" {
			listBlockPropertiesOut.Align = &align
		}
		if indentation, ok := properties["indentation"].(float32); ok {
			listBlockPropertiesOut.Indentation = &indentation
		}
		if orderedType, ok := properties["ordered_type"].(string); ok && orderedType != "" {
			listBlockPropertiesOut.OrderedType = &orderedType
		}
		if unorderedType, ok := properties["unordered_type"].(string); ok && unorderedType != "" {
			listBlockPropertiesOut.UnorderedType = &unorderedType
		}

		return &listBlockPropertiesOut
	}
	return nil
}

func buildDocumentText(textIn map[string]interface{}) *platformclientv2.Documenttext {
	if textList, _ := textIn["text"].([]interface{}); len(textList) > 0 {
		text := textList[0].(map[string]interface{})
		textOut := platformclientv2.Documenttext{}

		if textString, ok := text["text"].(string); ok && textString != "" {
			textOut.Text = &textString
		}
		if marks, ok := text["marks"].(*schema.Set); ok && marks != nil {
			markArr := lists.SetToStringList(marks)
			textOut.Marks = markArr
		}
		if hyperlink, ok := text["hyperlink"].(string); ok && hyperlink != "" {
			textOut.Hyperlink = &hyperlink
		}
		if properties, ok := text["properties"].(map[string]interface{}); ok && properties != nil {
			textOut.Properties = buildTextProperties(properties)
		}

		return &textOut
	}
	return nil
}

func buildTextProperties(textIn map[string]interface{}) *platformclientv2.Documenttextproperties {
	if paragraphProperties, _ := textIn["properties"].([]interface{}); len(paragraphProperties) > 0 {
		properties := paragraphProperties[0].(map[string]interface{})
		textPropertiesOut := platformclientv2.Documenttextproperties{}

		if fontSize, ok := properties["font_size"].(string); ok && fontSize != "" {
			textPropertiesOut.FontSize = &fontSize
		}
		if textColor, ok := properties["text_color"].(string); ok && textColor != "" {
			textPropertiesOut.TextColor = &textColor
		}
		if backgroundColor, ok := properties["background_color"].(string); ok && backgroundColor != "" {
			textPropertiesOut.BackgroundColor = &backgroundColor
		}

		return &textPropertiesOut
	}
	return nil
}

func buildDocumentParagraph(paragraphIn map[string]interface{}) *platformclientv2.Documentbodyparagraph {
	if paragraphList, _ := paragraphIn["paragraph"].([]interface{}); len(paragraphList) > 0 {
		paragraph := paragraphList[0].(map[string]interface{})

		return &platformclientv2.Documentbodyparagraph{
			Blocks:     buildDocumentContentBlocks(paragraph),
			Properties: buildParagraphProperties(paragraph),
		}
	}
	return nil
}

func buildParagraphProperties(paragraphIn map[string]interface{}) *platformclientv2.Documentbodyparagraphproperties {
	if paragraphProperties, _ := paragraphIn["properties"].([]interface{}); len(paragraphProperties) > 0 {
		properties := paragraphProperties[0].(map[string]interface{})
		paragraphPropertiesOut := platformclientv2.Documentbodyparagraphproperties{}

		if fontSize, ok := properties["font_size"].(string); ok && fontSize != "" {
			paragraphPropertiesOut.FontSize = &fontSize
		}
		if fontType, ok := properties["font_type"].(string); ok && fontType != "" {
			paragraphPropertiesOut.FontType = &fontType
		}
		if textColor, ok := properties["text_color"].(string); ok && textColor != "" {
			paragraphPropertiesOut.TextColor = &textColor
		}
		if backgroundColor, ok := properties["background_color"].(string); ok && backgroundColor != "" {
			paragraphPropertiesOut.BackgroundColor = &backgroundColor
		}
		if align, ok := properties["align"].(string); ok && align != "" {
			paragraphPropertiesOut.Align = &align
		}
		if indentation, ok := properties["indentation"].(float64); ok {
			indentation := float32(indentation)
			paragraphPropertiesOut.Indentation = &indentation
		}

		return &paragraphPropertiesOut
	}
	return nil
}

func buildDocumentImage(imageIn map[string]interface{}) *platformclientv2.Documentbodyimage {
	if imageList, _ := imageIn["image"].([]interface{}); len(imageList) > 0 {
		image := imageList[0].(map[string]interface{})

		url := image["url"].(string)
		imageOut := platformclientv2.Documentbodyimage{
			Url:        &url,
			Properties: buildDocumentImageProperties(image),
		}

		if hyperlink, ok := image["hyperlink"].(string); ok && hyperlink != "" {
			imageOut.Hyperlink = &hyperlink
		}
		return &imageOut
	}
	return nil
}

func buildDocumentVideoProperties(videoIn map[string]interface{}) *platformclientv2.Documentbodyvideoproperties {
	if videoProperties, _ := videoIn["properties"].([]interface{}); len(videoProperties) > 0 {
		properties := videoProperties[0].(map[string]interface{})
		videoPropertiesOut := platformclientv2.Documentbodyvideoproperties{}

		if backgroundColor, ok := properties["background_color"].(string); ok && backgroundColor != "" {
			videoPropertiesOut.BackgroundColor = &backgroundColor
		}
		if align, ok := properties["align"].(string); ok && align != "" {
			videoPropertiesOut.Align = &align
		}
		if indentation, ok := properties["indentation"].(float64); ok {
			indentation := float32(indentation)
			videoPropertiesOut.Indentation = &indentation
		}

		return &videoPropertiesOut
	}
	return nil
}

func buildDocumentImageProperties(imageIn map[string]interface{}) *platformclientv2.Documentbodyimageproperties {
	if imageProperties, _ := imageIn["properties"].([]interface{}); len(imageProperties) > 0 {
		properties := imageProperties[0].(map[string]interface{})
		imagePropertiesOut := platformclientv2.Documentbodyimageproperties{}

		if backgroundColor, ok := properties["background_color"].(string); ok && backgroundColor != "" {
			imagePropertiesOut.BackgroundColor = &backgroundColor
		}
		if align, ok := properties["align"].(string); ok && align != "" {
			imagePropertiesOut.Align = &align
		}
		if indentation, ok := properties["indentation"].(float64); ok {
			indentation := float32(indentation)
			imagePropertiesOut.Indentation = &indentation
		}

		return &imagePropertiesOut
	}
	return nil
}

func buildDocumentVideo(videoIn map[string]interface{}) *platformclientv2.Documentbodyvideo {
	if videoList, _ := videoIn["video"].([]interface{}); len(videoList) > 0 {
		video := videoList[0].(map[string]interface{})

		url := video["url"].(string)

		return &platformclientv2.Documentbodyvideo{
			Url:        &url,
			Properties: buildDocumentVideoProperties(video),
		}
	}
	return nil
}

func buildDocumentList(listIn map[string]interface{}) *platformclientv2.Documentbodylist {
	if listList, _ := listIn["list"].([]interface{}); len(listList) > 0 {
		list := listList[0].(map[string]interface{})

		listOut := platformclientv2.Documentbodylist{
			Blocks:     buildDocumentListBlocks(list),
			Properties: buildDocumentListProperties(list),
		}
		return &listOut
	}
	return nil
}

func buildDocumentListProperties(list map[string]interface{}) *platformclientv2.Documentbodylistblockproperties {
	if listProperties, _ := list["properties"].([]interface{}); len(listProperties) > 0 {
		properties := listProperties[0].(map[string]interface{})
		listPropertiesOut := platformclientv2.Documentbodylistblockproperties{}

		if orderedType, ok := properties["ordered_type"].(string); ok && orderedType != "" {
			listPropertiesOut.OrderedType = &orderedType
		}
		if unorderedType, ok := properties["unordered_type"].(string); ok && unorderedType != "" {
			listPropertiesOut.UnorderedType = &unorderedType
		}

		return &listPropertiesOut
	}
	return nil
}

func buildDocumentBodyBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodyblock {
	if documentBodyBlocks, _ := blocksIn["blocks"].([]interface{}); len(documentBodyBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentbodyblock, 0)

		for _, block := range documentBodyBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)
			blockOut := platformclientv2.Documentbodyblock{
				VarType:   &varType,
				Paragraph: buildDocumentParagraph(blockMap),
				Image:     buildDocumentImage(blockMap),
				Video:     buildDocumentVideo(blockMap),
				List:      buildDocumentList(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildVariationBody(bodyIn map[string]interface{}) *platformclientv2.Documentbodyrequest {
	if bodyList, _ := bodyIn["body"].([]interface{}); len(bodyList) > 0 {
		variationBody := bodyList[0].(map[string]interface{})

		bodyOut := platformclientv2.Documentbodyrequest{
			Blocks: buildDocumentBodyBlocks(variationBody),
		}
		return &bodyOut
	}
	return nil
}

func buildKnowledgeDocumentVariation(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	if variationIn != nil {
		variationName := variationIn["name"].(string)

		return &platformclientv2.Documentvariationrequest{
			Name:     &variationName,
			Body:     buildVariationBody(variationIn),
			Contexts: buildVariationContexts(variationIn),
		}
	}
	return nil
}

func buildVariationContexts(variationIn map[string]interface{}) *[]platformclientv2.Documentvariationcontext {
	if contextsList, _ := variationIn["contexts"].([]interface{}); len(contextsList) > 0 {
		contexts := contextsList[0].(map[string]interface{})

		return &[]platformclientv2.Documentvariationcontext{
			{
				Context: buildVariationContext(contexts),
				Values:  buildVariationContextValue(contexts),
			},
		}
	}
	return nil
}

func buildVariationContext(contextsIn map[string]interface{}) *platformclientv2.Knowledgecontextreference {
	if context, _ := contextsIn["context"].([]interface{}); len(context) > 0 {
		context := context[0].(map[string]interface{})
		contextId := context["context_id"].(string)

		return &platformclientv2.Knowledgecontextreference{
			Id: &contextId,
		}
	}
	return nil
}

func buildVariationContextValue(contextsIn map[string]interface{}) *[]platformclientv2.Knowledgecontextvaluereference {
	if values, _ := contextsIn["values"].([]interface{}); len(values) > 0 {
		value := values[0].(map[string]interface{})
		valueId := value["value_id"].(string)

		return &[]platformclientv2.Knowledgecontextvaluereference{
			{
				Id: &valueId,
			},
		}
	}
	return nil
}

func buildKnowledgeDocumentVariationUpdate(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	name := variationIn["name"].(string)

	variationOut := platformclientv2.Documentvariationrequest{
		Name: &name,
		Body: buildVariationBody(variationIn),
	}

	return &variationOut
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

		if block.VarType != nil {
			blockOutMap["type"] = *block.VarType
		}
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

		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

func flattenVariationContext(contextIn platformclientv2.Knowledgecontextreference) []interface{} {
	contextOut := make(map[string]interface{})

	if contextIn.Id != nil {
		contextOut["context_id"] = *contextIn.Id
	}

	return []interface{}{contextOut}
}

func flattenTextProperties(textIn platformclientv2.Documenttextproperties) []interface{} {
	properties := make(map[string]interface{})

	if textIn.FontSize != nil && *textIn.FontSize != "" {
		properties["font_size"] = textIn.FontSize
	}
	if textIn.TextColor != nil && *textIn.TextColor != "" {
		properties["text_color"] = textIn.TextColor
	}
	if textIn.BackgroundColor != nil && *textIn.BackgroundColor != "" {
		properties["background_color"] = textIn.BackgroundColor
	}

	return []interface{}{properties}
}

func flattenDocumentVideoProperties(property *platformclientv2.Documentbodyvideoproperties) []interface{} {
	if property == nil {
		return nil
	}

	videoProperties := make(map[string]interface{})

	if property.Align != nil && *property.Align != "" {
		videoProperties["align"] = property.Align
	}
	if property.BackgroundColor != nil && *property.BackgroundColor != "" {
		videoProperties["background_color"] = property.BackgroundColor
	}
	if property.Indentation != nil {
		videoProperties["indentation"] = property.Indentation
	}

	return []interface{}{videoProperties}
}

func flattenDocumentImageProperties(propertiesIn []platformclientv2.Documentbodyimageproperties) []interface{} {
	if len(propertiesIn) == 0 {
		return nil
	}

	propertiesOut := make([]interface{}, 0)
	for _, property := range propertiesIn {
		propertyOutMap := make(map[string]interface{})

		if property.Align != nil && *property.Align != "" {
			propertyOutMap["align"] = property.Align
		}
		if property.BackgroundColor != nil && *property.BackgroundColor != "" {
			propertyOutMap["background_color"] = property.BackgroundColor
		}
		if property.Indentation != nil {
			propertyOutMap["indentation"] = property.Indentation
		}

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
		if property.OrderedType != nil && *property.OrderedType != "" {
			blockOutMap["ordered_type"] = *property.OrderedType
		}
		if property.UnorderedType != nil && *property.UnorderedType != "" {
			blockOutMap["unordered_type"] = *property.UnorderedType
		}

		blocksOut = append(blocksOut, blockOutMap)
	}
	return blocksOut
}

func flattenDocumentListBlockProperties(listBlockProperties platformclientv2.Documentbodylistitemproperties) []interface{} {
	properties := make(map[string]interface{})

	if listBlockProperties.FontSize != nil && *listBlockProperties.FontSize != "" {
		properties["font_size"] = listBlockProperties.FontSize
	}
	if listBlockProperties.FontType != nil && *listBlockProperties.FontType != "" {
		properties["font_type"] = listBlockProperties.FontType
	}
	if listBlockProperties.TextColor != nil && *listBlockProperties.TextColor != "" {
		properties["text_color"] = listBlockProperties.TextColor
	}
	if listBlockProperties.Align != nil && *listBlockProperties.Align != "" {
		properties["align"] = listBlockProperties.Align
	}
	if listBlockProperties.BackgroundColor != nil && *listBlockProperties.BackgroundColor != "" {
		properties["background_color"] = listBlockProperties.BackgroundColor
	}
	if listBlockProperties.Indentation != nil {
		properties["indentation"] = listBlockProperties.Indentation
	}
	if listBlockProperties.OrderedType != nil && *listBlockProperties.OrderedType == "" {
		properties["ordered_type"] = *listBlockProperties.OrderedType
	}
	if listBlockProperties.UnorderedType != nil && *listBlockProperties.UnorderedType == "" {
		properties["unordered_type"] = *listBlockProperties.UnorderedType
	}
	return []interface{}{properties}
}

func flattenDocumentListBlocks(blocksIn []platformclientv2.Documentbodylistblock) []interface{} {
	if len(blocksIn) == 0 {
		return nil
	}

	blocksOut := make([]interface{}, 0)

	for _, block := range blocksIn {
		blockOutMap := make(map[string]interface{})

		if block.VarType != nil {
			blockOutMap["type"] = *block.VarType
		}
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

		if property.FontSize != nil && *property.FontSize != "" {
			propertyOutMap["font_size"] = property.FontSize
		}
		if property.FontType != nil && *property.FontType != "" {
			propertyOutMap["font_type"] = property.FontType
		}
		if property.TextColor != nil && *property.TextColor != "" {
			propertyOutMap["text_color"] = property.TextColor
		}
		if property.Align != nil && *property.Align != "" {
			propertyOutMap["align"] = property.Align
		}
		if property.BackgroundColor != nil && *property.BackgroundColor != "" {
			propertyOutMap["background_color"] = property.BackgroundColor
		}
		if property.Indentation != nil {
			propertyOutMap["indentation"] = property.Indentation
		}

		propertiesOut = append(propertiesOut, propertyOutMap)
	}
	return propertiesOut
}

func flattenDocumentImage(imageIn platformclientv2.Documentbodyimage) []interface{} {
	imageOut := make(map[string]interface{})

	if imageIn.Url != nil && *imageIn.Url != "" {
		imageOut["url"] = *imageIn.Url
	}
	if imageIn.Hyperlink != nil && *imageIn.Hyperlink != "" {
		imageOut["hyperlink"] = *imageIn.Hyperlink
	}
	if imageIn.Properties != nil {
		propertiesArray := []platformclientv2.Documentbodyimageproperties{*imageIn.Properties}
		imageOut["properties"] = flattenDocumentImageProperties(propertiesArray)
	}

	return []interface{}{imageOut}
}

func flattenDocumentVideo(videoIn platformclientv2.Documentbodyvideo) []interface{} {
	videoOut := make(map[string]interface{})

	if videoIn.Url != nil && *videoIn.Url != "" {
		videoOut["url"] = *videoIn.Url
	}
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

		if block.VarType != nil {
			blockOutMap["type"] = *block.VarType
		}
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

	if versionIn.Id != nil {
		versionOut["id"] = *versionIn.Id
	}

	return []interface{}{versionOut}
}

func flattenKnowledgeDocumentVariation(variationIn platformclientv2.Documentvariationresponse) []interface{} {
	variationOut := make(map[string]interface{})

	if variationIn.Body != nil {
		variationOut["body"] = flattenVariationBody(*variationIn.Body)
	}
	if variationIn.DocumentVersion != nil {
		variationOut["document_version"] = flattenDocumentVersion(*variationIn.DocumentVersion)
	}
	if variationIn.Name != nil {
		variationOut["name"] = *variationIn.Name
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
