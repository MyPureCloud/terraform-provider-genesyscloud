package agentic_virtual_agent_version

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func TestUnitMcpContractSchema(t *testing.T) {
	resourceSchema := ResourceAgenticVirtualAgentVersion()
	if err := resourceSchema.InternalValidate(nil, true); err != nil {
		t.Fatalf("AVA version resource schema is invalid: %v", err)
	}

	assertSchemaValueAccepted(t, toolResource().Schema["type"], "MCP")
	assertSchemaValueAccepted(t, typeDefinitionResource().Schema["type"], "McpError")
	assertSchemaValueAccepted(t, typeDefinitionResource().Schema["type"], "A2AError")

	inputSchema := toolInputResource().Schema
	if _, ok := inputSchema["fallback_to_failure"]; !ok {
		t.Fatal("expected tool inputs to expose fallback_to_failure")
	}

	errorCodesSchema, ok := typeDefinitionResource().Schema["error_codes"]
	if !ok {
		t.Fatal("expected type definitions to expose error_codes")
	}
	errorCodeItemSchema, ok := errorCodesSchema.Elem.(*schema.Schema)
	if !ok {
		t.Fatalf("expected error_codes elements to use *schema.Schema, got %T", errorCodesSchema.Elem)
	}
	for _, code := range []string{
		"ToolError",
		"EmptyToolResponse",
		"RemoteMcpError",
		"RemoteAuth",
		"InvalidRequest",
		"RateLimited",
		"InternalError",
		"McpTimeout",
		"AgentUnavailable",
	} {
		assertSchemaValueAccepted(t, errorCodeItemSchema, code)
	}
	if _, errors := errorCodeItemSchema.ValidateFunc("NotAnMcpErrorCode", "error_codes.0"); len(errors) == 0 {
		t.Fatal("expected an unknown MCP error code to fail validation")
	}
}

