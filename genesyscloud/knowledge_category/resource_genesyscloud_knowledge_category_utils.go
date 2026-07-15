package knowledge_category

import (
	"log"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func buildKnowledgeCategoryUpdate(categoryIn map[string]interface{}) *platformclientv2.Categoryupdaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categoryupdaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}

	parentId, ok := categoryIn["parent_id"].(string)
	if !ok || parentId == "" {
		return &categoryOut
	}

	parentCategoryId := parseParentCategoryID(parentId)
	categoryOut.ParentCategoryId = &parentCategoryId

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
		parentCategoryId := parseParentCategoryID(parentId)
		categoryOut.ParentCategoryId = &parentCategoryId
	}

	return &categoryOut
}

func parseParentCategoryID(parentId string) string {
	if !strings.Contains(parentId, resourcedata.CompositeIDSeparator) {
		return parentId
	}

	parentCategoryId, _, err := ParseCompositeKnowledgeCategoryID(parentId)
	if err != nil {
		log.Printf("failed to parse parent ID: %s. Using as is", parentId)
		return parentId
	}
	return parentCategoryId
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
		categoryOut["parent_id"] = BuildCompositeKnowledgeCategoryID(
			*(*categoryIn.ParentCategory).Id,
			*(*categoryIn.KnowledgeBase).Id,
		)
	}

	return []interface{}{categoryOut}
}

func BuildCompositeKnowledgeCategoryID(categoryId, knowledgeBaseId string) string {
	return resourcedata.BuildCompositeID(categoryId, knowledgeBaseId)
}

func ParseCompositeKnowledgeCategoryID(id string) (categoryId, knowledgeBaseId string, _ error) {
	categoryId, relatedIds, err := resourcedata.ParseCompositeID(id)
	if err != nil {
		return "", "", err
	}
	return categoryId, relatedIds[0], nil
}
