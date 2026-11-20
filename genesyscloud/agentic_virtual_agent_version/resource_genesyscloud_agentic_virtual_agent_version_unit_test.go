package agentic_virtual_agent_version

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func TestUnitAgenticVirtualAgentVersionExportCache(t *testing.T) {
	cache := newAgenticVirtualAgentVersionExportCache()
	resources := resourceExporter.ResourceIDMetaMap{
		"agent-1/1.0": &resourceExporter.ResourceMeta{},
		"agent-2/2.0": &resourceExporter.ResourceMeta{},
	}

	if err := cache.replace(resources); err != nil {
		t.Fatalf("failed to populate export cache: %v", err)
	}
	if versionID, ok := cache.versionID("agent-1"); !ok || versionID != "1.0" {
		t.Fatalf("expected agent-1 to map to version 1.0, got %q (found=%t)", versionID, ok)
	}

	if err := cache.replace(resourceExporter.ResourceIDMetaMap{"agent-1/3.0": &resourceExporter.ResourceMeta{}}); err != nil {
		t.Fatalf("failed to replace export cache: %v", err)
	}
	if _, ok := cache.versionID("agent-2"); ok {
		t.Fatal("expected replacing the cache to remove stale agent entries")
	}
	if versionID, ok := cache.versionID("agent-1"); !ok || versionID != "3.0" {
		t.Fatalf("expected agent-1 to map to version 3.0, got %q (found=%t)", versionID, ok)
	}

	if err := cache.replace(resourceExporter.ResourceIDMetaMap{"invalid-resource-id": &resourceExporter.ResourceMeta{}}); err == nil {
		t.Fatal("expected invalid resource ID to fail cache replacement")
	}
	if versionID, ok := cache.versionID("agent-1"); !ok || versionID != "3.0" {
		t.Fatalf("expected failed replacement to preserve the previous cache, got %q (found=%t)", versionID, ok)
	}
}

func TestUnitBuildDefinitionContextRetentionEnabled(t *testing.T) {
	tests := []struct {
		name       string
		rawConfig  cty.Value
		stateValue string
		want       *bool
	}{
		{
			name:       "removed from configuration",
			rawConfig:  cty.NullVal(cty.Bool),
			stateValue: "true",
		},
		{
			name:       "explicit true",
			rawConfig:  cty.True,
			stateValue: "true",
			want:       boolPtr(true),
		},
		{
			name:       "explicit false",
			rawConfig:  cty.False,
			stateValue: "false",
			want:       boolPtr(false),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := schema.InternalMap(ResourceAgenticVirtualAgentVersion().Schema).Data(
				testAgenticVirtualAgentState(test.stateValue, test.rawConfig), nil,
			)
			if err != nil {
				t.Fatalf("failed to construct ResourceData: %v", err)
			}

			definition := buildDefinitionFromResourceData(d)
			got := definition.Tools[0].ContextRetentionEnabled
			if test.want == nil {
				if got != nil {
					t.Fatalf("expected context retention to remain unset, got %t", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected context retention to be %t, got nil", *test.want)
			}
			if *got != *test.want {
				t.Fatalf("expected context retention to be %t, got %t", *test.want, *got)
			}
		})
	}
}

func TestUnitToolContextRetentionValuesReadsRawConfig(t *testing.T) {
	values := toolContextRetentionValues([]cty.Value{
		testContextRetentionToolConfig(cty.NullVal(cty.Bool)),
		testContextRetentionToolConfig(cty.True),
		testContextRetentionToolConfig(cty.False),
		testContextRetentionToolConfig(cty.UnknownVal(cty.Bool)),
	})

	if len(values) != 4 {
		t.Fatalf("expected four context retention values, got %d", len(values))
	}
	if values[0] != nil {
		t.Fatalf("expected omitted value to be nil, got %t", *values[0])
	}
	if values[1] == nil || !*values[1] {
		t.Fatalf("expected explicit true value")
	}
	if values[2] == nil || *values[2] {
		t.Fatalf("expected explicit false value")
	}
	if values[3] != nil {
		t.Fatalf("expected unknown value to be nil, got %t", *values[3])
	}
}

func TestUnitContextRetentionPresenceMarkerFromRawConfig(t *testing.T) {
	tests := []struct {
		name      string
		rawConfig cty.Value
		want      string
		known     bool
	}{
		{
			name: "explicit true and false and omitted",
			rawConfig: testContextRetentionRawConfig(
				[]cty.Value{cty.True, cty.False, cty.NullVal(cty.Bool)},
			),
			want:  "110",
			known: true,
		},
		{
			name: "unknown value",
			rawConfig: testContextRetentionRawConfig(
				[]cty.Value{cty.UnknownVal(cty.Bool)},
			),
			known: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, known := contextRetentionPresenceMarkerFromRawConfig(test.rawConfig)
			if known != test.known {
				t.Fatalf("expected known=%t, got %t", test.known, known)
			}
			if got != test.want {
				t.Fatalf("expected marker %q, got %q", test.want, got)
			}
		})
	}
}

