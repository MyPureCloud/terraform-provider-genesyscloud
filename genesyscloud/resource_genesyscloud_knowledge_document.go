package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

var (
	knowledgeDocument = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Document type according to assigned template",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Faq", "Article"}, false),
			},
			"external_url": {
				Description: "External Url to the document",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"faq": {
				Description: "Faq document details",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        documentFaq,
			},
			"categories": {
				Description: "List of knowledge base category names for the document",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"article": {
				Description: "Article details",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        documentArticle,
			},
		},
	}
	documentFaq = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"question": {
				Description: "The question for this FAQ",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"answer": {
				Description: "The answer for this FAQ",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"alternatives": {
				Description: "List of Alternative questions related to the answer which helps in improving the likelihood of a match to user query",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
	documentArticle = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"title": {
				Description: "The title of the Article.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"content_location_url": {
				Description: "Presigned URL to retrieve the document content.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"alternatives": {
				Description: "List of Alternative questions related to the title which helps in improving the likelihood of a match to user query.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

func getAllKnowledgeDocuments(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	resources := make(ResourceIDMetaMap)
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		knowledgeBases, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), "", "", false, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of knowledge bases: %v", getErr)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}
		for _, knowledgeBase := range *knowledgeBases.Entities {
			knowledgeBaseList = append(knowledgeBaseList, knowledgeBase)
		}
	}
	for _, knowledgeBase := range knowledgeBaseList {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			knowledgeDocuments, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocuments(*knowledgeBase.Id, *knowledgeBase.CoreLanguage, "", "", "", fmt.Sprintf("%v", pageSize), "", "", "", "", nil)
			if getErr != nil {
				return nil, diag.Errorf("Failed to get page of Knowledge documents: %v", getErr)
			}

			if knowledgeDocuments.Entities == nil || len(*knowledgeDocuments.Entities) == 0 {
				break
			}
			for _, knowledgeDocument := range *knowledgeDocuments.Entities {
				id := fmt.Sprintf("%s %s %s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.LanguageCode)
				resources[id] = &ResourceMeta{Name: *knowledgeDocument.Name}
			}
		}
	}

	return resources, nil
}

func knowledgeDocumentExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllKnowledgeDocuments),
		RefAttrs: map[string]*RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

func resourceKnowledgeDocument() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge document",

		CreateContext: createWithPooledClient(createKnowledgeDocument),
		ReadContext:   readWithPooledClient(readKnowledgeDocument),
		UpdateContext: updateWithPooledClient(updateKnowledgeDocument),
		DeleteContext: deleteWithPooledClient(deleteKnowledgeDocument),
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
			"language_code": {
				Description:  "Language code",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"en-US", "en-UK", "en-AU", "de-DE", "es-US", "es-ES", "fr-FR", "pt-BR", "nl-NL", "it-IT", "fr-CA"}, false),
			},
			"knowledge_document": {
				Description: "Knowledge document request body",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeDocument,
			},
		},
	}
}

func buildFaq(requestBody map[string]interface{}) *platformclientv2.Documentfaq {
	faqIn := requestBody["faq"]

	if faqIn == nil || len(faqIn.([]interface{})) <= 0 {
		return nil
	}

	faqOut := platformclientv2.Documentfaq{}
	temp := faqIn.([]interface{})[0].(map[string]interface{})
	if question, ok := temp["question"].(string); ok && question != "" {
		faqOut.Question = &question
	}
	if answer, ok := temp["answer"].(string); ok && answer != "" {
		faqOut.Answer = &answer
	}
	if alternatives, ok := temp["alternatives"].(*schema.Set); ok {
		faqOut.Alternatives = setToStringList(alternatives)
	}
	return &faqOut
}

func buildCategories(requestBody map[string]interface{}, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string, languageCode string) (*[]platformclientv2.Documentcategoryinput, diag.Diagnostics) {
	if requestBody["categories"] == nil {
		return nil, nil
	}

	categories := make([]platformclientv2.Documentcategoryinput, 0)

	categoryList := setToStringList(requestBody["categories"].(*schema.Set))
	for _, categoryName := range *categoryList {
		pageSize := 100
		knowledgeCategories, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategories(knowledgeBaseId, languageCode, "", "", "", fmt.Sprintf("%v", pageSize), categoryName)

		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of knowledge categories: %v", getErr)
		}

		matchingCategory := (*knowledgeCategories.Entities)[0]
		category := platformclientv2.Documentcategoryinput{
			Id: matchingCategory.Id,
		}

		categories = append(categories, category)
	}

	return &categories, nil
}

func buildArticle(requestBody map[string]interface{}) *platformclientv2.Documentarticle {
	articleIn := requestBody["article"]
	if articleIn == nil || len(articleIn.([]interface{})) <= 0 {
		return nil
	}

	articleOut := platformclientv2.Documentarticle{}

	temp := articleIn.([]interface{})[0].(map[string]interface{})
	if title, ok := temp["title"].(string); ok && title != "" {
		articleOut.Title = &title
	}
	if contentLocationUrl, ok := temp["content_location_url"].(string); ok {
		articleContentBody := platformclientv2.Articlecontentbody{
			LocationUrl: &contentLocationUrl,
		}
		articleContent := platformclientv2.Articlecontent{
			Body: &articleContentBody,
		}
		articleOut.Content = &articleContent
	}
	if alternatives, ok := temp["alternatives"].(*schema.Set); ok {
		articleOut.Alternatives = setToStringList(alternatives)
	}

	return &articleOut
}

