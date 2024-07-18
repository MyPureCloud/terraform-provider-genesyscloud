package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	knowledgeDocumentV1 = &schema.Resource{
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

func getAllKnowledgeDocumentsV1(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	documentEntities := make([]platformclientv2.Knowledgedocument, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	// get published knowledge bases
	publishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, true)
	if err != nil {
		return nil, err
	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, false)
	if err != nil {
		return nil, err
	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		partialEntities, err := getAllKnowledgeV1DocumentEntities(*knowledgeAPI, &knowledgeBase)
		if err != nil {
			return nil, err
		}
		documentEntities = append(documentEntities, *partialEntities...)
	}

	for _, knowledgeDocument := range documentEntities {
		id := fmt.Sprintf("%s %s %s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.LanguageCode)
		var name string
		if knowledgeDocument.Name != nil {
			name = *knowledgeDocument.Name
		} else {
			name = fmt.Sprintf("document " + uuid.NewString())
		}
		resources[id] = &resourceExporter.ResourceMeta{Name: name}
	}

	return resources, nil
}

func getAllKnowledgeV1DocumentEntities(knowledgeAPI platformclientv2.KnowledgeApi, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Knowledgedocument, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Knowledgedocument
	)

	const pageSize = 100
	for {
		knowledgeDocuments, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocuments(*knowledgeBase.Id, *knowledgeBase.CoreLanguage, "", after, "", fmt.Sprintf("%v", pageSize), "", "", "", "", nil)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to get page of knowledge documents error: %s", getErr), resp)
		}

		if knowledgeDocuments.Entities == nil || len(*knowledgeDocuments.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeDocuments.Entities...)

		if knowledgeDocuments.NextUri == nil || *knowledgeDocuments.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeDocuments.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to parse after cursor from knowledge document nextUri"), err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}

func KnowledgeDocumentExporterV1() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeDocumentsV1),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

func ResourceKnowledgeDocumentV1() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge document",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeDocumentV1),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeDocumentV1),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeDocumentV1),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeDocumentV1),
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
				Description:      "Language code",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLanguageCode,
			},
			"knowledge_document": {
				Description: "Knowledge document request body",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeDocumentV1,
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
		faqOut.Alternatives = lists.SetToStringList(alternatives)
	}
	return &faqOut
}

func buildCategories(requestBody map[string]interface{}, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string, languageCode string) (*[]platformclientv2.Documentcategoryinput, diag.Diagnostics) {
	if requestBody["categories"] == nil {
		return nil, nil
	}

	categories := make([]platformclientv2.Documentcategoryinput, 0)

	categoryList := lists.SetToStringList(requestBody["categories"].(*schema.Set))
	for _, categoryName := range *categoryList {
		pageSize := 100
		knowledgeCategories, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategories(knowledgeBaseId, languageCode, "", "", "", fmt.Sprintf("%v", pageSize), categoryName)

		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to get page of knowledge categories error: %s", getErr), resp)
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
		articleOut.Alternatives = lists.SetToStringList(alternatives)
	}

	return &articleOut
}

func buildKnowledgeDocumentRequestV1(d *schema.ResourceData, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string, languageCode string) platformclientv2.Knowledgedocumentrequest {
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

func flattenKnowledgeDocumentV1(document *platformclientv2.Knowledgedocument) []interface{} {
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
		faqMap["alternatives"] = lists.StringListToSet(*faq.Alternatives)
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
		articleMap["alternatives"] = lists.StringListToSet(*article.Alternatives)
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

	return lists.StringListToSet(categoryList)
}

func createKnowledgeDocumentV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	languageCode := d.Get("language_code").(string)
	body := buildKnowledgeDocumentRequestV1(d, knowledgeAPI, knowledgeBaseId, languageCode)

	log.Printf("Creating knowledge document")
	knowledgeDocument, resp, err := knowledgeAPI.PostKnowledgeKnowledgebaseLanguageDocuments(knowledgeBaseId, languageCode, body)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to create knowledge document %s error: %s", d.Id(), err), resp)
	}

	id := fmt.Sprintf("%s %s %s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.LanguageCode)
	d.SetId(id)

	log.Printf("Created knowledge document %s", *knowledgeDocument.Id)
	return readKnowledgeDocumentV1(ctx, d, meta)
}

func readKnowledgeDocumentV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocument(), constants.DefaultConsistencyChecks, "genesyscloud_knowledge_v1_document")

	log.Printf("Reading knowledge document %s", knowledgeDocumentId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeDocument, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to read knowledge document %s | error: %s", knowledgeDocumentId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to read knowledge document %s | error: %s", knowledgeDocumentId, getErr), resp))
		}

		// required
		newId := fmt.Sprintf("%s %s %s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id, *knowledgeDocument.LanguageCode)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeDocument.KnowledgeBase.Id)
		d.Set("language_code", *knowledgeDocument.LanguageCode)
		d.Set("knowledge_document", flattenKnowledgeDocumentV1(knowledgeDocument))

		log.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		return cc.CheckState(d)
	})
}

func updateKnowledgeDocumentV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating Knowledge document %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Knowledge document version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to read knowledge document %s error: %s", knowledgeDocumentId, getErr), resp)
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
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to update knowledge document %s error: %s", knowledgeDocumentId, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Knowledge document %s", knowledgeDocumentId)
	return readKnowledgeDocumentV1(ctx, d, meta)
}

func deleteKnowledgeDocumentV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Knowledge document %s", knowledgeDocumentId)
	_, resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Failed to delete knowledge document %s error: %s", knowledgeDocumentId, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, languageCode)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge document deleted
				log.Printf("Deleted Knowledge document %s", knowledgeDocumentId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Error deleting Knowledge document %s | error: %s", knowledgeDocumentId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_document", fmt.Sprintf("Knowledge document %s still exists", knowledgeDocumentId), resp))
	})
}