func TestUnitContextRetentionPresenceChangesCreateDiff(t *testing.T) {
	falseValue := false
	tests := []struct {
		name             string
		stateMarker      string
		rawConfigValue   cty.Value
		configuredValue  *bool
		wantMarkerChange bool
	}{
		{
			name:             "explicit false to omitted",
			stateMarker:      "1",
			rawConfigValue:   cty.NullVal(cty.Bool),
			configuredValue:  nil,
			wantMarkerChange: true,
		},
		{
			name:             "omitted to explicit false",
			stateMarker:      "0",
			rawConfigValue:   cty.False,
			configuredValue:  &falseValue,
			wantMarkerChange: true,
		},
		{
			name:             "omitted remains omitted",
			stateMarker:      "0",
			rawConfigValue:   cty.NullVal(cty.Bool),
			configuredValue:  nil,
			wantMarkerChange: false,
		},
		{
			name:             "explicit false remains explicit false",
			stateMarker:      "1",
			rawConfigValue:   cty.False,
			configuredValue:  &falseValue,
			wantMarkerChange: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := testAgenticVirtualAgentState("false", test.rawConfigValue)
			state.Attributes[contextRetentionPresenceAttribute] = test.stateMarker

			diff, err := ResourceAgenticVirtualAgentVersion().Diff(
				context.Background(),
				state,
				terraform.NewResourceConfigRaw(testAgenticVirtualAgentConfig(test.configuredValue)),
				nil,
			)
			if err != nil {
				t.Fatalf("failed to calculate diff: %v", err)
			}

			markerChanged := false
			if diff != nil {
				_, markerChanged = diff.GetAttribute(contextRetentionPresenceAttribute)
			}
			if markerChanged != test.wantMarkerChange {
				t.Fatalf("expected marker change=%t, got %t; diff=%#v", test.wantMarkerChange, markerChanged, diff)
			}
		})
	}
}

func TestUnitExpandToolsMatchesContextRetentionToToolPosition(t *testing.T) {
	prunable := false
	tools := expandTools(
		[]interface{}{
			testDataActionToolMap("FirstTool"),
			testDataActionToolMap("SecondTool"),
			testDataActionToolMap("ThirdTool"),
		},
		[]*bool{nil, &prunable, nil},
	)

	if len(tools) != 3 {
		t.Fatalf("expected three tools, got %d", len(tools))
	}
	if tools[0].ContextRetentionEnabled != nil {
		t.Fatalf("expected first tool context retention to be unset")
	}
	if tools[1].ContextRetentionEnabled == nil || *tools[1].ContextRetentionEnabled {
		t.Fatalf("expected second tool context retention to be false")
	}
	if tools[2].ContextRetentionEnabled != nil {
		t.Fatalf("expected third tool context retention to be unset")
	}
}

func TestUnitOmittedContextRetentionPaths(t *testing.T) {
	disabled := false
	definition := &AgenticVirtualAgentVersionDefinition{
		Tools: []AgenticVirtualAgentTool{
			{ContextRetentionEnabled: nil},
			{ContextRetentionEnabled: &disabled},
			{ContextRetentionEnabled: nil},
		},
	}

	got := omittedContextRetentionPaths(definition)
	want := []string{
		"definition.0.tools.0.context_retention_enabled",
		"definition.0.tools.2.context_retention_enabled",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d omitted paths, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("expected omitted path %q at index %d, got %q", want[i], i, got[i])
		}
	}
}

