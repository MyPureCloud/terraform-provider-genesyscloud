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
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
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

func buildKnowledgeDocumentCreateRequest(ctx context.Context, d *schema.ResourceData, proxy *knowledgeDocumentProxy, knowledgeBaseId string) (*platformclientv2.Knowledgedocumentreq, diag.Diagnostics) {
	logSuffix := "Knowledge Base: " + strconv.Quote(knowledgeBaseId)
	log.Println("Building Knowledgedocumentcreaterequest object. ", logSuffix)

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

	log.Println("Successfully built Knowledgedocumentcreaterequest object. ", logSuffix)
	return &requestOut, nil
}

func buildKnowledgeDocumentCategoryId(ctx context.Context, knowledgeBaseId, categoryName string, proxy *knowledgeDocumentProxy) (string, diag.Diagnostics) {
	knowledgeCategories, resp, getErr := proxy.getKnowledgeKnowledgebaseCategories(ctx, knowledgeBaseId, categoryName)
	if getErr != nil {
		return "", util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge categories error: %s", getErr), resp)
	}

	if knowledgeCategories == nil || knowledgeCategories.Entities == nil || len(*knowledgeCategories.Entities) == 0 {
		return "", nil
	}
	// DEVTOOLING-1821: the category name query is a partial match, so it can return several
	// categories whose names contain the requested string. Select the entity whose name matches
	// exactly instead of taking the first hit. This matches the exact-name convention already
	// used by the genesyscloud_knowledge_label data source.
	matchingCategory, ambiguous := findExactCategoryMatch(knowledgeCategories.Entities, categoryName)
	if matchingCategory != nil && matchingCategory.Id != nil {
		return *matchingCategory.Id, nil
	}
	if ambiguous {
		// Multiple categories match case-insensitively and none matches exactly; picking one would
		// risk attaching the wrong category. Fail loudly so the user disambiguates by exact casing.
		return "", util.BuildDiagnosticError(ResourceType, fmt.Sprintf("multiple knowledge categories match name %q case-insensitively in knowledge base %s; specify the exact category name", categoryName, knowledgeBaseId), nil)
	}

	return "", nil
}

// findExactCategoryMatch selects the category matching categoryName.
// It prefers a case-sensitive exact match. If none exists, it falls back to a unique
// case-insensitive match (the Knowledge API name filter is case-insensitive, so a config whose
// casing differs from the stored name should still resolve). It returns (nil, true) only when the
// case-insensitive fallback is ambiguous (more than one differently-cased candidate), so the caller
// can fail loudly instead of guessing. Returns (nil, false) when nothing matches at all.
func findExactCategoryMatch(entities *[]platformclientv2.Categoryresponse, categoryName string) (match *platformclientv2.Categoryresponse, ambiguous bool) {
	if entities == nil {
		return nil, false
	}
	var caseInsensitive []*platformclientv2.Categoryresponse
	for i := range *entities {
		category := &(*entities)[i]
		if category.Name == nil {
			continue
		}
		if *category.Name == categoryName {
			return category, false
		}
		if strings.EqualFold(*category.Name, categoryName) {
			caseInsensitive = append(caseInsensitive, category)
		}
	}
	if len(caseInsensitive) == 1 {
		return caseInsensitive[0], false
	}
	if len(caseInsensitive) > 1 {
		return nil, true
	}
	return nil, false
}

func buildKnowledgeDocumentLabelIds(ctx context.Context, proxy *knowledgeDocumentProxy, knowledgeBaseId string, labelNames []any) ([]string, diag.Diagnostics) {
	labelStringList := lists.InterfaceListToStrings(labelNames)
	labelIds := make([]string, 0)
	for _, labelName := range labelStringList {
		knowledgeLabels, resp, getErr := proxy.getKnowledgeKnowledgebaseLabels(ctx, knowledgeBaseId, labelName)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge labels error: %s", getErr), resp)
		}
		if knowledgeLabels == nil || knowledgeLabels.Entities == nil || len(*knowledgeLabels.Entities) == 0 {
			continue
		}
		// DEVTOOLING-1821: the label name query is a partial match, so it can return several labels
		// whose names contain the requested string (e.g. searching "Home" also returns "Auto and
		// Home"). Select the entity whose name matches exactly instead of taking the first hit,
		// which previously attached the wrong label and caused a label_names consistency mismatch.
		// This matches the exact-name convention already used by the genesyscloud_knowledge_label
		// data source.
		matchingLabel, ambiguous := findExactLabelMatch(knowledgeLabels.Entities, labelName)
		if ambiguous {
			// Multiple labels match case-insensitively and none matches exactly; picking one would
			// risk attaching the wrong label (the original DEVTOOLING-1821 bug). Fail loudly so the
			// user disambiguates by exact casing.
			return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("multiple knowledge labels match name %q case-insensitively in knowledge base %s; specify the exact label name", labelName, knowledgeBaseId), nil)
		}
		if matchingLabel != nil && matchingLabel.Id != nil {
			labelIds = append(labelIds, *matchingLabel.Id)
		}
	}
	return labelIds, nil
}

// findExactLabelMatch selects the label matching labelName.
// It prefers a case-sensitive exact match. If none exists, it falls back to a unique
// case-insensitive match (the Knowledge API name filter is case-insensitive, so a config whose
// casing differs from the stored name should still resolve). It returns (nil, true) only when the
// case-insensitive fallback is ambiguous (more than one differently-cased candidate), so the caller
// can fail loudly instead of guessing. Returns (nil, false) when nothing matches at all.
func findExactLabelMatch(entities *[]platformclientv2.Labelresponse, labelName string) (match *platformclientv2.Labelresponse, ambiguous bool) {
	if entities == nil {
		return nil, false
	}
	var caseInsensitive []*platformclientv2.Labelresponse
	for i := range *entities {
		label := &(*entities)[i]
		if label.Name == nil {
			continue
		}
		if *label.Name == labelName {
			return label, false
		}
		if strings.EqualFold(*label.Name, labelName) {
			caseInsensitive = append(caseInsensitive, label)
		}
	}
	if len(caseInsensitive) == 1 {
		return caseInsensitive[0], false
	}
	if len(caseInsensitive) > 1 {
		return nil, true
	}
	return nil, false
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
	} else {
		documentOut["label_names"] = []string{}
	}

	return []interface{}{documentOut}, nil
}
