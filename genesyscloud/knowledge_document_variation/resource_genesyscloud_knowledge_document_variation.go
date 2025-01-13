package knowledgedocumentvariation

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	knowledgeDocument "terraform-provider-genesyscloud/genesyscloud/knowledge_document"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func getAllKnowledgeDocumentVariations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)

	knowledgeProxy := knowledgeDocument.GetKnowledgeDocumentProxy(clientConfig)
	knowledgeApi := knowledgeProxy.KnowledgeApi
	// get published knowledge bases
	publishedEntities, response, err := knowledgeProxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, response, err := knowledgeProxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		variationEntities, response, err := knowledgeProxy.GetAllKnowledgeDocumentEntities(ctx, &knowledgeBase)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", err), response)
		}

		// retrieve the documents for each knowledge base
		for _, knowledgeDocument := range *variationEntities {
			const pageSize = 100

			// parse document state
			var documentState string
			isValidState := strings.EqualFold(*knowledgeDocument.State, "Published") || strings.EqualFold(*knowledgeDocument.State, "Draft")
			if isValidState {
				documentState = *knowledgeDocument.State
			}

			// get the variations for each document
			knowledgeDocumentVariations, resp, getErr := knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariations(*knowledgeBase.Id, *knowledgeDocument.Id, "", "", fmt.Sprintf("%v", pageSize), documentState, nil)
			if getErr != nil {
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge document variations error: %v", err), resp)
			}

			if knowledgeDocumentVariations.Entities == nil || len(*knowledgeDocumentVariations.Entities) == 0 {
				break
			}

			for _, knowledgeDocumentVariation := range *knowledgeDocumentVariations.Entities {
				id := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariation.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.Id)
				blockLabel := *knowledgeBase.Name + "_" + *knowledgeDocument.Title
				if knowledgeDocumentVariation.Name != nil && *knowledgeDocumentVariation.Name != "" {
					blockLabel = blockLabel + "_" + *knowledgeDocumentVariation.Name
				} else {
					blockLabel = blockLabel + "_" + *knowledgeDocumentVariation.Id
				}
				resources[id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
			}
		}
	}

	return resources, nil
}

func createKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeProxy := getVariationRequestProxy(sdkConfig)

	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	documentResourceId := d.Get("knowledge_document_id").(string)
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]
	knowledgeDocumentVariation := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	knowledgeDocumentVariationRequest := buildKnowledgeDocumentVariation(knowledgeDocumentVariation)

	log.Printf("Creating knowledge document variation for document %s", knowledgeDocumentId)

	knowledgeDocumentVariationResponse, resp, err := knowledgeProxy.CreateVariation(ctx, knowledgeDocumentVariationRequest, knowledgeBaseId, knowledgeDocumentId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create variation for knowledge document %s error: %s", d.Id(), err), resp)
	}

	if published == true {
		_, resp, versionErr := knowledgeProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, knowledgeDocumentId, knowledgeBaseId, &platformclientv2.Knowledgedocumentversion{})
		if versionErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
		}
	}

	id := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariationResponse.Id, knowledgeBaseId, documentResourceId)
	d.SetId(id)

	log.Printf("Created knowledge document variation %s", *knowledgeDocumentVariationResponse.Id)
	return readKnowledgeDocumentVariation(ctx, d, meta)
}

func readKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocumentVariation(), constants.ConsistencyChecks(), ResourceType)

	documentState := ""

	// If the published flag is set, use it to set documentState param
	if published, ok := d.GetOk("published"); ok {
		if published == true {
			documentState = "Published"
		} else {
			documentState = "Draft"
		}
	}

	log.Printf("Reading knowledge document variation %s", documentVariationId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		var knowledgeDocumentVariation *platformclientv2.Documentvariationresponse
		/*
		 * If the published flag is not set, get both published and draft variation and choose the most recent
		 * If it is set, base the document state param off the flag.
		 * The published flag has to be optional for the import case, where only the resource ID is available.
		 * If the published flag were required, it would cause consistency issues for the import state.
		 */
		if documentState == "" {
			publishedVariation, resp, publishedErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published", nil)

			if publishedErr != nil {
				// Published version may or may not exist, so if status is 404, sleep and retry once and then move on to retrieve draft variation.
				if util.IsStatus404(resp) {
					time.Sleep(2 * time.Second)
					retryVariation, retryResp, retryErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published", nil)

					if retryErr != nil {
						if !util.IsStatus404(retryResp) {
							log.Printf("%s", retryErr)
							return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, retryErr), retryResp))
						}
					} else {
						publishedVariation = retryVariation
					}

				} else {
					log.Printf("%s", publishedErr)
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, publishedErr), resp))
				}
			}

			draftVariation, resp, draftErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)
			if draftErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, draftErr), resp))
				}
				log.Printf("%s", draftErr)
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, draftErr), resp))
			}

			if publishedVariation != nil && publishedVariation.DateModified != nil && publishedVariation.DateModified.After(*draftVariation.DateModified) {
				knowledgeDocumentVariation = publishedVariation
			} else {
				knowledgeDocumentVariation = draftVariation
			}
		} else {
			variation, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, documentState, nil)
			if getErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s | error: %s", documentVariationId, getErr), resp))
				}
				log.Printf("%s", getErr)
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s | error: %s", documentVariationId, getErr), resp))
			}

			knowledgeDocumentVariation = variation
		}

		newId := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariation.Id, *knowledgeDocumentVariation.Document.KnowledgeBase.Id, documentResourceId)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeDocumentVariation.Document.KnowledgeBase.Id)
		d.Set("knowledge_document_id", documentResourceId)
		d.Set("knowledge_document_variation", flattenKnowledgeDocumentVariation(*knowledgeDocumentVariation))

		if knowledgeDocumentVariation.DocumentVersion != nil && knowledgeDocumentVariation.DocumentVersion.Id != nil && len(*knowledgeDocumentVariation.DocumentVersion.Id) > 0 {
			d.Set("published", true)
		} else {
			d.Set("published", false)
		}

		log.Printf("Read knowledge document variation %s", documentVariationId)

		return cc.CheckState(d)
	})
}

func updateKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]
	knowledgeDocumentVariation := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})
	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge document variation %s", documentVariationId)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge document variation version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s error: %s", id, getErr), resp)
		}

		knowledgeDocumentVariationUpdate := buildKnowledgeDocumentVariationUpdate(knowledgeDocumentVariation)

		log.Printf("Updating knowledge document variation %s", documentVariationId)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, *knowledgeDocumentVariationUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update knowledge document variation %s error: %s", documentVariationId, putErr), resp)
		}
		if published == true {
			_, resp, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document %s error: %s", id, versionErr), resp)
			}
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge document variation %s", documentVariationId)
	return readKnowledgeDocumentVariation(ctx, d, meta)
}

func deleteKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeProxy := getVariationRequestProxy(sdkConfig)

	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	log.Printf("Deleting knowledge document variation %s", id)

	resp, err := knowledgeProxy.deleteVariationRequest(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete knowledge document variation %s error: %s", id, err), resp)
	}

	if published == true {
		/*
		 * If the published flag is set, attempt to publish a new document version without the variation.
		 * However a document cannot be published if it has no variations, so first check that the document has other variations
		 * A new document version can only be published if there are other variations than the one being removed
		 */
		pageSize := 3

		//variations, resp, variationErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariations(knowledgeBaseId, knowledgeDocumentId, "", "", fmt.Sprintf("%v", pageSize), "Draft", nil)
		variations, resp, variationErr := knowledgeProxy.getAllVariations(ctx)
		if variationErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve knowledge document variations error: %s", err), resp)
		}

		if len(*variations.Entities) > 0 {
			_, resp, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
			}
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		// The DELETE resource for knowledge document variations only removes draft variations. So set the documentState param to "Draft" for the check
		_, resp, err := knowledgeProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted knowledge document variation %s", documentVariationId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting knowledge document variation %s | error: %s", documentVariationId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Knowledge document variation %s still exists", documentVariationId), resp))
	})
}

// UTILS

func buildDocumentContentListBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentlistcontentblock {
	if documentContentBlocks := blocksIn["blocks"].([]interface{}); documentContentBlocks != nil && len(documentContentBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentlistcontentblock, 0)
		for _, block := range documentContentBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)
			blockOut := platformclientv2.Documentlistcontentblock{
				VarType: &varType,
				Text:    buildDocumentText(blockMap),
				Image:   buildDocumentImage(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildDocumentContentBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentcontentblock {
	if documentContentBlocks := blocksIn["blocks"].([]interface{}); documentContentBlocks != nil && len(documentContentBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentcontentblock, 0)
		for _, block := range documentContentBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)
			blockOut := platformclientv2.Documentcontentblock{
				VarType: &varType,
				Text:    buildDocumentText(blockMap),
				Image:   buildDocumentImage(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildDocumentListBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodylistblock {
	if documentListBlocks := blocksIn["blocks"].([]interface{}); documentListBlocks != nil && len(documentListBlocks) > 0 {
		blocksOut := make([]platformclientv2.Documentbodylistblock, 0)
		for _, block := range documentListBlocks {
			blockMap := block.(map[string]interface{})
			varType := blockMap["type"].(string)
			blockOut := platformclientv2.Documentbodylistblock{
				VarType: &varType,
				Blocks:  buildDocumentContentListBlocks(blockMap),
			}
			blocksOut = append(blocksOut, blockOut)
		}
		return &blocksOut
	}
	return nil
}

func buildDocumentText(textIn map[string]interface{}) *platformclientv2.Documenttext {
	if textList := textIn["text"].([]interface{}); textList != nil && len(textList) > 0 {
		text := textList[0].(map[string]interface{})
		textString := text["text"].(string)
		textOut := platformclientv2.Documenttext{
			Text: &textString,
		}
		if marks, ok := text["marks"].(*schema.Set); ok {
			markArr := lists.SetToStringList(marks)
			textOut.Marks = markArr
		}
		if hyperlink, ok := text["hyperlink"].(string); ok {
			if len(hyperlink) > 0 {
				textOut.Hyperlink = &hyperlink
			}
		}
		return &textOut
	}
	return nil
}

func buildDocumentParagraph(paragraphIn map[string]interface{}) *platformclientv2.Documentbodyparagraph {
	if paragraphList := paragraphIn["paragraph"].([]interface{}); paragraphList != nil && len(paragraphList) > 0 {
		paragraph := paragraphList[0].(map[string]interface{})
		paragraphOut := platformclientv2.Documentbodyparagraph{
			Blocks: buildDocumentContentBlocks(paragraph),
		}

		return &paragraphOut
	}
	return nil
}

func buildDocumentImage(imageIn map[string]interface{}) *platformclientv2.Documentbodyimage {
	if imageList := imageIn["image"].([]interface{}); imageList != nil && len(imageList) > 0 {
		image := imageList[0].(map[string]interface{})
		url := image["url"].(string)
		imageOut := platformclientv2.Documentbodyimage{
			Url: &url,
		}
		if hyperlink, ok := image["hyperlink"].(string); ok {
			if len(hyperlink) > 0 {
				imageOut.Hyperlink = &hyperlink
			}
		}
		return &imageOut
	}
	return nil
}

func buildDocumentVideo(videoIn map[string]interface{}) *platformclientv2.Documentbodyvideo {
	if videoList := videoIn["video"].([]interface{}); videoList != nil && len(videoList) > 0 {
		video := videoList[0].(map[string]interface{})
		url := video["url"].(string)
		videoOut := platformclientv2.Documentbodyvideo{
			Url: &url,
		}
		return &videoOut
	}
	return nil
}

func buildDocumentList(listIn map[string]interface{}) *platformclientv2.Documentbodylist {
	if listList := listIn["list"].([]interface{}); listList != nil && len(listList) > 0 {
		list := listList[0].(map[string]interface{})
		listOut := platformclientv2.Documentbodylist{
			Blocks: buildDocumentListBlocks(list),
		}

		return &listOut
	}
	return nil
}

func buildDocumentBodyBlocks(blocksIn map[string]interface{}) *[]platformclientv2.Documentbodyblock {
	if documentBodyBlocks := blocksIn["blocks"].([]interface{}); documentBodyBlocks != nil && len(documentBodyBlocks) > 0 {
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
	if bodyList := bodyIn["body"].([]interface{}); bodyList != nil && len(bodyList) > 0 {
		variationBody := bodyList[0].(map[string]interface{})
		bodyOut := platformclientv2.Documentbodyrequest{
			Blocks: buildDocumentBodyBlocks(variationBody),
		}
		return &bodyOut
	}
	return nil
}

func buildKnowledgeDocumentVariation(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	name := variationIn["name"].(string)
	variationOut := platformclientv2.Documentvariationrequest{
		Body: buildVariationBody(variationIn),
		Name: &name,
		//Contexts: buildVariationContexts(variationIn),
	}
	return &variationOut
}

// func buildVariationContexts(variationIn map[string]interface{}) *[]platformclientv2.Documentvariationcontext {
// 	if contextsList := variationIn["contexts"].([]interface{}); contextsList != nil && len(contextsList) > 0 {
// 		contexts := contextsList[0].(map[string]interface{})

// 		return &[]platformclientv2.Documentvariationcontext{
// 			{
// 				Context: buildVariationContext(contexts),
// 				Values:  buildVariationContextValue(contexts),
// 			},
// 		}
// 	}
// 	return nil
// }

// func buildVariationContext(contexts map[string]interface{}) *platformclientv2.Knowledgecontextreference {
// 	if context := contexts["context"].(string); context != nil && len(context) > 0{

// 		return &platformclientv2.Knowledgecontextreference{
// 			Id: &context,
// 		}
// 	}
// 	return nil
// }

// func buildVariationContextValue(contexts map[string]interface{}) *[]platformclientv2.Knowledgecontextvaluereference {
// 	if contextId := contexts["context_id"].(string)
// 	contextOut := platformclientv2.Knowledgecontextvaluereference{
// 		Id: &contextId,
// 	}
// 	return &[]platformclientv2.Knowledgecontextvaluereference{contextOut}
// }

// func flattenVariationContexts() {

// }

// func flattenVariationContext() {

// }

// func flattenVariationContextValue() {

// }

func buildKnowledgeDocumentVariationUpdate(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	name := variationIn["name"].(string)

	variationOut := platformclientv2.Documentvariationrequest{
		Body: buildVariationBody(variationIn),
		Name: &name,
	}

	return &variationOut
}

func flattenDocumentText(textIn platformclientv2.Documenttext) []interface{} {
	textOut := make(map[string]interface{})

	if textIn.Text != nil {
		textOut["text"] = *textIn.Text
	}
	if textIn.Marks != nil {
		markSet := lists.StringListToSet(*textIn.Marks)
		textOut["marks"] = markSet
	}
	if textIn.Hyperlink != nil && len(*textIn.Hyperlink) > 0 {
		textOut["hyperlink"] = *textIn.Hyperlink
	}

	return []interface{}{textOut}
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

		blocksOut = append(blocksOut, blockOutMap)
	}

	return blocksOut
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
		blocksOut = append(blocksOut, blockOutMap)
	}

	return blocksOut
}

func flattenDocumentParagraph(paragraphIn platformclientv2.Documentbodyparagraph) []interface{} {
	paragraphOut := make(map[string]interface{})

	if paragraphIn.Blocks != nil {
		paragraphOut["blocks"] = flattenDocumentContentBlocks(*paragraphIn.Blocks)
	}

	return []interface{}{paragraphOut}
}

func flattenDocumentImage(imageIn platformclientv2.Documentbodyimage) []interface{} {
	imageOut := make(map[string]interface{})

	if imageIn.Url != nil {
		imageOut["url"] = *imageIn.Url
	}
	if imageIn.Hyperlink != nil && len(*imageIn.Hyperlink) > 0 {
		imageOut["hyperlink"] = *imageIn.Hyperlink
	}

	return []interface{}{imageOut}
}

func flattenDocumentVideo(imageIn platformclientv2.Documentbodyvideo) []interface{} {
	imageOut := make(map[string]interface{})

	if imageIn.Url != nil {
		imageOut["url"] = *imageIn.Url
	}

	return []interface{}{imageOut}
}

func flattenDocumentList(listIn platformclientv2.Documentbodylist) []interface{} {
	listOut := make(map[string]interface{})

	if listIn.Blocks != nil {
		listOut["blocks"] = flattenDocumentListBlocks(*listIn.Blocks)
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

	return []interface{}{variationOut}
}