func testDataActionToolMap(name string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "DataAction",
		"name":        name,
		"description": "Looks up a customer.",
		"target": []interface{}{
			map[string]interface{}{
				"id":   "action-id",
				"name": "LookupCustomerAction",
			},
		},
	}
}

func testContextRetentionToolConfig(value cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"context_retention_enabled": value,
	})
}

func testContextRetentionRawConfig(values []cty.Value) cty.Value {
	tools := make([]cty.Value, 0, len(values))
	for _, value := range values {
		tools = append(tools, testContextRetentionToolConfig(value))
	}

	return cty.ObjectVal(map[string]cty.Value{
		"definition": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"tools": cty.ListVal(tools),
			}),
		}),
	})
}

func testAgenticVirtualAgentConfig(contextRetentionEnabled *bool) map[string]interface{} {
	tool := testDataActionToolMap("LookupCustomer")
	if contextRetentionEnabled != nil {
		tool["context_retention_enabled"] = *contextRetentionEnabled
	}

	return map[string]interface{}{
		"agent_id": "agent-id",
		"definition": []interface{}{
			map[string]interface{}{
				"role":         "A helpful assistant.",
				"instructions": []interface{}{"Be helpful."},
				"tools":        []interface{}{tool},
			},
		},
	}
}

func testAgenticVirtualAgentState(contextRetentionState string, rawContextRetention cty.Value) *terraform.InstanceState {
	return &terraform.InstanceState{
		ID: "agent-id/1.0",
		Attributes: map[string]string{
			"agent_id":                                       "agent-id",
			"definition.#":                                   "1",
			"definition.0.role":                              "A helpful assistant.",
			"definition.0.instructions.#":                    "1",
			"definition.0.instructions.0":                    "Be helpful.",
			"definition.0.tools.#":                           "1",
			"definition.0.tools.0.type":                      "DataAction",
			"definition.0.tools.0.name":                      "LookupCustomer",
			"definition.0.tools.0.description":               "Looks up a customer.",
			"definition.0.tools.0.target.#":                  "1",
			"definition.0.tools.0.target.0.id":               "action-id",
			"definition.0.tools.0.target.0.name":             "LookupCustomerAction",
			"definition.0.tools.0.context_retention_enabled": contextRetentionState,
		},
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"definition": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"tools": cty.ListVal([]cty.Value{testContextRetentionToolConfig(rawContextRetention)}),
				}),
			}),
		}),
	}
}

func TestUnitFlattenToolsContextRetentionEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value *bool
	}{
		{name: "unset"},
		{name: "true", value: boolPtr(true)},
		{name: "false", value: boolPtr(false)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flattened := flattenTools([]AgenticVirtualAgentTool{{
				Type:                    "DataAction",
				Name:                    "LookupCustomer",
				Description:             "Looks up a customer.",
				Target:                  DomainEntityRef{Id: "action-id"},
				ContextRetentionEnabled: test.value,
			}})
			tool := flattened[0].(map[string]interface{})
			got, exists := tool["context_retention_enabled"]

			if test.value == nil {
				if exists {
					t.Fatalf("expected unset context retention to be omitted, got %v", got)
				}
				return
			}
			if !exists {
				t.Fatalf("expected context retention to be flattened")
			}
			if got != *test.value {
				t.Fatalf("expected context retention to be %t, got %v", *test.value, got)
			}
		})
	}
}

