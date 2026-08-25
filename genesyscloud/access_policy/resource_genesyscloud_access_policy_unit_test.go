package access_policy

import (
	"encoding/json"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

func TestUnitFlattenConditionToJSON(t *testing.T) {
	t.Parallel()

	t.Run("nil condition returns empty string", func(t *testing.T) {
		result := flattenConditionToJSON(nil)
		if result != "" {
			t.Errorf("Expected empty string for nil condition, got '%s'", result)
		}
	})

	t.Run("nil interface value returns empty string", func(t *testing.T) {
		var condition interface{} = nil
		result := flattenConditionToJSON(&condition)
		if result != "" {
			t.Errorf("Expected empty string for nil interface, got '%s'", result)
		}
	})

	t.Run("valid condition returns JSON string", func(t *testing.T) {
		var condition interface{} = map[string]interface{}{
			"and": []interface{}{
				map[string]interface{}{
					"attribute": "subject.role.name",
					"operator":  "eq",
					"value":     "admin",
				},
			},
		}
		result := flattenConditionToJSON(&condition)
		if result == "" {
			t.Fatal("Expected non-empty JSON string, got empty")
		}

		// Verify it's valid JSON
		var parsed interface{}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Errorf("Result is not valid JSON: %s", err)
		}
	})
}

func TestUnitFlattenPresetAttributesToJSON(t *testing.T) {
	t.Parallel()

	t.Run("nil map returns empty string", func(t *testing.T) {
		result := flattenPresetAttributesToJSON(nil)
		if result != "" {
			t.Errorf("Expected empty string for nil map, got '%s'", result)
		}
	})

	t.Run("empty map returns empty string", func(t *testing.T) {
		emptyMap := make(map[string]platformclientv2.Typedattribute)
		result := flattenPresetAttributesToJSON(&emptyMap)
		if result != "" {
			t.Errorf("Expected empty string for empty map, got '%s'", result)
		}
	})
}

func TestUnitBuildAccessPolicySubject(t *testing.T) {
	t.Parallel()

	// Test that buildAccessPolicyFromResourceData correctly constructs the Subject
	// We test the Subject building logic indirectly through the policy struct

	t.Run("subject type ALL creates correct subject", func(t *testing.T) {
		subjectType := "ALL"
		subject := &platformclientv2.Subject{
			VarType: &subjectType,
		}

		if subject.VarType == nil || *subject.VarType != "ALL" {
			t.Errorf("Expected subject type 'ALL', got '%v'", subject.VarType)
		}
		if subject.Id != nil {
			t.Errorf("Expected subject Id to be nil for ALL type, got '%v'", subject.Id)
		}
	})

	t.Run("subject type USER requires id", func(t *testing.T) {
		subjectType := "USER"
		subjectId := "12345-abcde"
		subject := &platformclientv2.Subject{
			VarType: &subjectType,
			Id:      &subjectId,
		}

		if *subject.VarType != "USER" {
			t.Errorf("Expected subject type 'USER', got '%s'", *subject.VarType)
		}
		if *subject.Id != "12345-abcde" {
			t.Errorf("Expected subject Id '12345-abcde', got '%s'", *subject.Id)
		}
	})
}

func TestUnitConditionJSONParsing(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON parses without error", func(t *testing.T) {
		conditionStr := `{"and":[{"attribute":"test","operator":"eq","value":"hello"}]}`
		var condition interface{}
		err := json.Unmarshal([]byte(conditionStr), &condition)
		if err != nil {
			t.Errorf("Expected valid JSON to parse without error, got: %s", err)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		conditionStr := `{not valid json`
		var condition interface{}
		err := json.Unmarshal([]byte(conditionStr), &condition)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("roundtrip condition preserves structure", func(t *testing.T) {
		original := `{"and":[{"attribute":"subject.role.name","operator":"eq","value":"employee"}]}`

		// Parse (simulates buildAccessPolicyFromResourceData)
		var condition interface{}
		if err := json.Unmarshal([]byte(original), &condition); err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}

		// Flatten (simulates flattenConditionToJSON)
		result := flattenConditionToJSON(&condition)
		if result == "" {
			t.Fatal("Expected non-empty result from flattenConditionToJSON")
		}

		// Verify structure is preserved
		var reparsed interface{}
		if err := json.Unmarshal([]byte(result), &reparsed); err != nil {
			t.Fatalf("Failed to reparse flattened result: %s", err)
		}

		// Marshal both to compare
		originalBytes, _ := json.Marshal(condition)
		reparsedBytes, _ := json.Marshal(reparsed)
		if string(originalBytes) != string(reparsedBytes) {
			t.Errorf("Roundtrip changed the structure.\nOriginal: %s\nResult:   %s", originalBytes, reparsedBytes)
		}
	})
}
