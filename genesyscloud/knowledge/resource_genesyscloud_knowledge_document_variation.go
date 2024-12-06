package knowledge

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

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

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

func getAllKnowledgeDocumentVariations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)

	knowledgeProxy := knowledgeDocument.GetKnowledgeDocumentProxy(clientConfig)
	knowledgeApi := knowledgeProxy.KnowledgeApi
	// get published knowledge bases
	publishedEntities, response, err := knowledgeProxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("%v", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, response, err := knowledgeProxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("%v", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		variationEntities, response, err := knowledgeProxy.GetAllKnowledgeDocumentEntities(ctx, &knowledgeBase)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("%v", err), response)
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
			knowledgeDocumentVariations, resp, getErr := knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariations(*knowledgeBase.Id, *knowledgeDocument.Id, "", "", fmt.Sprintf("%v", pageSize), documentState)
			if getErr != nil {
				return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to get page of knowledge document variations error: %v", err), resp)
			}

			if knowledgeDocumentVariations.Entities == nil || len(*knowledgeDocumentVariations.Entities) == 0 {
				break
			}

			for _, knowledgeDocumentVariation := range *knowledgeDocumentVariations.Entities {
				id := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariation.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.Id)
				resources[id] = &resourceExporter.ResourceMeta{BlockLabel: "variation " + uuid.NewString()}
			}
		}
	}

	return resources, nil
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

func createKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	documentResourceId := d.Get("knowledge_document_id").(string)
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]
	knowledgeDocumentVariation := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})
	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeDocumentVariationRequest := buildKnowledgeDocumentVariation(knowledgeDocumentVariation)

	log.Printf("Creating knowledge document variation for document %s", knowledgeDocumentId)

	knowledgeDocumentVariationResponse, resp, err := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVariations(knowledgeBaseId, knowledgeDocumentId, *knowledgeDocumentVariationRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to create variation for knowledge document %s error: %s", d.Id(), err), resp)
	}

	if published == true {
		_, resp, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, platformclientv2.Knowledgedocumentversion{})

		if versionErr != nil {
			return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
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
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocumentVariation(), constants.ConsistencyChecks(), "genesyscloud_knowledge_document_variation")

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
			publishedVariation, resp, publishedErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published")

			if publishedErr != nil {
				// Published version may or may not exist, so if status is 404, sleep and retry once and then move on to retrieve draft variation.
				if util.IsStatus404(resp) {
					time.Sleep(2 * time.Second)
					retryVariation, retryResp, retryErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published")

					if retryErr != nil {
						if !util.IsStatus404(retryResp) {
							log.Printf("%s", retryErr)
							return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, retryErr), retryResp))
						}
					} else {
						publishedVariation = retryVariation
					}

				} else {
					log.Printf("%s", publishedErr)
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, publishedErr), resp))
				}
			}

			draftVariation, resp, draftErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft")
			if draftErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, draftErr), resp))
				}
				log.Printf("%s", draftErr)
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, draftErr), resp))
			}

			if publishedVariation != nil && publishedVariation.DateModified != nil && publishedVariation.DateModified.After(*draftVariation.DateModified) {
				knowledgeDocumentVariation = publishedVariation
			} else {
				knowledgeDocumentVariation = draftVariation
			}
		} else {
			variation, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, documentState)
			if getErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s | error: %s", documentVariationId, getErr), resp))
				}
				log.Printf("%s", getErr)
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s | error: %s", documentVariationId, getErr), resp))
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
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft")
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to read knowledge document variation %s error: %s", id, getErr), resp)
		}

		knowledgeDocumentVariationUpdate := buildKnowledgeDocumentVariationUpdate(knowledgeDocumentVariation)

		log.Printf("Updating knowledge document variation %s", documentVariationId)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, *knowledgeDocumentVariationUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to update knowledge document variation %s error: %s", documentVariationId, putErr), resp)
		}
		if published == true {
			_, resp, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to publish knowledge document %s error: %s", id, versionErr), resp)
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
	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	log.Printf("Deleting knowledge document variation %s", id)
	resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to delete knowledge document variation %s error: %s", id, err), resp)
	}

	if published == true {
		/*
		 * If the published flag is set, attempt to publish a new document version without the variation.
		 * However a document cannot be published if it has no variations, so first check that the document has other variations
		 * A new document version can only be published if there are other variations than the one being removed
		 */
		pageSize := 3
		variations, resp, variationErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariations(knowledgeBaseId, knowledgeDocumentId, "", "", fmt.Sprintf("%v", pageSize), "Draft")

		if variationErr != nil {
			return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to retrieve knowledge document variations error: %s", err), resp)
		}

		if len(*variations.Entities) > 0 {
			_, resp, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
			}
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		// The DELETE resource for knowledge document variations only removes draft variations. So set the documentState param to "Draft" for the check
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft")
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge base deleted
				log.Printf("Deleted knowledge document variation %s", documentVariationId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Error deleting knowledge document variation %s | error: %s", documentVariationId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document_variation", fmt.Sprintf("Knowledge document variation %s still exists", documentVariationId), resp))
	})
}

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
	variationOut := platformclientv2.Documentvariationrequest{
		Body: buildVariationBody(variationIn),
	}
	return &variationOut
}

func buildKnowledgeDocumentVariationUpdate(variationIn map[string]interface{}) *platformclientv2.Documentvariationrequest {
	variationOut := platformclientv2.Documentvariationrequest{
		Body: buildVariationBody(variationIn),
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

	return []interface{}{variationOut}
}