func TestUnitAgenticVirtualAgentToolContextRetentionJSON(t *testing.T) {
	tests := []struct {
		name  string
		value *bool
	}{
		{name: "unset"},
		{name: "true", value: boolPtr(true)},
		{name: "false", value: boolPtr(false)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := json.Marshal(AgenticVirtualAgentTool{
				ContextRetentionEnabled: test.value,
			})
			if err != nil {
				t.Fatalf("failed to marshal tool: %v", err)
			}

			var decoded map[string]interface{}
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatalf("failed to unmarshal tool: %v", err)
			}

			got, exists := decoded["contextRetentionEnabled"]
			if test.value == nil {
				if exists {
					t.Fatalf("expected unset context retention to be omitted, got %v", got)
				}
				return
			}
			if !exists {
				t.Fatalf("expected contextRetentionEnabled JSON property")
			}
			if got != *test.value {
				t.Fatalf("expected contextRetentionEnabled to be %t, got %v", *test.value, got)
			}
		})
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func TestUnitRawConfigDefinitionIsWhollyKnown(t *testing.T) {
	if !rawConfigDefinitionIsWhollyKnown(testFallbackValidationRawConfig(cty.StringVal("External"))) {
		t.Fatal("expected a fully known nested definition to be validated")
	}
	if rawConfigDefinitionIsWhollyKnown(testFallbackValidationRawConfig(cty.UnknownVal(cty.String))) {
		t.Fatal("expected an unknown nested input source to defer validation")
	}
	if rawConfigDefinitionIsWhollyKnown(cty.NullVal(cty.Bool)) {
		t.Fatal("expected a null raw configuration to defer validation")
	}
}

func testFallbackValidationRawConfig(source cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"definition": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"tools": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"inputs": cty.ListVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"fallback_to_failure": cty.True,
								"source":              source,
							}),
						}),
					}),
				}),
			}),
		}),
	})
}

func TestUnitCustomizeDiffValidatesKnownFallbackToFailureInput(t *testing.T) {
	config := map[string]interface{}{
		"agent_id": "agent-id",
		"definition": []interface{}{
			map[string]interface{}{
				"role":         "You help customers with orders.",
				"instructions": []interface{}{"Use the MCP tool."},
				"tools": []interface{}{
					map[string]interface{}{
						"type":        "MCP",
						"name":        "LookupOrder",
						"description": "Looks up an order.",
						"target": []interface{}{
							map[string]interface{}{"id": "integration-id", "name": "get_order"},
						},
						"inputs": []interface{}{
							map[string]interface{}{
								"target_name":         "tenant_token",
								"type":                "TenantToken",
								"source":              "User",
								"required":            true,
								"fallback_to_failure": true,
							},
						},
					},
				},
				"types": []interface{}{
					map[string]interface{}{
						"name":        "TenantToken",
						"type":        "string",
						"undisclosed": true,
					},
				},
			},
		},
	}

	_, err := ResourceAgenticVirtualAgentVersion().Diff(
		context.Background(),
		&terraform.InstanceState{RawConfig: testFallbackValidationRawConfig(cty.StringVal("User"))},
		terraform.NewResourceConfigRaw(config),
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "requires source to be External") {
		t.Fatalf("expected CustomizeDiff to reject a known invalid source, got %v", err)
	}

	_, err = ResourceAgenticVirtualAgentVersion().Diff(
		context.Background(),
		&terraform.InstanceState{RawConfig: testFallbackValidationRawConfig(cty.UnknownVal(cty.String))},
		terraform.NewResourceConfigRaw(config),
		nil,
	)
	if err != nil {
		t.Fatalf("expected CustomizeDiff to defer validation for an unknown nested source, got %v", err)
	}
}

const testComfortStatementDefault = "Before the tool call, provide one filler line in under eight words naming only the action being performed."

