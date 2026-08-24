package access_policy

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

// buildAccessPolicyFromResourceData constructs an SDK Authorizationpolicy from Terraform ResourceData
func buildAccessPolicyFromResourceData(d *schema.ResourceData) (*platformclientv2.Authorizationpolicy, error) {
	name := d.Get("name").(string)
	effect := d.Get("effect").(string)
	enabled := d.Get("enabled").(bool)
	applyToClients := d.Get("apply_to_clients").(bool)
	subjectType := d.Get("subject_type").(string)

	policy := &platformclientv2.Authorizationpolicy{
		Name:           &name,
		Effect:         &effect,
		Active:         &enabled,
		ApplyToClients: &applyToClients,
	}

	// Set target_resource
	if v, ok := d.GetOk("target_resource"); ok {
		targetResource := v.(string)
		policy.TargetResource = &targetResource
	}

	// Set description
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		policy.Description = &description
	}

	// Build subject
	subject := &platformclientv2.Subject{
		VarType: &subjectType,
	}
	if v, ok := d.GetOk("subject_id"); ok {
		subjectId := v.(string)
		subject.Id = &subjectId
	}
	policy.Subject = subject

	// Parse condition JSON
	if v, ok := d.GetOk("condition_json"); ok {
		conditionStr := v.(string)
		var condition interface{}
		if err := json.Unmarshal([]byte(conditionStr), &condition); err != nil {
			return nil, fmt.Errorf("error parsing condition_json: %s", err)
		}
		policy.Condition = &condition
	}

	// Parse preset attributes JSON
	if v, ok := d.GetOk("preset_attributes_json"); ok {
		presetStr := v.(string)
		var presetAttrs map[string]platformclientv2.Typedattribute
		if err := json.Unmarshal([]byte(presetStr), &presetAttrs); err != nil {
			return nil, fmt.Errorf("error parsing preset_attributes_json: %s", err)
		}
		policy.PresetAttributes = &presetAttrs
	}

	return policy, nil
}

// flattenConditionToJSON converts the condition interface from the SDK to a JSON string
func flattenConditionToJSON(condition *interface{}) string {
	if condition == nil || *condition == nil {
		return ""
	}
	bytes, err := json.Marshal(*condition)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// flattenPresetAttributesToJSON converts the preset attributes map from the SDK to a JSON string
func flattenPresetAttributesToJSON(presetAttributes *map[string]platformclientv2.Typedattribute) string {
	if presetAttributes == nil || len(*presetAttributes) == 0 {
		return ""
	}
	bytes, err := json.Marshal(*presetAttributes)
	if err != nil {
		return ""
	}
	return string(bytes)
}
