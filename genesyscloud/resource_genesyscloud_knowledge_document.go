package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v94/platformclientv2"
)

var (
	knowledgeDocument = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"title": {
				Description: "Document title",
				Type:        schema.TypeString,
				Required:    true,
			},
			"visible": {
				Description: "Indicates if the knowledge document should be included in search results.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"alternatives": {
				Description: "List of alternate phrases related to the title which improves search results.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        documentAlternative,
			},
			"category_name": {
				Description: "The name of the category associated with the document.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"label_names": {
				Description: "The names of labels associated with the document.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	documentAlternative = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"phrase": {
				Description: "Alternate phrasing to the document title.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"autocomplete": {
				Description: "Autocomplete enabled for the alternate phrase.",
				Type:        schema.TypeBool,
				Optional:    true,
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
			knowledgeDocuments, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocuments(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), "", nil, nil, true, true, nil, nil)
			if getErr != nil {
				return nil, diag.Errorf("Failed to get page of Knowledge documents: %v", getErr)
			}

			if knowledgeDocuments.Entities == nil || len(*knowledgeDocuments.Entities) == 0 {
				break
			}
			for _, knowledgeDocument := range *knowledgeDocuments.Entities {
				id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id)
				resources[id] = &ResourceMeta{Name: *knowledgeDocument.Title}
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
			"knowledge_document": {
				Description: "Knowledge document request body",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeDocument,
			},
			"published": {
				Description: "If true, the knowledge document will be published. If false, it will be a draft",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func buildDocumentAlternatives(requestIn map[string]interface{}, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string) *[]platformclientv2.Knowledgedocumentalternative {
	if alternativesIn, ok := requestIn["alternatives"].([]interface{}); ok {
		alternativesOut := make([]platformclientv2.Knowledgedocumentalternative, 0)

		for _, alternative := range alternativesIn {
			alternativeMap := alternative.(map[string]interface{})
			phrase := alternativeMap["phrase"].(string)
			autocomplete := alternativeMap["autocomplete"].(bool)

			alternativeOut := platformclientv2.Knowledgedocumentalternative{
				Phrase:       &phrase,
				Autocomplete: &autocomplete,
			}

			alternativesOut = append(alternativesOut, alternativeOut)
		}

		return &alternativesOut
	}
	return nil
}

func buildKnowledgeDocumentRequest(d *schema.ResourceData, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string) platformclientv2.Knowledgedocumentreq {
	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	title := requestIn["title"].(string)
	visible := requestIn["visible"].(bool)

	requestOut := platformclientv2.Knowledgedocumentreq{
		Title:        &title,
		Visible:      &visible,
		Alternatives: buildDocumentAlternatives(requestIn, knowledgeAPI, knowledgeBaseId),
	}

	if categoryName, ok := requestIn["category_name"].(string); ok && categoryName != "" {
		pageSize := 1
		knowledgeCategories, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategories(knowledgeBaseId, "", "", fmt.Sprintf("%v", pageSize), "", false, categoryName, "", "", false)

		if getErr != nil {
			fmt.Errorf("Failed to get page of knowledge categories: %v", getErr)
		} else if len(*knowledgeCategories.Entities) > 0 {
			matchingCategory := (*knowledgeCategories.Entities)[0]
			requestOut.CategoryId = matchingCategory.Id
		}
	}
	if labelNames, ok := requestIn["label_names"].([]interface{}); ok && labelNames != nil {
		labelStringList := InterfaceListToStrings(labelNames)
		pageSize := 1
		labelIds := make([]string, 0)
		for _, labelName := range labelStringList {
			knowledgeLabels, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(knowledgeBaseId, "", "", fmt.Sprintf("%v", pageSize), labelName, false)

			if getErr != nil {
				fmt.Errorf("Failed to get page of knowledge labels: %v", getErr)
			} else if len(*knowledgeLabels.Entities) > 0 {
				matchingLabel := (*knowledgeLabels.Entities)[0]
				labelIds = append(labelIds, *matchingLabel.Id)
			}
		}
		requestOut.LabelIds = &labelIds
	}

	return requestOut
}

func flattenDocumentAlternatives(alternativesIn *[]platformclientv2.Knowledgedocumentalternative) []interface{} {
	if alternativesIn == nil || len(*alternativesIn) == 0 {
		return nil
	}

	alternativesOut := make([]interface{}, 0)

	for _, alternativeIn := range *alternativesIn {
		alternativeOut := make(map[string]interface{})

		if alternativeIn.Phrase != nil {
			alternativeOut["phrase"] = *alternativeIn.Phrase
		}
		if alternativeIn.Autocomplete != nil {
			alternativeOut["autocomplete"] = *alternativeIn.Autocomplete
		}
		alternativesOut = append(alternativesOut, alternativeOut)
	}

	return alternativesOut
}

func flattenKnowledgeDocument(documentIn *platformclientv2.Knowledgedocumentresponse, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string) []interface{} {
	if documentIn == nil {
		return nil
	}

	documentOut := make(map[string]interface{})

	documentOut["alternatives"] = flattenDocumentAlternatives(documentIn.Alternatives)

	if documentIn.Title != nil {
		documentOut["title"] = *documentIn.Title
	}
	if documentIn.Visible != nil {
		documentOut["visible"] = *documentIn.Visible
	}
	if documentIn.Category != nil {
		// use the id to retrieve the category name
		knowledgeCategory, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, *documentIn.Category.Id)

		if getErr != nil {
			fmt.Errorf("Failed to get knowledge category: %v", getErr)
		} else if knowledgeCategory.Name != nil {
			documentOut["category_name"] = knowledgeCategory.Name
		}
	}
	if documentIn.Labels != nil && len(*documentIn.Labels) > 0 {
		labelNames := make([]string, 0)
		for _, label := range *documentIn.Labels {
			knowledgeLabel, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, *label.Id)

			if getErr != nil {
				fmt.Errorf("Failed to get knowledge label: %v", getErr)
			} else if knowledgeLabel.Name != nil {
				labelNames = append(labelNames, *knowledgeLabel.Name)
			}
		}
		documentOut["label_names"] = labelNames
	}

	return []interface{}{documentOut}
}

func createKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	published := d.Get("published").(bool)
	body := buildKnowledgeDocumentRequest(d, knowledgeAPI, knowledgeBaseId)

	log.Printf("Creating knowledge document")
	knowledgeDocument, _, err := knowledgeAPI.PostKnowledgeKnowledgebaseDocuments(knowledgeBaseId, body)

	if err != nil {
		return diag.Errorf("Failed to create knowledge document: %s", err)
	}

	if published == true {
		_, _, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, *knowledgeDocument.Id, platformclientv2.Knowledgedocumentversion{})
		if versionErr != nil {
			return diag.Errorf("Failed to publish knowledge document: %s", err)
		}
	}

	id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, knowledgeBaseId)
	d.SetId(id)

	log.Printf("Created knowledge document %s", *knowledgeDocument.Id)
	return readKnowledgeDocument(ctx, d, meta)
}

func readKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	state := "Draft"
	if d.Get("published").(bool) == true {
		state = "Published"
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Reading knowledge document %s", knowledgeDocumentId)
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		knowledgeDocument, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, nil, state)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceKnowledgeDocument())

		// required
		id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, knowledgeBaseId)
		d.SetId(id)
		d.Set("knowledge_base_id", *knowledgeDocument.KnowledgeBase.Id)
		d.Set("knowledge_document", flattenKnowledgeDocument(knowledgeDocument, knowledgeAPI, knowledgeBaseId))

		if *knowledgeDocument.State == "Published" {
			d.Set("published", true)
		} else {
			d.Set("published", false)
		}

		log.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		fmt.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		checkState := cc.CheckState()
		return checkState
	})
}

func updateKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := d.Get("knowledge_base_id").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating Knowledge document %s", knowledgeDocumentId)
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Knowledge document version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, nil, "")
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Knowledge document %s: %s", knowledgeDocumentId, getErr)
		}

		update := buildKnowledgeDocumentRequest(d, knowledgeAPI, knowledgeBaseId)

		log.Printf("Updating knowledge document %s", knowledgeDocumentId)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, update)
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
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Knowledge document %s", knowledgeDocumentId)
	_, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId)
	if err != nil {
		return diag.Errorf("Failed to delete Knowledge document %s: %s", knowledgeDocumentId, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		state := "Draft"
		if d.Get("published").(bool) == true {
			state = "Published"
		}

		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeDocumentId, knowledgeBaseId, nil, state)
		if err != nil {
			if isStatus404(resp) {
				// Knowledge document deleted
				log.Printf("Deleted Knowledge document %s", knowledgeDocumentId)
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Knowledge document %s: %s", knowledgeDocumentId, err))
		}

		return resource.RetryableError(fmt.Errorf("Knowledge document %s still exists", knowledgeDocumentId))
	})
}