func TestUnitMcpToolContractRoundTrip(t *testing.T) {
	retained := true
	tools := expandTools([]interface{}{
		map[string]interface{}{
			"type":        "MCP",
			"name":        "Lookup Order Status",
			"description": "Looks up the current status of an order.",
			"target": []interface{}{
				map[string]interface{}{
					"id":   "mcp-integration-id",
					"name": "get_order_status",
				},
			},
			"input_instructions": []interface{}{"Use this for an existing order."},
			"inputs": []interface{}{
				map[string]interface{}{
					"target_name":         "tenant_token",
					"type":                "TenantToken",
					"source":              "External",
					"required":            true,
					"fallback_to_failure": true,
				},
			},
			"output": "OrderStatus",
			"input_validation": []interface{}{
				map[string]interface{}{
					"type":             "Python",
					"if_condition":     "tenant_token is not None",
					"else_instruction": "End the session because authentication context is missing.",
				},
			},
			"errors": []interface{}{
				map[string]interface{}{
					"type":        "OrderLookupFailed",
					"instruction": "Apologise and offer to retry later.",
				},
			},
		},
	}, []*bool{&retained})

	if len(tools) != 1 {
		t.Fatalf("expected one MCP tool, got %d", len(tools))
	}
	tool := tools[0]
	if tool.Type != "MCP" {
		t.Fatalf("expected MCP tool type, got %q", tool.Type)
	}
	if tool.Target.Id != "mcp-integration-id" || tool.Target.Name != "get_order_status" {
		t.Fatalf("unexpected MCP target: %#v", tool.Target)
	}
	if tool.ContextRetentionEnabled == nil || !*tool.ContextRetentionEnabled {
		t.Fatal("expected contextRetentionEnabled=true")
	}
	if len(tool.Inputs) != 1 || tool.Inputs[0].FallbackToFailure == nil || !*tool.Inputs[0].FallbackToFailure {
		t.Fatalf("expected fallbackToFailure=true on the MCP input: %#v", tool.Inputs)
	}
	if tool.Output != "OrderStatus" {
		t.Fatalf("expected OrderStatus output, got %q", tool.Output)
	}
	if len(tool.InputValidation) != 1 || len(tool.Errors) != 1 {
		t.Fatalf("expected MCP validation and error handling, got %#v", tool)
	}

	types := expandTypes([]interface{}{
		map[string]interface{}{
			"name":                "OrderLookupFailed",
			"description":         "The order lookup tool failed.",
			"type":                "McpError",
			"error_codes":         []interface{}{"ToolError", "RemoteMcpError", "McpTimeout", "RateLimited"},
			"default_instruction": "Tell the customer the order system is unavailable.",
		},
	})
	if len(types) != 1 || !reflect.DeepEqual(types[0].ErrorCodes, []string{"ToolError", "RemoteMcpError", "McpTimeout", "RateLimited"}) {
		t.Fatalf("unexpected MCP error type: %#v", types)
	}

	request := AgenticVirtualAgentVersionCreate{Definition: &AgenticVirtualAgentVersionDefinition{
		Role:         "You help customers with orders.",
		Instructions: []string{"Use the MCP tool for order status."},
		Tools:        tools,
		Types:        types,
	}}
	actualJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal MCP request: %v", err)
	}
	assertJSONEqual(t, string(actualJSON), `{
		"definition": {
			"role": "You help customers with orders.",
			"instructions": ["Use the MCP tool for order status."],
			"tools": [{
				"type": "MCP",
				"name": "Lookup Order Status",
				"description": "Looks up the current status of an order.",
				"target": {"id": "mcp-integration-id", "name": "get_order_status"},
				"inputInstructions": ["Use this for an existing order."],
				"contextRetentionEnabled": true,
				"errors": [{"type": "OrderLookupFailed", "instruction": "Apologise and offer to retry later."}],
				"inputs": [{
					"targetName": "tenant_token",
					"type": "TenantToken",
					"source": "External",
					"required": true,
					"fallbackToFailure": true
				}],
				"output": "OrderStatus",
				"inputValidation": [{
					"type": "Python",
					"if": "tenant_token is not None",
					"else": "End the session because authentication context is missing."
				}]
			}],
			"types": [{
				"name": "OrderLookupFailed",
				"description": "The order lookup tool failed.",
				"type": "McpError",
				"errorCodes": ["ToolError", "RemoteMcpError", "McpTimeout", "RateLimited"],
				"defaultInstruction": "Tell the customer the order system is unavailable."
			}]
		}
	}`)

	flattenedTool := flattenTools(tools)[0].(map[string]interface{})
	flattenedInput := flattenedTool["inputs"].([]interface{})[0].(map[string]interface{})
	if value, ok := flattenedInput["fallback_to_failure"]; !ok || value != true {
		t.Fatalf("expected flattened fallback_to_failure=true, got %v (present=%t)", value, ok)
	}
	flattenedType := flattenTypes(types)[0].(map[string]interface{})
	if !reflect.DeepEqual(flattenedType["error_codes"], []string{"ToolError", "RemoteMcpError", "McpTimeout", "RateLimited"}) {
		t.Fatalf("unexpected flattened MCP error codes: %#v", flattenedType["error_codes"])
	}
}

func TestUnitMcpSchemaMetadataIsReadOnly(t *testing.T) {
	var tool AgenticVirtualAgentTool
	if err := json.Unmarshal([]byte(`{
		"type": "MCP",
		"name": "Lookup Order Status",
		"description": "Looks up an order.",
		"target": {
			"id": "mcp-integration-id",
			"name": "get_order_status",
			"selfUri": "/api/v2/integrations/mcp/mcp-integration-id/tools/get_order_status"
		},
		"contextRetentionEnabled": false,
		"schemas": {
			"inputs": {"type": "object", "required": ["order_number"]}
		}
	}`), &tool); err != nil {
		t.Fatalf("failed to unmarshal MCP response: %v", err)
	}
	if tool.Schemas == nil || tool.Schemas.Inputs["type"] != "object" || tool.Schemas.Outputs != nil {
		t.Fatalf("unexpected MCP schema metadata: %#v", tool.Schemas)
	}
	if tool.ContextRetentionEnabled == nil || *tool.ContextRetentionEnabled {
		t.Fatalf("expected contextRetentionEnabled=false, got %#v", tool.ContextRetentionEnabled)
	}

	flattened := flattenTools([]AgenticVirtualAgentTool{tool})[0].(map[string]interface{})
	if _, ok := flattened["schemas"]; ok {
		t.Fatal("server-populated MCP schemas must not be flattened into configurable Terraform state")
	}
	if value, ok := flattened["context_retention_enabled"]; !ok || value != false {
		t.Fatalf("expected flattened context_retention_enabled=false, got %v (present=%t)", value, ok)
	}
}

