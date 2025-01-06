package knowledge_category

import (
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func buildKnowledgeCategoryUpdate(categoryIn map[string]interface{}) *platformclientv2.Categoryupdaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categoryupdaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}

	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		if strings.Contains(parentId, ",") {
			ids := strings.Split(parentId, ",")
			parent_Id := ids[0]
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
		if strings.Contains(parentId, ",") {
			ids := strings.Split(parentId, ",")
			parent_Id := ids[0]
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
		categoryOut["parent_id"] = *(*categoryIn.ParentCategory).Id + "," + *(*categoryIn.KnowledgeBase).Id
	}

	return []interface{}{categoryOut}
}