func buildKnowledgeDocumentRequest(d *schema.ResourceData, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string, languageCode string) platformclientv2.Knowledgedocumentrequest {
	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	categories, _ := buildCategories(requestIn, knowledgeAPI, knowledgeBaseId, languageCode)
	requestOut := platformclientv2.Knowledgedocumentrequest{
		Faq:        buildFaq(requestIn),
		Categories: categories,
		Article:    buildArticle(requestIn),
	}

	if externalUrl, ok := requestIn["external_url"].(string); ok && externalUrl != "" {
		requestOut.ExternalUrl = &externalUrl
	}
	if varType, ok := requestIn["type"].(string); ok && varType != "" {
		requestOut.VarType = &varType
	}

	return requestOut
}

func flattenKnowledgeDocument(document *platformclientv2.Knowledgedocument) []interface{} {
	if document == nil {
		return nil
	}

	documentMap := make(map[string]interface{})

	documentMap["categories"] = flattenCategories(document.Categories)
	documentMap["faq"] = flattenFaq(document.Faq)
	documentMap["article"] = flattenArticle(document.Article)

	if document.VarType != nil {
		documentMap["type"] = *document.VarType
	}
	if document.ExternalUrl != nil {
		documentMap["external_url"] = *document.ExternalUrl
	}

	return []interface{}{documentMap}
}

func flattenFaq(faq *platformclientv2.Documentfaq) []interface{} {
	if faq == nil {
		return nil
	}

	faqMap := make(map[string]interface{})
	if faq.Question != nil {
		faqMap["question"] = faq.Question
	}
	if faq.Answer != nil {
		faqMap["answer"] = faq.Answer
	}
	if faq.Alternatives != nil {
		faqMap["alternatives"] = stringListToSet(*faq.Alternatives)
	}

	return []interface{}{faqMap}
}

func flattenArticle(article *platformclientv2.Documentarticle) []interface{} {
	if article == nil {
		return nil
	}

	articleMap := make(map[string]interface{})
	if article.Title != nil {
		articleMap["title"] = article.Title
	}
	if article.Content != nil && article.Content.Body != nil && article.Content.Body.LocationUrl != nil {
		articleMap["content_location_url"] = article.Content.Body.LocationUrl
	}
	if article.Alternatives != nil {
		articleMap["alternatives"] = stringListToSet(*article.Alternatives)
	}

	return []interface{}{articleMap}
}

func flattenCategories(categories *[]platformclientv2.Knowledgecategory) *schema.Set {
	if categories == nil {
		return nil
	}

	categoryList := make([]string, 0)

	for _, category := range *categories {
		if category.Id != nil {
			categoryList = append(categoryList, *category.Name)
		}
	}

	return stringListToSet(categoryList)
}

func createKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	languageCode := d.Get("language_code").(string)
	body := buildKnowledgeDocumentRequest(d, knowledgeAPI, knowledgeBaseId, languageCode)

	log.Printf("Creating knowledge document")
	knowledgeDocument, _, err := knowledgeAPI.PostKnowledgeKnowledgebaseLanguageDocuments(knowledgeBaseId, languageCode, body)
	if err != nil {
		return diag.Errorf("Failed to create knowledge document: %s", err)
	}

	id := fmt.Sprintf("%s %s %s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.LanguageCode)
	d.SetId(id)

	log.Printf("Created knowledge document %s", *knowledgeDocument.Id)
	return readKnowledgeDocument(ctx, d, meta)
}

func readKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Reading knowledge document %s", knowledgeDocumentId)
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		knowledgeDocument, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.NonRetryableError(fmt.Errorf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceKnowledgeDocument())

		// required
		newId := fmt.Sprintf("%s %s %s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.LanguageCode)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeDocument.KnowledgeBase.Id)
		d.Set("language_code", *knowledgeDocument.LanguageCode)
		d.Set("knowledge_document", flattenKnowledgeDocument(knowledgeDocument))

		log.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		checkState := cc.CheckState()
		return checkState
	})
}

func updateKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating Knowledge document %s", d.Id())
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Knowledge document version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Knowledge document %s: %s", knowledgeDocumentId, getErr)
		}

		body := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
		update := platformclientv2.Knowledgedocumentrequest{
			Faq:     buildFaq(body),
			Article: buildArticle(body),
		}

		categories, _ := buildCategories(body, knowledgeAPI, knowledgeBaseId, languageCode)
		if categories != nil {
			update.Categories = categories
		}
		if varType, ok := body["type"].(string); ok {
			update.VarType = &varType
		}
		if externalUrl, ok := body["external_url"].(string); ok {
			update.ExternalUrl = &externalUrl
		}

		log.Printf("Updating knowledge document %s", knowledgeDocumentId)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode, update)
		if putErr != nil {
			return resp, diag.Errorf("Failed to update knowledge document %s: %s", knowledgeDocumentId, putErr)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Knowledge document %s", knowledgeDocumentId)
	return readKnowledgeDocument(ctx, d, meta)
}

func deleteKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Knowledge document %s", knowledgeCategoryId)
	_, _, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseLanguageDocument(knowledgeCategoryId, knowledgeBaseId, languageCode)
	if err != nil {
		return diag.Errorf("Failed to delete Knowledge document %s: %s", knowledgeCategoryId, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if err != nil {
			if isStatus404(resp) {
				// Knowledge document deleted
				log.Printf("Deleted Knowledge document %s", knowledgeCategoryId)
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Knowledge document %s: %s", knowledgeCategoryId, err))
		}

		return resource.RetryableError(fmt.Errorf("Knowledge document %s still exists", knowledgeCategoryId))
	})
}
