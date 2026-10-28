package agentic_virtual_agent_version

import (
	"encoding/json"
	"strings"
	"testing"
)

/*
   genesyscloud_agentic_virtual_agent_version_unit_test.go contains unit tests that exercise the
   package-private expand/flatten helpers and JSON serialization directly (no live API).

   Run with: TF_UNIT=1 go test ./... -run TestUnit
*/

// TestUnitExpandToolInputsDependencyOnly verifies the dependency_only field is read into the
// Go struct as a *bool and that it is omitted (nil) when not supplied — matching the API's
// "null when omitted" behavior for dependencyOnly.
func TestUnitExpandToolInputsDependencyOnly(t *testing.T) {
	// dependency_only = true, ToolOutput source (Knowledge-tool gating input)
	depInput := map[string]interface{}{
		"target_name":      "auth_token",
		"type":             "AuthResult",
		"source":           "ToolOutput",
		"required":         true,
		"fallback_to_user": false,
		"mapping":          `["AuthResult","token"]`,
		"dependency_only":  true,
	}
	// A plain input that never sets dependency_only → should stay nil.
	plainInput := map[string]interface{}{
		"target_name": "customerId",
		"type":        "CustomerId",
		"source":      "User",
	}

	got := expandToolInputs([]interface{}{depInput, plainInput})
	if len(got) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(got))
	}

	// First input: dependency_only true.
	if got[0].DependencyOnly == nil {
		t.Fatalf("expected DependencyOnly to be set for the first input, got nil")
	}
	if *got[0].DependencyOnly != true {
		t.Errorf("expected DependencyOnly=true, got %v", *got[0].DependencyOnly)
	}
	if got[0].Source != "ToolOutput" {
		t.Errorf("expected source ToolOutput, got %q", got[0].Source)
	}
	if len(got[0].Mapping) != 2 {
		t.Errorf("expected mapping of length 2, got %v", got[0].Mapping)
	}

	// Second input: the raw map has no dependency_only key, so expandToolInputs leaves the pointer
	// nil (the nil branch of the type assertion).
	//
	// NOTE: this is a synthetic case. Through real Terraform, schema.TypeBool is always present in
	// the block map (false when the user omits it), so expandToolInputs would set DependencyOnly to
	// &false — never nil. We build the map WITHOUT the key here only to cover the nil branch of the
	// expand code. On the wire via Terraform, every input carries an explicit dependencyOnly.
	if got[1].DependencyOnly != nil {
		t.Errorf("expected DependencyOnly to be nil when key absent from the raw map, got %v", *got[1].DependencyOnly)
	}
}

// TestUnitFlattenToolInputsDependencyOnly verifies the round-trip: expand → flatten preserves
// dependency_only, and that a nil *bool is not written back into state.
func TestUnitFlattenToolInputsDependencyOnly(t *testing.T) {
	tr := true
	inputs := []AgenticVirtualAgentToolInput{
		{
			TargetName:     "auth_token",
			Type:           "AuthResult",
			Source:         "ToolOutput",
			Mapping:        []interface{}{"AuthResult", "token"},
			DependencyOnly: &tr,
		},
		{
			TargetName: "customerId",
			Type:       "CustomerId",
			Source:     "User",
			// DependencyOnly nil → must be absent from the flattened map.
		},
	}

	flat := flattenToolInputs(inputs)
	if len(flat) != 2 {
		t.Fatalf("expected 2 flattened inputs, got %d", len(flat))
	}

	first := flat[0].(map[string]interface{})
	if v, ok := first["dependency_only"]; !ok {
		t.Errorf("expected dependency_only key present for first input")
	} else if v != true {
		t.Errorf("expected dependency_only=true, got %v", v)
	}

	second := flat[1].(map[string]interface{})
	if _, ok := second["dependency_only"]; ok {
		t.Errorf("expected dependency_only key absent when DependencyOnly is nil, but it was present")
	}
}

// TestUnitKnowledgeToolInputsMarshalToJSON verifies that a Knowledge tool (KnowledgeSetting)
// carrying a dependency-only input and inputValidation serializes with the exact JSON keys the
// public API expects: inputs[].dependencyOnly (bool) and inputValidation[] with if/else keys.
func TestUnitKnowledgeToolInputsMarshalToJSON(t *testing.T) {
	tr := true
	fls := false

	tool := AgenticVirtualAgentTool{
		Type:        "KnowledgeSetting",
		Name:        "Delivery FAQs Tool",
		Description: "Searches delivery FAQ knowledge.",
		Target:      DomainEntityRef{Id: "knowledge-setting-id", Name: "Delivery FAQs"},
		Inputs: []AgenticVirtualAgentToolInput{
			{
				TargetName:     "auth_token",
				Type:           "AuthResult",
				Source:         "ToolOutput",
				Required:       &tr,
				FallbackToUser: &fls,
				Mapping:        []interface{}{"AuthResult", "token"},
				DependencyOnly: &tr,
			},
		},
		InputValidation: []AgenticVirtualAgentInputValidation{
			{
				Type: "Structured",
				If:   json.RawMessage(`{"group":"And","rules":[{"name":"auth_token","operator":"IsNotNull"}]}`),
				Else: "Ask the customer to authenticate first.",
			},
		},
	}

	b, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal failed: %s", err)
	}
	out := string(b)

	// The dependencyOnly key must be present as a boolean true.
	if !strings.Contains(out, `"dependencyOnly":true`) {
		t.Errorf("expected serialized dependencyOnly:true, got: %s", out)
	}
	// inputs and inputValidation must serialize on the Knowledge tool.
	if !strings.Contains(out, `"inputs":`) {
		t.Errorf("expected inputs key in JSON, got: %s", out)
	}
	if !strings.Contains(out, `"inputValidation":`) {
		t.Errorf("expected inputValidation key in JSON, got: %s", out)
	}
	// inputValidation uses if/else wire keys.
	if !strings.Contains(out, `"if":`) || !strings.Contains(out, `"else":`) {
		t.Errorf("expected if/else keys in inputValidation JSON, got: %s", out)
	}
	// DataAction-only fields must NOT leak onto a Knowledge tool that didn't set them.
	if strings.Contains(out, `"output":`) {
		t.Errorf("did not expect output key on a Knowledge tool with no output, got: %s", out)
	}
}

// TestUnitToolInputDependencyOnlyOmittedFromJSONWhenNil verifies that an input without
// DependencyOnly set does not emit the dependencyOnly key at all (nullable/omitempty behavior).
func TestUnitToolInputDependencyOnlyOmittedFromJSONWhenNil(t *testing.T) {
	input := AgenticVirtualAgentToolInput{
		TargetName: "customerId",
		Type:       "CustomerId",
		Source:     "User",
	}
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %s", err)
	}
	if strings.Contains(string(b), "dependencyOnly") {
		t.Errorf("expected dependencyOnly to be omitted when nil, got: %s", string(b))
	}
}