func TestUnitBuiltinInputInstructionJSON(t *testing.T) {
	override := "Let the user know you're fetching their order status."
	tool := AgenticVirtualAgentTool{
		Type: "DataAction",
		BuiltinInputInstructions: []AgenticVirtualAgentBuiltinInputInstruction{
			{Type: builtinInputInstructionTypeComfortStatement, Index: 0, Default: testComfortStatementDefault},
			{Type: "Other", Index: 2, Default: "d", Override: &override},
		},
	}

	payload, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded struct {
		BuiltinInputInstructions []map[string]interface{} `json:"builtinInputInstructions"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.BuiltinInputInstructions) != 2 {
		t.Fatalf("expected 2 builtinInputInstructions, got %s", payload)
	}

	first := decoded.BuiltinInputInstructions[0]
	if index, ok := first["index"]; !ok || index != float64(0) {
		t.Errorf("expected index 0 to be serialized, got %s", payload)
	}
	if _, ok := first["override"]; ok {
		t.Errorf("expected override to be omitted when unset, got %s", payload)
	}
	if first["type"] != builtinInputInstructionTypeComfortStatement || first["default"] != testComfortStatementDefault {
		t.Errorf("unexpected first built-in: %v", first)
	}

	if decoded.BuiltinInputInstructions[1]["override"] != override {
		t.Errorf("expected override to be serialized, got %s", payload)
	}

	payload, _ = json.Marshal(AgenticVirtualAgentTool{Type: "DataAction"})
	if strings.Contains(string(payload), "builtinInputInstructions") {
		t.Errorf("expected builtinInputInstructions to be omitted when empty, got %s", payload)
	}
}

func TestUnitExpandFlattenBuiltinInputInstructionsRoundTrip(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"type":     builtinInputInstructionTypeComfortStatement,
			"index":    1,
			"default":  testComfortStatementDefault,
			"override": "",
		},
	}

	expanded := expandBuiltinInputInstructions(input)
	if len(expanded) != 1 {
		t.Fatalf("expected 1 built-in, got %d", len(expanded))
	}
	if expanded[0].Index != 1 || expanded[0].Type != builtinInputInstructionTypeComfortStatement || expanded[0].Default != testComfortStatementDefault {
		t.Errorf("unexpected expanded built-in: %+v", expanded[0])
	}
	if expanded[0].Override != nil {
		t.Errorf("expected empty override to expand to nil, got %q", *expanded[0].Override)
	}

	override := "Tell the user you're checking their order."
	expanded[0].Override = &override
	flattened := flattenBuiltinInputInstructions(expanded)
	m := flattened[0].(map[string]interface{})
	if m["index"] != 1 || m["type"] != builtinInputInstructionTypeComfortStatement || m["default"] != testComfortStatementDefault || m["override"] != override {
		t.Errorf("unexpected flattened built-in: %v", m)
	}

	expanded[0].Override = nil
	m = flattenBuiltinInputInstructions(expanded)[0].(map[string]interface{})
	if _, ok := m["override"]; ok {
		t.Errorf("expected nil override to be absent from state, got %v", m["override"])
	}
}

func TestUnitExpandToolsIncludesBuiltinInputInstructions(t *testing.T) {
	toolMap := testDataActionToolMap("GetOrder")
	toolMap["input_instructions"] = []interface{}{"Call this when the user asks about their order."}
	toolMap["builtin_input_instructions"] = []interface{}{
		map[string]interface{}{"type": builtinInputInstructionTypeComfortStatement, "index": 0, "default": testComfortStatementDefault},
	}

	tools := expandTools([]interface{}{toolMap}, nil)
	if len(tools[0].BuiltinInputInstructions) != 1 {
		t.Fatalf("expected the built-in to be expanded onto the tool, got %+v", tools[0])
	}
	if len(tools[0].InputInstructions) != 1 {
		t.Errorf("expected built-ins to stay separate from input_instructions, got %v", tools[0].InputInstructions)
	}
}

func TestUnitBuildDefinitionDropsV1ComfortStatementWithV2(t *testing.T) {
	tests := []struct {
		name      string
		withV2    bool
		wantV1Set bool
	}{
		{name: "V1 kept without a V2 comfort statement", withV2: false, wantV1Set: true},
		{name: "V1 dropped with a V2 comfort statement", withV2: true, wantV1Set: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool := testDataActionToolMap("GetOrder")
			if test.withV2 {
				tool["builtin_input_instructions"] = []interface{}{
					map[string]interface{}{"type": builtinInputInstructionTypeComfortStatement, "index": 0, "default": testComfortStatementDefault},
				}
			}
			d := schema.TestResourceDataRaw(t, ResourceAgenticVirtualAgentVersion().Schema, map[string]interface{}{
				"agent_id": "agent-id",
				"definition": []interface{}{
					map[string]interface{}{
						"role":         "A helpful assistant.",
						"instructions": []interface{}{"Be helpful."},
						"tools":        []interface{}{tool},
						"settings": []interface{}{
							map[string]interface{}{
								"comfort_statement": []interface{}{map[string]interface{}{"enabled": true}},
							},
						},
					},
				},
			})

			definition := buildDefinitionFromResourceData(d)
			if definition.Settings == nil {
				t.Fatalf("expected settings to be built")
			}
			if gotV1Set := definition.Settings.ComfortStatement != nil; gotV1Set != test.wantV1Set {
				t.Fatalf("expected V1 comfort statement set=%t, got %t", test.wantV1Set, gotV1Set)
			}
		})
	}
}

func TestUnitValidateComfortStatementConfiguration(t *testing.T) {
	v2Tool := cty.ObjectVal(map[string]cty.Value{
		"builtin_input_instructions": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{"type": cty.StringVal(builtinInputInstructionTypeComfortStatement)}),
		}),
	})
	plainTool := cty.ObjectVal(map[string]cty.Value{
		"builtin_input_instructions": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"type": cty.String}))),
	})
	unknownTypeTool := cty.ObjectVal(map[string]cty.Value{
		"builtin_input_instructions": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{"type": cty.UnknownVal(cty.String)}),
		}),
	})
	v1Settings := cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"comfort_statement": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"enabled": cty.True})}),
	})})
	noV1Settings := cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"comfort_statement": cty.ListValEmpty(cty.Object(map[string]cty.Type{"enabled": cty.Bool})),
	})})

	tests := []struct {
		name     string
		tool     cty.Value
		settings cty.Value
		wantErr  bool
	}{
		{name: "V2 and V1 together", tool: v2Tool, settings: v1Settings, wantErr: true},
		{name: "V2 only", tool: v2Tool, settings: noV1Settings},
		{name: "V1 only", tool: plainTool, settings: v1Settings},
		{name: "unknown built-in type is not checked", tool: unknownTypeTool, settings: v1Settings},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rawConfig := cty.ObjectVal(map[string]cty.Value{
				"definition": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"tools":    cty.ListVal([]cty.Value{test.tool}),
					"settings": test.settings,
				})}),
			})
			err := validateComfortStatementConfiguration(rawConfig)
			if (err != nil) != test.wantErr {
				t.Fatalf("expected error=%t, got %v", test.wantErr, err)
			}
		})
	}
}

func TestUnitComfortStatementConflictFailsPlan(t *testing.T) {
	tool := testDataActionToolMap("GetOrder")
	tool["builtin_input_instructions"] = []interface{}{
		map[string]interface{}{"type": builtinInputInstructionTypeComfortStatement, "index": 0, "default": testComfortStatementDefault},
	}
	config := map[string]interface{}{
		"agent_id": "agent-id",
		"definition": []interface{}{
			map[string]interface{}{
				"role":         "A helpful assistant.",
				"instructions": []interface{}{"Be helpful."},
				"tools":        []interface{}{tool},
				"settings": []interface{}{
					map[string]interface{}{
						"comfort_statement": []interface{}{map[string]interface{}{"enabled": true}},
					},
				},
			},
		},
	}

	priorState := &terraform.InstanceState{
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"definition": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
				"tools": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"builtin_input_instructions": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{"type": cty.StringVal(builtinInputInstructionTypeComfortStatement)}),
					}),
				})}),
				"settings": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"comfort_statement": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"enabled": cty.True})}),
				})}),
			})}),
		}),
	}

	_, err := ResourceAgenticVirtualAgentVersion().Diff(context.Background(), priorState, terraform.NewResourceConfigRaw(config), nil)
	if err == nil || !strings.Contains(err.Error(), "settings.comfort_statement cannot be set") {
		t.Fatalf("expected a comfort statement conflict error, got %v", err)
	}
}

func TestUnitHasContextualComfortStatement(t *testing.T) {
	if hasContextualComfortStatement(nil) {
		t.Error("expected no tools to report false")
	}
	if hasContextualComfortStatement([]AgenticVirtualAgentTool{{Type: "DataAction"}}) {
		t.Error("expected a tool without built-ins to report false")
	}
	tools := []AgenticVirtualAgentTool{
		{Type: "DataAction"},
		{Type: "KnowledgeBase", BuiltinInputInstructions: []AgenticVirtualAgentBuiltinInputInstruction{{Type: builtinInputInstructionTypeComfortStatement}}},
	}
	if !hasContextualComfortStatement(tools) {
		t.Error("expected a ComfortStatement built-in on any tool to report true")
	}
}
