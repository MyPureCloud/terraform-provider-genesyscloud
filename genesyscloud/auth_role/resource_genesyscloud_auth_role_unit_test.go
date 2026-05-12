package auth_role

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

// TestFlattenWildcardActionSetSuppression verifies that when the user's config has
// action_set = ["*"] and the API returns an expanded list of actions, the flatten
// function preserves the wildcard in state to suppress the diff.
func TestFlattenWildcardActionSetSuppression(t *testing.T) {
	// Simulate API response: expanded actions instead of wildcard
	domain := "directory"
	entityName := "user"
	actions := []string{"add", "edit", "view", "delete"}
	apiPolicies := []platformclientv2.Domainpermissionpolicy{
		{
			Domain:     &domain,
			EntityName: &entityName,
			ActionSet:  &actions,
		},
	}

	// Simulate user's config: action_set = ["*"]
	configuredPolicies := []interface{}{
		map[string]interface{}{
			"domain":      "directory",
			"entity_name": "user",
			"action_set":  schema.NewSet(schema.HashString, []interface{}{"*"}),
		},
	}

	// Run the flatten function with wildcard suppression
	result := flattenRolePermissionPoliciesWithWildcardSuppress(apiPolicies, configuredPolicies)

	// Verify the result preserves the wildcard
	policies := result.List()
	if len(policies) != 1 {
		t.Fatalf("Expected 1 policy, got %d", len(policies))
	}

	policyMap := policies[0].(map[string]interface{})

	// Check domain and entity_name are preserved
	if policyMap["domain"] != "directory" {
		t.Errorf("Expected domain 'directory', got '%s'", policyMap["domain"])
	}
	if policyMap["entity_name"] != "user" {
		t.Errorf("Expected entity_name 'user', got '%s'", policyMap["entity_name"])
	}

	// Check action_set is collapsed back to ["*"]
	actionSet := policyMap["action_set"].(*schema.Set)
	actionList := actionSet.List()
	if len(actionList) != 1 {
		t.Fatalf("Expected action_set to have 1 element (wildcard), got %d: %v", len(actionList), actionList)
	}
	if actionList[0].(string) != "*" {
		t.Errorf("Expected action_set to be ['*'], got ['%s']", actionList[0].(string))
	}
}

// TestFlattenExplicitActionsNoSuppression verifies that when the user's config has
// explicit actions (not wildcards), the flatten function returns the API response as-is.
func TestFlattenExplicitActionsNoSuppression(t *testing.T) {
	// Simulate API response
	domain := "routing"
	entityName := "queue"
	actions := []string{"add", "edit", "view"}
	apiPolicies := []platformclientv2.Domainpermissionpolicy{
		{
			Domain:     &domain,
			EntityName: &entityName,
			ActionSet:  &actions,
		},
	}

	// Simulate user's config: explicit actions, no wildcard
	configuredPolicies := []interface{}{
		map[string]interface{}{
			"domain":      "routing",
			"entity_name": "queue",
			"action_set":  schema.NewSet(schema.HashString, []interface{}{"add", "edit", "view"}),
		},
	}

	// Run the flatten function
	result := flattenRolePermissionPoliciesWithWildcardSuppress(apiPolicies, configuredPolicies)

	// Verify the result keeps explicit actions
	policies := result.List()
	if len(policies) != 1 {
		t.Fatalf("Expected 1 policy, got %d", len(policies))
	}

	policyMap := policies[0].(map[string]interface{})
	actionSet := policyMap["action_set"].(*schema.Set)
	actionList := actionSet.List()
	if len(actionList) != 3 {
		t.Fatalf("Expected action_set to have 3 elements, got %d: %v", len(actionList), actionList)
	}

	expectedActions := map[string]bool{"add": true, "edit": true, "view": true}
	for _, action := range actionList {
		if !expectedActions[action.(string)] {
			t.Errorf("Unexpected action in action_set: '%s'", action.(string))
		}
	}
}

