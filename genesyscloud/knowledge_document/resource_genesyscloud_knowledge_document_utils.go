package knowledge_document

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

const documentIDSeparator = ","

func BuildDocumentResourceDataID(knowledgeDocumentId, knowledgeBaseId string) string {
	return knowledgeDocumentId + documentIDSeparator + knowledgeBaseId
}

func parseDocumentResourceDataID(id string) (knowledgeDocumentID, knowledgeBaseID string) {
	split := strings.Split(id, documentIDSeparator)
	return split[0], split[1]
}

func buildDocumentAlternatives(requestIn map[string]any) *[]platformclientv2.Knowledgedocumentalternative {
	alternativesIn, ok := requestIn["alternatives"].([]any)
	if !ok || len(alternativesIn) == 0 {
		return nil
	}

	alternativesOut := make([]platformclientv2.Knowledgedocumentalternative, 0)

	for _, alternative := range alternativesIn {
		alternativeMap, ok := alternative.(map[string]any)
		if !ok {
			log.Printf("invalid type for alternatives item. Expected map[string]any, got %T", alternative)
			continue
		}
		alternativeOut := platformclientv2.Knowledgedocumentalternative{
			Phrase:       resourcedata.GetNillableValueFromMap[string](alternativeMap, "phrase", true),
			Autocomplete: resourcedata.GetNillableValueFromMap[bool](alternativeMap, "autocomplete", true),
		}
		alternativesOut = append(alternativesOut, alternativeOut)
	}

	return &alternativesOut
}

func buildKnowledgeDocumentCreateRequest(ctx context.Context, d *schema.ResourceData, proxy *knowledgeDocumentProxy, knowledgeBaseId string) (*platformclientv2.Knowledgedocumentcreaterequest, diag.Diagnostics) {
	logSuffix := "Knowledge Base: " + strconv.Quote(knowledgeBaseId)
	log.Println("Building Knowledgedocumentcreaterequest object. ", logSuffix)

	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	title := requestIn["title"].(string)
	visible := requestIn["visible"].(bool)

	requestOut := platformclientv2.Knowledgedocumentcreaterequest{
		Title:        &title,
		Visible:      &visible,
		Alternatives: buildDocumentAlternatives(requestIn),
	}

	categoryName, ok := requestIn["category_name"].(string)
	if ok && categoryName != "" {
		log.Printf("Retrieving category ID for category name %s. %s", categoryName, logSuffix)
		categoryId, diagErr := buildKnowledgeDocumentCategoryId(ctx, knowledgeBaseId, categoryName, proxy)
		if diagErr != nil {
			log.Printf("Encountered error while retrieving category ID for category name %s: %v. %s", strconv.Quote(categoryName), diagErr, logSuffix)
			return nil, diagErr
		}

		if categoryId != "" {
			requestOut.CategoryId = &categoryId
		}
	}

	labelNames, ok := requestIn["label_names"].([]any)
	if !ok || labelNames == nil {
		return &requestOut, nil
	}

	log.Printf("Retrieving label IDs. %s", logSuffix)
	labelIds, diagErr := buildKnowledgeDocumentLabelIds(ctx, proxy, knowledgeBaseId, labelNames)
	if diagErr != nil {
		log.Printf("Encountered error while retrieving label IDs: %v. %s", diagErr, logSuffix)
		return nil, diagErr
	}

	if len(labelIds) != 0 {
		requestOut.LabelIds = &labelIds
	}

	log.Println("Successfully built Knowledgedocumentcreaterequest object. ", logSuffix)
	return &requestOut, nil
}

func buildKnowledgeDocumentCategoryId(ctx context.Context, knowledgeBaseId, categoryName string, proxy *knowledgeDocumentProxy) (string, diag.Diagnostics) {
	knowledgeCategories, resp, getErr := proxy.getKnowledgeKnowledgebaseCategories(ctx, knowledgeBaseId, categoryName)
	if getErr != nil {
		return "", util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge categories error: %s", getErr), resp)
	}

	if len(*knowledgeCategories.Entities) > 0 {
		matchingCategory := (*knowledgeCategories.Entities)[0]
		return *matchingCategory.Id, nil
	}

	return "", nil
}

