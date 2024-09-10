package knowledge_document

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllKnowledgeDocuments(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	documentEntities := make([]platformclientv2.Knowledgedocumentresponse, 0)
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
		partialEntities, err := GetAllKnowledgeDocumentEntities(*knowledgeAPI, &knowledgeBase, clientConfig)
		if err != nil {
			return nil, err
		}
		documentEntities = append(documentEntities, *partialEntities...)
	}

	for _, knowledgeDocument := range documentEntities {
		id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id)
		resources[id] = &resourceExporter.ResourceMeta{Name: *knowledgeDocument.Title}
	}

	return resources, nil
}

func GetAllKnowledgeDocumentEntities(knowledgeAPI platformclientv2.KnowledgeApi, knowledgeBase *platformclientv2.Knowledgebase, clientConfig *platformclientv2.Configuration) (*[]platformclientv2.Knowledgedocumentresponse, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Knowledgedocumentresponse
	)

	resources := make(resourceExporter.ResourceIDMetaMap)

	const pageSize = 100
	// prepare base url
	resourcePath := fmt.Sprintf("/api/v2/knowledge/knowledgebases/%s/documents", url.PathEscape(*knowledgeBase.Id))
	listDocumentsBaseUrl := fmt.Sprintf("%s%s", knowledgeAPI.Configuration.BasePath, resourcePath)

	for {
		// prepare query params
		queryParams := make(map[string]string, 0)
		queryParams["after"] = after
		queryParams["pageSize"] = fmt.Sprintf("%v", pageSize)
		queryParams["includeDrafts"] = "true"

		// prepare headers
		headers := make(map[string]string)
		headers["Authorization"] = fmt.Sprintf("Bearer %s", clientConfig.AccessToken)
		headers["Content-Type"] = "application/json"
		headers["Accept"] = "application/json"

		// execute request
		response, err := clientConfig.APIClient.CallAPI(listDocumentsBaseUrl, "GET", nil, headers, queryParams, nil, "", nil)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document list response error: %s", err), response)
		}

		// process response
		var knowledgeDocuments platformclientv2.Knowledgedocumentresponselisting
		unmarshalErr := json.Unmarshal(response.RawBody, &knowledgeDocuments)
		if unmarshalErr != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to unmarshal knowledge document list response"), unmarshalErr)
		}

		/**
		 * Todo: restore direct SDK invocation and remove workaround once the SDK supports optional boolean args.
		 */
		// knowledgeDocuments, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocuments(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", nil, nil, true, true, nil, nil)
		// if getErr != nil {
		// 	return nil, diag.Errorf("Failed to get page of knowledge documents: %v", getErr)
		// }

		if knowledgeDocuments.Entities == nil || len(*knowledgeDocuments.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeDocuments.Entities...)

		if knowledgeDocuments.NextUri == nil || *knowledgeDocuments.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*knowledgeDocuments.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to parse after cursor from knowledge document nextUri"), err)
		}
		if after == "" {
			break
		}
		for _, knowledgeDocument := range *knowledgeDocuments.Entities {
			id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id)
			resources[id] = &resourceExporter.ResourceMeta{Name: *knowledgeDocument.Title}
		}
	}

	return &entities, nil
}

func buildDocumentAlternatives(requestIn map[string]interface{}) *[]platformclientv2.Knowledgedocumentalternative {
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

func buildKnowledgeDocumentRequest(d *schema.ResourceData, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string) (*platformclientv2.Knowledgedocumentreq, diag.Diagnostics) {
	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	title := requestIn["title"].(string)
	visible := requestIn["visible"].(bool)

	requestOut := platformclientv2.Knowledgedocumentreq{
		Title:        &title,
		Visible:      &visible,
		Alternatives: buildDocumentAlternatives(requestIn),
	}

	if categoryName, ok := requestIn["category_name"].(string); ok && categoryName != "" {
		pageSize := 1
		knowledgeCategories, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategories(knowledgeBaseId, "", "", fmt.Sprintf("%v", pageSize), "", false, categoryName, "", "", false)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to get page of knowledge categories error: %s", getErr), resp)
		}
		if len(*knowledgeCategories.Entities) > 0 {
			matchingCategory := (*knowledgeCategories.Entities)[0]
			requestOut.CategoryId = matchingCategory.Id
		}
	}
	if labelNames, ok := requestIn["label_names"].([]interface{}); ok && labelNames != nil {
		labelStringList := lists.InterfaceListToStrings(labelNames)
		pageSize := 1
		labelIds := make([]string, 0)
		for _, labelName := range labelStringList {
			knowledgeLabels, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(knowledgeBaseId, "", "", fmt.Sprintf("%v", pageSize), labelName, false)
			if getErr != nil {
				return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to get page of knowledge labels error: %s", getErr), resp)
			}
			if len(*knowledgeLabels.Entities) > 0 {
				matchingLabel := (*knowledgeLabels.Entities)[0]
				labelIds = append(labelIds, *matchingLabel.Id)
			}
		}
		requestOut.LabelIds = &labelIds
	}

	return &requestOut, nil
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

func flattenKnowledgeDocument(documentIn *platformclientv2.Knowledgedocumentresponse, knowledgeAPI *platformclientv2.KnowledgeApi, knowledgeBaseId string) ([]interface{}, error) {
	if documentIn == nil {
		return nil, nil
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
			return nil, fmt.Errorf("Failed to get knowledge category: %v", getErr)
		}
		if knowledgeCategory.Name != nil {
			documentOut["category_name"] = knowledgeCategory.Name
		}
	}
	if documentIn.Labels != nil && len(*documentIn.Labels) > 0 {
		labelNames := make([]string, 0)
		for _, label := range *documentIn.Labels {
			knowledgeLabel, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, *label.Id)

			if getErr != nil {
				return nil, fmt.Errorf("Failed to get knowledge label: %v", getErr)
			}
			if knowledgeLabel.Name != nil {
				labelNames = append(labelNames, *knowledgeLabel.Name)
			}
		}
		documentOut["label_names"] = labelNames
	}

	return []interface{}{documentOut}, nil
}

func createKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	published := d.Get("published").(bool)

	body, buildErr := buildKnowledgeDocumentRequest(d, knowledgeAPI, knowledgeBaseId)
	if buildErr != nil {
		return buildErr
	}

	log.Printf("Creating knowledge document")
	knowledgeDocument, resp, err := knowledgeAPI.PostKnowledgeKnowledgebaseDocuments(knowledgeBaseId, *body)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to create knowledge document %s error: %s", d.Id(), err), resp)
	}

	if published {
		_, resp, versionErr := knowledgeAPI.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, *knowledgeDocument.Id, platformclientv2.Knowledgedocumentversion{})
		if versionErr != nil {
			_, deleteError := knowledgeAPI.DeleteKnowledgeKnowledgebaseDocument(knowledgeBaseId, *knowledgeDocument.Id)
			if deleteError != nil {
				log.Printf("failed to delete draft knowledge document %s error: %s", *knowledgeDocument.Id, deleteError)
			}
			return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to publish knowledge document error: %s", versionErr), resp)
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocument(), constants.DefaultConsistencyChecks, "genesyscloud_knowledge_document")

	log.Printf("Reading knowledge document %s", knowledgeDocumentId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeDocument, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, nil, state)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr), resp))
		}

		// required
		id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, knowledgeBaseId)
		d.SetId(id)
		d.Set("knowledge_base_id", *knowledgeDocument.KnowledgeBase.Id)

		flattenedDocument, err := flattenKnowledgeDocument(knowledgeDocument, knowledgeAPI, knowledgeBaseId)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		d.Set("knowledge_document", flattenedDocument)

		if *knowledgeDocument.State == "Published" {
			d.Set("published", true)
		} else {
			d.Set("published", false)
		}

		log.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		checkState := cc.CheckState(d)
		return checkState
	})
}

func updateKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	state := "Draft"
	if d.Get("published").(bool) == true {
		state = "Published"
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating Knowledge document %s", knowledgeDocumentId)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Knowledge document version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, nil, state)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document %s error: %s", knowledgeDocumentId, getErr), resp)
		}

		update, err := buildKnowledgeDocumentRequest(d, knowledgeAPI, knowledgeBaseId)
		if err != nil {
			return nil, err
		}

		log.Printf("Updating knowledge document %s", knowledgeDocumentId)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, *update)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to update knowledge document %s error: %s", knowledgeDocumentId, putErr), resp)
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Knowledge document %s", knowledgeDocumentId)
	resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to delete knowledge document %s error: %s", knowledgeDocumentId, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		state := "Draft"
		if d.Get("published").(bool) == true {
			state = "Published"
		}

		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeDocumentId, knowledgeBaseId, nil, state)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge document deleted
				log.Printf("Deleted Knowledge document %s", knowledgeDocumentId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Error deleting Knowledge document %s | error: %s", knowledgeDocumentId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Knowledge document %s still exists", knowledgeDocumentId), resp))
	})
}

func getAllKnowledgebaseEntities(knowledgeApi platformclientv2.KnowledgeApi, published bool) (*[]platformclientv2.Knowledgebase, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for {
		knowledgeBases, resp, getErr := knowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to get page of knowledge bases error: %s", getErr), resp)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to parse after cursor from knowledge base nextUri"), err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}
