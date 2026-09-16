package knowledge_category

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

const knowledgeCategoryIdSeparator = ","

// BuildKnowledgeCategoryId builds the knowledge_category composite resource ID, which is
// always in the format <knowledge-category-id>,<knowledge-base-id>.
func BuildKnowledgeCategoryId(knowledgeCategoryId, knowledgeBaseId string) (id string) {
	return fmt.Sprintf("%s%s%s", knowledgeCategoryId, knowledgeCategoryIdSeparator, knowledgeBaseId)
}

// SplitKnowledgeCategoryId splits the knowledge_category composite resource ID, which is
// always in the format <knowledge-category-id>,<knowledge-base-id>, back into its parts.
func SplitKnowledgeCategoryId(id string) (knowledgeCategoryId string, knowledgeBaseId string) {
	parts := strings.Split(id, knowledgeCategoryIdSeparator)
	return parts[0], parts[1]
}

func buildKnowledgeCategoryUpdate(categoryIn map[string]interface{}) *platformclientv2.Categoryupdaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categoryupdaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}

	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		if strings.Contains(parentId, knowledgeCategoryIdSeparator) {
			parent_Id, _ := SplitKnowledgeCategoryId(parentId)
			categoryOut.ParentCategoryId = &parent_Id
		} else {
			categoryOut.ParentCategoryId = &parentId
		}
	}
	return &categoryOut
}

func buildKnowledgeCategoryCreate(categoryIn map[string]interface{}) *platformclientv2.Categorycreaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categorycreaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}
	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		if strings.Contains(parentId, knowledgeCategoryIdSeparator) {
			parent_Id, _ := SplitKnowledgeCategoryId(parentId)
			categoryOut.ParentCategoryId = &parent_Id
		} else {
			categoryOut.ParentCategoryId = &parentId
		}
	}

	return &categoryOut
}

func flattenKnowledgeCategory(categoryIn platformclientv2.Categoryresponse) []interface{} {
	categoryOut := make(map[string]interface{})

	if categoryIn.Name != nil {
		categoryOut["name"] = *categoryIn.Name
	}
	if categoryIn.Description != nil {
		categoryOut["description"] = *categoryIn.Description
	}
	if categoryIn.ParentCategory != nil && (*categoryIn.ParentCategory).Id != nil {
		categoryOut["parent_id"] = BuildKnowledgeCategoryId(*(*categoryIn.ParentCategory).Id, *(*categoryIn.KnowledgeBase).Id)
	}

	return []interface{}{categoryOut}
}
