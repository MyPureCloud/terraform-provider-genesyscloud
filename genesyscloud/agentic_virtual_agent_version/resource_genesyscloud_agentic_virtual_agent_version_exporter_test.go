package agentic_virtual_agent_version

import (
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func TestAgenticVirtualAgentVersionExporter_ToolTargetResolverConfigured(t *testing.T) {
	exporter := AgenticVirtualAgentVersionExporter()
	if exporter == nil || exporter.CustomAttributeResolver == nil {
		t.Fatal("expected exporter CustomAttributeResolver to be defined")
	}

	targetResolver, ok := exporter.CustomAttributeResolver["definition.tools.target.id"]
	if !ok || targetResolver == nil || targetResolver.ResolveRefTypeFunc == nil {
		t.Fatal("expected CustomAttributeResolver for definition.tools.target.id with ResolveRefTypeFunc")
	}

	toolsResolver, ok := exporter.CustomAttributeResolver["definition.tools"]
	if !ok || toolsResolver == nil || toolsResolver.ResolverFunc == nil {
		t.Fatal("expected CustomAttributeResolver for definition.tools with ResolverFunc")
	}
}

func TestResolveToolTargetRefType(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		wantType string
		wantErr  bool
	}{
		{
			name: "knowledge base from self_uri",
			config: map[string]interface{}{
				"id":       "a7fdcc0e-adb9-4250-8656-8194f4ed7e6b",
				"self_uri": "/api/v2/knowledge/knowledgebases/a7fdcc0e-adb9-4250-8656-8194f4ed7e6b",
			},
			wantType: knowledgeKnowledgebaseResourceType,
		},
		{
			name: "data action from self_uri",
			config: map[string]interface{}{
				"id":       "custom_-_aaaa-1111-bbbb-2222",
				"self_uri": "/api/v2/integrations/actions/custom_-_aaaa-1111-bbbb-2222",
			},
			wantType: integrationActionResourceType,
		},
		{
			name: "custom data action from id prefix",
			config: map[string]interface{}{
				"id": "custom_-_aaaa-1111-bbbb-2222",
			},
			wantType: integrationActionResourceType,
		},
		{
			name: "static data action from id prefix",
			config: map[string]interface{}{
				"id": "static-get-user-guid",
			},
			wantType: integrationActionResourceType,
		},
		{
			name: "unknown uuid without self_uri",
			config: map[string]interface{}{
				"id": "a7fdcc0e-adb9-4250-8656-8194f4ed7e6b",
			},
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := resolveToolTargetRefType(tt.config)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotType != tt.wantType {
				t.Fatalf("got ref type %q, want %q", gotType, tt.wantType)
			}
		})
	}
}

func TestResolveAgenticToolTargets(t *testing.T) {
	kbID := "a7fdcc0e-adb9-4250-8656-8194f4ed7e6b"
	actionID := "custom_-_aaaa-1111-bbbb-2222"

	exporters := map[string]*resourceExporter.ResourceExporter{
		knowledgeKnowledgebaseResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
				kbID: {BlockLabel: "example_kb"},
			},
		},
		integrationActionResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
				actionID: {BlockLabel: "lookup_customer"},
			},
		},
	}

	configMap := map[string]interface{}{
		"tools": []interface{}{
			map[string]interface{}{
				"type": "KnowledgeBase",
				"target": []interface{}{
					map[string]interface{}{
						"id": kbID,
					},
				},
			},
			map[string]interface{}{
				"type": "DataAction",
				"target": []interface{}{
					map[string]interface{}{
						"id": actionID,
					},
				},
			},
		},
	}

	if err := resolveAgenticToolTargets(configMap, exporters, "test_label"); err != nil {
		t.Fatalf("resolveAgenticToolTargets returned error: %v", err)
	}

	tools := configMap["tools"].([]interface{})

	kbTarget := tools[0].(map[string]interface{})["target"].([]interface{})[0].(map[string]interface{})
	if kbTarget["id"] != "${genesyscloud_knowledge_knowledgebase.example_kb.id}" {
		t.Fatalf("knowledge base target id = %q, want resource reference", kbTarget["id"])
	}

	actionTarget := tools[1].(map[string]interface{})["target"].([]interface{})[0].(map[string]interface{})
	if actionTarget["id"] != "${genesyscloud_integration_action.lookup_customer.id}" {
		t.Fatalf("data action target id = %q, want resource reference", actionTarget["id"])
	}
}
