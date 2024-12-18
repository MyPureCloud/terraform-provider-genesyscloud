package knowledge_document

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

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

func buildKnowledgeDocumentCreateRequest(ctx context.Context, d *schema.ResourceData, proxy *knowledgeDocumentProxy, knowledgeBaseId string) (*platformclientv2.Knowledgedocumentcreaterequest, diag.Diagnostics) {
	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	title := requestIn["title"].(string)
	visible := requestIn["visible"].(bool)

	requestOut := platformclientv2.Knowledgedocumentcreaterequest{
		Title:        &title,
		Visible:      &visible,
		Alternatives: buildDocumentAlternatives(requestIn),
	}

	if categoryName, ok := requestIn["category_name"].(string); ok && categoryName != "" {
		knowledgeCategories, resp, getErr := proxy.getKnowledgeKnowledgebaseCategories(ctx, knowledgeBaseId, categoryName)
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
		labelIds := make([]string, 0)
		for _, labelName := range labelStringList {
			knowledgeLabels, resp, getErr := proxy.getKnowledgeKnowledgebaseLabels(ctx, knowledgeBaseId, labelName)
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

func buildKnowledgeDocumentRequest(ctx context.Context, d *schema.ResourceData, proxy *knowledgeDocumentProxy, knowledgeBaseId string) (*platformclientv2.Knowledgedocumentreq, diag.Diagnostics) {
	requestIn := d.Get("knowledge_document").([]interface{})[0].(map[string]interface{})
	title := requestIn["title"].(string)
	visible := requestIn["visible"].(bool)

	requestOut := platformclientv2.Knowledgedocumentreq{
		Title:        &title,
		Visible:      &visible,
		Alternatives: buildDocumentAlternatives(requestIn),
	}

	if categoryName, ok := requestIn["category_name"].(string); ok && categoryName != "" {
		knowledgeCategories, resp, getErr := proxy.getKnowledgeKnowledgebaseCategories(ctx, knowledgeBaseId, categoryName)
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
		labelIds := make([]string, 0)
		for _, labelName := range labelStringList {
			knowledgeLabels, resp, getErr := proxy.getKnowledgeKnowledgebaseLabels(ctx, knowledgeBaseId, labelName)
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
