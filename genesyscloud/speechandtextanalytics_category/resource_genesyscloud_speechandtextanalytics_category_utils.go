package speechandtextanalytics_category

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

// buildCategoryFromResourceData constructs an SDK Categoryrequest from Terraform ResourceData.
// Read-only fields (id, createdBy, dateCreated, modifiedBy, dateModified, selfUri) are intentionally omitted
// from the request body.
func buildCategoryFromResourceData(d *schema.ResourceData) (*platformclientv2.Categoryrequest, error) {
	name := d.Get("name").(string)

	category := &platformclientv2.Categoryrequest{
		Name: &name,
	}

	// interaction_type is required by the API
	interactionType := d.Get("interaction_type").(string)
	category.InteractionType = &interactionType

	// Use d.Get for description so clearing the field takes effect
	description := d.Get("description").(string)
	category.Description = &description

	// criteria is required by the API; parse the JSON string into the SDK Operand model
	criteriaStr := d.Get("criteria").(string)
	var criteria platformclientv2.Operand
	if err := json.Unmarshal([]byte(criteriaStr), &criteria); err != nil {
		return nil, fmt.Errorf("error parsing criteria: %s", err)
	}
	category.Criteria = &criteria

	return category, nil
}

// setCategoryToResourceData flattens an SDK Stacategory into Terraform ResourceData.
func setCategoryToResourceData(d *schema.ResourceData, category *platformclientv2.Stacategory) {
	if category.Name != nil {
		_ = d.Set("name", *category.Name)
	}
	if category.Description != nil {
		_ = d.Set("description", *category.Description)
	}
	if category.InteractionType != nil {
		_ = d.Set("interaction_type", *category.InteractionType)
	}

	if criteriaJSON := flattenCriteriaToJSON(category.Criteria); criteriaJSON != "" {
		_ = d.Set("criteria", criteriaJSON)
	}
}

// flattenCriteriaToJSON converts the criteria Operand from the SDK to a JSON string
func flattenCriteriaToJSON(criteria *platformclientv2.Operand) string {
	if criteria == nil {
		return ""
	}
	bytes, err := json.Marshal(criteria)
	if err != nil {
		return ""
	}
	return string(bytes)
}