// TestFlattenWildcardEntityNameSuppression verifies that when the user's config has
// entity_name = "*" and the API returns separate policies for each entity in the domain,
// the flatten function collapses them back into a single wildcard policy.
func TestFlattenWildcardEntityNameSuppression(t *testing.T) {
	// Simulate API response: multiple policies for the same domain (expanded from entity_name = "*")
	domain := "analytics"
	entity1 := "userObservation"
	entity2 := "conversationDetail"
	actions1 := []string{"view", "edit"}
	actions2 := []string{"view"}
	apiPolicies := []platformclientv2.Domainpermissionpolicy{
		{
			Domain:     &domain,
			EntityName: &entity1,
			ActionSet:  &actions1,
		},
		{
			Domain:     &domain,
			EntityName: &entity2,
			ActionSet:  &actions2,
		},
	}

	// Simulate user's config: entity_name = "*", action_set = ["*"]
	configuredPolicies := []interface{}{
		map[string]interface{}{
			"domain":      "analytics",
			"entity_name": "*",
			"action_set":  schema.NewSet(schema.HashString, []interface{}{"*"}),
		},
	}

	// Run the flatten function with wildcard suppression
	result := flattenRolePermissionPoliciesWithWildcardSuppress(apiPolicies, configuredPolicies)

	// Verify the result collapses back to a single policy with entity_name = "*"
	policies := result.List()
	if len(policies) != 1 {
		t.Fatalf("Expected 1 policy (collapsed wildcard), got %d", len(policies))
	}

	policyMap := policies[0].(map[string]interface{})

	if policyMap["domain"] != "analytics" {
		t.Errorf("Expected domain 'analytics', got '%s'", policyMap["domain"])
	}
	if policyMap["entity_name"] != "*" {
		t.Errorf("Expected entity_name '*', got '%s'", policyMap["entity_name"])
	}

	// Check action_set is also preserved as wildcard
	actionSet := policyMap["action_set"].(*schema.Set)
	actionList := actionSet.List()
	if len(actionList) != 1 {
		t.Fatalf("Expected action_set to have 1 element (wildcard), got %d: %v", len(actionList), actionList)
	}
	if actionList[0].(string) != "*" {
		t.Errorf("Expected action_set to be ['*'], got ['%s']", actionList[0].(string))
	}
}

// TestFlattenMixedPoliciesWithWildcard verifies that when the config has a mix of
// wildcard and explicit policies, only the wildcard ones are suppressed.
func TestFlattenMixedPoliciesWithWildcard(t *testing.T) {
	// Simulate API response
	domain1 := "directory"
	entity1 := "user"
	actions1 := []string{"add", "edit", "view", "delete"}

	domain2 := "routing"
	entity2 := "queue"
	actions2 := []string{"add", "edit", "view"}

	apiPolicies := []platformclientv2.Domainpermissionpolicy{
		{
			Domain:     &domain1,
			EntityName: &entity1,
			ActionSet:  &actions1,
		},
		{
			Domain:     &domain2,
			EntityName: &entity2,
			ActionSet:  &actions2,
		},
	}

	// Config: directory:user has wildcard, routing:queue has explicit actions
	configuredPolicies := []interface{}{
		map[string]interface{}{
			"domain":      "directory",
			"entity_name": "user",
			"action_set":  schema.NewSet(schema.HashString, []interface{}{"*"}),
		},
		map[string]interface{}{
			"domain":      "routing",
			"entity_name": "queue",
			"action_set":  schema.NewSet(schema.HashString, []interface{}{"add", "edit", "view"}),
		},
	}

	// Run the flatten function
	result := flattenRolePermissionPoliciesWithWildcardSuppress(apiPolicies, configuredPolicies)

	// Verify we get 2 policies
	policies := result.List()
	if len(policies) != 2 {
		t.Fatalf("Expected 2 policies, got %d", len(policies))
	}

	// Find each policy and verify
	for _, policy := range policies {
		policyMap := policy.(map[string]interface{})
		domain := policyMap["domain"].(string)
		actionSet := policyMap["action_set"].(*schema.Set)
		actionList := actionSet.List()

		switch domain {
		case "directory":
			// Should be collapsed to wildcard
			if len(actionList) != 1 || actionList[0].(string) != "*" {
				t.Errorf("Expected directory policy action_set to be ['*'], got %v", actionList)
			}
		case "routing":
			// Should remain explicit
			if len(actionList) != 3 {
				t.Errorf("Expected routing policy action_set to have 3 elements, got %d: %v", len(actionList), actionList)
			}
		default:
			t.Errorf("Unexpected domain: %s", domain)
		}
	}
}

// TestFlattenNoConfiguredPolicies verifies that when there are no configured policies
// (e.g., during import), the function behaves like the original flatten without suppression.
func TestFlattenNoConfiguredPolicies(t *testing.T) {
	// Simulate API response
	domain := "directory"
	entityName := "user"
	actions := []string{"add", "edit", "view", "delete"}
	apiPolicies := []platformclientv2.Domainpermissionpolicy{
		{
			Domain:     &domain,
			EntityName: &entityName,
			ActionSet:  &actions,
		},
	}

	// No configured policies (import scenario)
	var configuredPolicies []interface{}

	// Run the flatten function
	result := flattenRolePermissionPoliciesWithWildcardSuppress(apiPolicies, configuredPolicies)

	// Verify the result returns the API response as-is (no suppression)
	policies := result.List()
	if len(policies) != 1 {
		t.Fatalf("Expected 1 policy, got %d", len(policies))
	}

	policyMap := policies[0].(map[string]interface{})
	actionSet := policyMap["action_set"].(*schema.Set)
	actionList := actionSet.List()
	if len(actionList) != 4 {
		t.Fatalf("Expected action_set to have 4 elements, got %d: %v", len(actionList), actionList)
	}
}