func TestUnitToolInputFallbackToFailureFalseAndNilHandling(t *testing.T) {
	falseValue := false
	flattened := flattenToolInputs([]AgenticVirtualAgentToolInput{
		{
			TargetName:        "token",
			Type:              "TenantToken",
			Source:            "External",
			FallbackToFailure: &falseValue,
		},
		{
			TargetName: "order_number",
			Type:       "OrderNumber",
			Source:     "User",
		},
	})

	first := flattened[0].(map[string]interface{})
	if value, ok := first["fallback_to_failure"]; !ok || value != false {
		t.Fatalf("expected API fallbackToFailure=false to flatten into state, got %v (present=%t)", value, ok)
	}
	second := flattened[1].(map[string]interface{})
	if _, ok := second["fallback_to_failure"]; ok {
		t.Fatal("expected nil fallbackToFailure to remain absent when flattened")
	}

	expanded := expandToolInputs(flattened)
	if expanded[0].FallbackToFailure != nil {
		t.Fatal("expected fallback_to_failure=false to be omitted from the API request")
	}
	if expanded[1].FallbackToFailure != nil {
		t.Fatal("expected an absent fallback_to_failure to remain absent from the API request")
	}
}

func TestUnitValidateFallbackToFailureInputs(t *testing.T) {
	tests := []struct {
		name              string
		source            string
		required          bool
		fallbackToUser    bool
		fallbackToFailure bool
		includeType       bool
		typeValue         string
		undisclosed       bool
		wantError         string
	}{
		{
			name:              "valid external undisclosed input",
			source:            "External",
			required:          true,
			fallbackToFailure: true,
			includeType:       true,
			typeValue:         "string",
			undisclosed:       true,
		},
		{
			name:              "false is ignored",
			source:            "User",
			fallbackToFailure: false,
		},
		{
			name:              "mutually exclusive user fallback",
			source:            "External",
			required:          true,
			fallbackToUser:    true,
			fallbackToFailure: true,
			includeType:       true,
			typeValue:         "string",
			undisclosed:       true,
			wantError:         "cannot be true when fallback_to_user is true",
		},
		{
			name:              "external source required",
			source:            "User",
			required:          true,
			fallbackToFailure: true,
			includeType:       true,
			typeValue:         "string",
			undisclosed:       true,
			wantError:         "requires source to be External",
		},
		{
			name:              "required input required",
			source:            "External",
			fallbackToFailure: true,
			includeType:       true,
			typeValue:         "string",
			undisclosed:       true,
			wantError:         "requires required to be true",
		},
		{
			name:              "referenced type required",
			source:            "External",
			required:          true,
			fallbackToFailure: true,
			wantError:         "to be defined as an undisclosed string type",
		},
		{
			name:              "undisclosed type required",
			source:            "External",
			required:          true,
			fallbackToFailure: true,
			includeType:       true,
			typeValue:         "string",
			wantError:         "to be an undisclosed string type",
		},
		{
			name:              "string type required",
			source:            "External",
			required:          true,
			fallbackToFailure: true,
			includeType:       true,
			typeValue:         "object",
			undisclosed:       true,
			wantError:         "to be an undisclosed string type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{
						"inputs": []interface{}{
							map[string]interface{}{
								"target_name":         "tenant_token",
								"type":                "TenantToken",
								"source":              test.source,
								"required":            test.required,
								"fallback_to_user":    test.fallbackToUser,
								"fallback_to_failure": test.fallbackToFailure,
							},
						},
					},
				},
			}
			if test.includeType {
				definition["types"] = []interface{}{
					map[string]interface{}{
						"name":        "TenantToken",
						"type":        test.typeValue,
						"undisclosed": test.undisclosed,
					},
				}
			}

			err := validateFallbackToFailureInputs([]interface{}{definition})
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("expected valid fallback_to_failure configuration: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("expected error containing %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestUnitBuildDefinitionOnlySendsTrueFallbackToFailure(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceAgenticVirtualAgentVersion().Schema, map[string]interface{}{
		"agent_id": "agent-id",
		"definition": []interface{}{
			map[string]interface{}{
				"role":         "You help customers with orders.",
				"instructions": []interface{}{"Use the order tool."},
				"tools": []interface{}{
					map[string]interface{}{
						"type":        "MCP",
						"name":        "Lookup Order Status",
						"description": "Looks up an order.",
						"target": []interface{}{
							map[string]interface{}{"id": "mcp-integration-id", "name": "get_order_status"},
						},
						"inputs": []interface{}{
							map[string]interface{}{
								"target_name": "order_number",
								"type":        "OrderNumber",
								"source":      "User",
								"required":    true,
							},
							map[string]interface{}{
								"target_name":         "tenant_token",
								"type":                "TenantToken",
								"source":              "External",
								"required":            true,
								"fallback_to_failure": true,
							},
						},
					},
				},
				"types": []interface{}{
					map[string]interface{}{"name": "OrderNumber", "type": "string"},
					map[string]interface{}{"name": "TenantToken", "type": "string", "undisclosed": true},
				},
			},
		},
	})

	definition := buildDefinitionFromResourceData(d)
	if definition.Tools[0].Inputs[0].FallbackToFailure != nil {
		t.Fatal("an omitted fallback_to_failure must not become false in the API request")
	}
	if definition.Tools[0].Inputs[1].FallbackToFailure == nil || !*definition.Tools[0].Inputs[1].FallbackToFailure {
		t.Fatal("expected configured fallback_to_failure=true in the API request")
	}

	payload, err := json.Marshal(AgenticVirtualAgentVersionCreate{Definition: definition})
	if err != nil {
		t.Fatalf("failed to marshal version request: %v", err)
	}
	var request map[string]interface{}
	if err := json.Unmarshal(payload, &request); err != nil {
		t.Fatalf("failed to inspect version request: %v", err)
	}
	requestDefinition := request["definition"].(map[string]interface{})
	requestTool := requestDefinition["tools"].([]interface{})[0].(map[string]interface{})
	requestInputs := requestTool["inputs"].([]interface{})
	if _, exists := requestInputs[0].(map[string]interface{})["fallbackToFailure"]; exists {
		t.Fatalf("unexpected fallbackToFailure on User input: %s", payload)
	}
	if value, exists := requestInputs[1].(map[string]interface{})["fallbackToFailure"]; !exists || value != true {
		t.Fatalf("expected fallbackToFailure=true on External input: %s", payload)
	}
	if _, exists := requestTool["schemas"]; exists {
		t.Fatalf("server-populated schemas must not be emitted in a create request: %s", payload)
	}
	requestTarget := requestTool["target"].(map[string]interface{})
	if _, exists := requestTarget["selfUri"]; exists {
		t.Fatalf("server-populated target selfUri must not be emitted in a create request: %s", payload)
	}
}

func assertSchemaValueAccepted(t *testing.T, valueSchema *schema.Schema, value string) {
	t.Helper()
	if valueSchema.ValidateFunc == nil {
		t.Fatalf("schema for %q has no validation function", value)
	}
	if _, errors := valueSchema.ValidateFunc(value, "test"); len(errors) > 0 {
		t.Fatalf("expected schema to accept %q: %v", value, errors)
	}
}

func assertJSONEqual(t *testing.T, actual, expected string) {
	t.Helper()
	var actualValue interface{}
	if err := json.Unmarshal([]byte(actual), &actualValue); err != nil {
		t.Fatalf("actual JSON is invalid: %v\n%s", err, actual)
	}
	var expectedValue interface{}
	if err := json.Unmarshal([]byte(expected), &expectedValue); err != nil {
		t.Fatalf("expected JSON is invalid: %v\n%s", err, expected)
	}
	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Fatalf("JSON mismatch\nactual:   %s\nexpected: %s", actual, expected)
	}
}
