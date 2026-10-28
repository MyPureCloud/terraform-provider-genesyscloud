package agentic_virtual_agent_version

import (
	"context"
	"encoding/json"
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