func buildKnowledgeDocumentLabelIds(ctx context.Context, proxy *knowledgeDocumentProxy, knowledgeBaseId string, labelNames []any) ([]string, diag.Diagnostics) {
	labelStringList := lists.InterfaceListToStrings(labelNames)
	labelIds := make([]string, 0)
	for _, labelName := range labelStringList {
		knowledgeLabels, resp, getErr := proxy.getKnowledgeKnowledgebaseLabels(ctx, knowledgeBaseId, labelName)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge labels error: %s", getErr), resp)
		}
		if len(*knowledgeLabels.Entities) > 0 {
			matchingLabel := (*knowledgeLabels.Entities)[0]
			labelIds = append(labelIds, *matchingLabel.Id)
		}
	}
	return labelIds, nil
}

func buildKnowledgeDocumentRequest(ctx context.Context, d *schema.ResourceData, proxy *knowledgeDocumentProxy, knowledgeBaseId string) (*platformclientv2.Knowledgedocumentreq, diag.Diagnostics) {
	logSuffix := "Knowledge Base: " + strconv.Quote(knowledgeBaseId)
	log.Println("Building Knowledgedocumentreq object. ", logSuffix)

	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	title := requestIn["title"].(string)
	visible := requestIn["visible"].(bool)

	requestOut := platformclientv2.Knowledgedocumentreq{
		Title:        &title,
		Visible:      &visible,
		Alternatives: buildDocumentAlternatives(requestIn),
	}

	categoryName, ok := requestIn["category_name"].(string)
	if ok && categoryName != "" {
		log.Printf("Retrieving category ID for category name %s. %s", categoryName, logSuffix)
		categoryId, diagErr := buildKnowledgeDocumentCategoryId(ctx, knowledgeBaseId, categoryName, proxy)
		if diagErr != nil {
			log.Printf("Encountered error while retrieving category ID for category name %s: %v. %s", strconv.Quote(categoryName), diagErr, logSuffix)
			return nil, diagErr
		}

		if categoryId != "" {
			requestOut.CategoryId = &categoryId
		}
	}

	labelNames, ok := requestIn["label_names"].([]any)
	if !ok || labelNames == nil {
		return &requestOut, nil
	}

	log.Printf("Retrieving label IDs. %s", logSuffix)
	labelIds, diagErr := buildKnowledgeDocumentLabelIds(ctx, proxy, knowledgeBaseId, labelNames)
	if diagErr != nil {
		log.Printf("Encountered error while retrieving label IDs: %v. %s", diagErr, logSuffix)
		return nil, diagErr
	}

	if len(labelIds) != 0 {
		requestOut.LabelIds = &labelIds
	}

	log.Println("Successfully built Knowledgedocumentreq object. ", logSuffix)

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

func flattenKnowledgeDocument(ctx context.Context, documentIn *platformclientv2.Knowledgedocumentresponse, proxy *knowledgeDocumentProxy, knowledgeBaseId string) ([]interface{}, error) {
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
		knowledgeCategory, _, getErr := proxy.getKnowledgeKnowledgebaseCategory(ctx, knowledgeBaseId, *documentIn.Category.Id)

		if getErr != nil {
			return nil, fmt.Errorf("failed to get knowledge category: %v", getErr)
		}
		if knowledgeCategory.Name != nil {
			documentOut["category_name"] = knowledgeCategory.Name
		}
	}
	if documentIn.Labels != nil && len(*documentIn.Labels) > 0 {
		labelNames := make([]string, 0)
		for _, label := range *documentIn.Labels {
			knowledgeLabel, _, getErr := proxy.getKnowledgeKnowledgebaseLabel(ctx, knowledgeBaseId, *label.Id)

			if getErr != nil {
				return nil, fmt.Errorf("failed to get knowledge label: %v", getErr)
			}
			if knowledgeLabel.Name != nil {
				labelNames = append(labelNames, *knowledgeLabel.Name)
			}
		}
		documentOut["label_names"] = labelNames
	}

	return []interface{}{documentOut}, nil
}
