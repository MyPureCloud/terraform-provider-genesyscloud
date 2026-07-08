package task_management_worktype_status

import (
	"fmt"
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func TestUnitWorktypeStatusRefResolver_HappyPath(t *testing.T) {
	worktypeId := "wt-111"
	statusId := "st-222"
	compositeId := worktypeId + "/" + statusId
	blockLabel := "MyWorktype_MyStatus"

	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
				compositeId: &resourceExporter.ResourceMeta{BlockLabel: blockLabel},
			},
		},
	}

	configMap := map[string]interface{}{
		"status_id": statusId,
	}

	resolver := WorktypeStatusRefResolver("status_id")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := fmt.Sprintf("${%s.%s.id}", ResourceType, blockLabel)
	if configMap["status_id"] != expected {
		t.Errorf("expected %s, got %s", expected, configMap["status_id"])
	}
}

func TestUnitWorktypeStatusRefResolver_AlreadyResolved(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{},
		},
	}

	configMap := map[string]interface{}{
		"status_id": "${genesyscloud_task_management_worktype_status.MyStatus.id}",
	}

	resolver := WorktypeStatusRefResolver("status_id")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should remain unchanged
	if configMap["status_id"] != "${genesyscloud_task_management_worktype_status.MyStatus.id}" {
		t.Errorf("already-resolved value was modified: %s", configMap["status_id"])
	}
}

func TestUnitWorktypeStatusRefResolver_NilExporter(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{}

	configMap := map[string]interface{}{
		"status_id": "st-222",
	}

	resolver := WorktypeStatusRefResolver("status_id")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should remain unchanged when exporter is missing
	if configMap["status_id"] != "st-222" {
		t.Errorf("value was modified when exporter is nil: %s", configMap["status_id"])
	}
}

func TestUnitWorktypeStatusRefResolver_NotFound(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
				"wt-111/st-999": &resourceExporter.ResourceMeta{BlockLabel: "OtherStatus"},
			},
		},
	}

	configMap := map[string]interface{}{
		"status_id": "st-not-in-map",
	}

	resolver := WorktypeStatusRefResolver("status_id")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should remain unchanged when not found
	if configMap["status_id"] != "st-not-in-map" {
		t.Errorf("value was modified when ID not found: %s", configMap["status_id"])
	}
}

func TestUnitWorktypeStatusRefResolver_NilValue(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{},
		},
	}

	configMap := map[string]interface{}{
		"status_id": nil,
	}

	resolver := WorktypeStatusRefResolver("status_id")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if configMap["status_id"] != nil {
		t.Errorf("nil value was modified: %v", configMap["status_id"])
	}
}

func TestUnitWorktypeStatusRefResolver_MissingAttribute(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{},
		},
	}

	configMap := map[string]interface{}{}

	resolver := WorktypeStatusRefResolver("status_id")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- WorktypeStatusArrayRefResolver tests ---

func TestUnitWorktypeStatusArrayRefResolver_HappyPath(t *testing.T) {
	worktypeId := "wt-111"
	statusId1 := "st-aaa"
	statusId2 := "st-bbb"

	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
				worktypeId + "/" + statusId1: &resourceExporter.ResourceMeta{BlockLabel: "Worktype_StatusA"},
				worktypeId + "/" + statusId2: &resourceExporter.ResourceMeta{BlockLabel: "Worktype_StatusB"},
			},
		},
	}

	configMap := map[string]interface{}{
		"destination_status_ids": []interface{}{statusId1, statusId2},
	}

	resolver := WorktypeStatusArrayRefResolver("destination_status_ids")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := configMap["destination_status_ids"].([]interface{})
	expected1 := fmt.Sprintf("${%s.%s.id}", ResourceType, "Worktype_StatusA")
	expected2 := fmt.Sprintf("${%s.%s.id}", ResourceType, "Worktype_StatusB")

	if arr[0] != expected1 {
		t.Errorf("expected %s, got %s", expected1, arr[0])
	}
	if arr[1] != expected2 {
		t.Errorf("expected %s, got %s", expected2, arr[1])
	}
}

func TestUnitWorktypeStatusArrayRefResolver_AlreadyResolved(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{},
		},
	}

	alreadyResolved := "${genesyscloud_task_management_worktype_status.MyStatus.id}"
	configMap := map[string]interface{}{
		"destination_status_ids": []interface{}{alreadyResolved},
	}

	resolver := WorktypeStatusArrayRefResolver("destination_status_ids")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := configMap["destination_status_ids"].([]interface{})
	if arr[0] != alreadyResolved {
		t.Errorf("already-resolved value was modified: %s", arr[0])
	}
}

func TestUnitWorktypeStatusArrayRefResolver_NilExporter(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{}

	configMap := map[string]interface{}{
		"destination_status_ids": []interface{}{"st-aaa", "st-bbb"},
	}

	resolver := WorktypeStatusArrayRefResolver("destination_status_ids")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := configMap["destination_status_ids"].([]interface{})
	if arr[0] != "st-aaa" || arr[1] != "st-bbb" {
		t.Errorf("values were modified when exporter is nil: %v", arr)
	}
}

func TestUnitWorktypeStatusArrayRefResolver_PartialResolution(t *testing.T) {
	worktypeId := "wt-111"
	statusId1 := "st-aaa"
	statusId2 := "st-not-found"

	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
				worktypeId + "/" + statusId1: &resourceExporter.ResourceMeta{BlockLabel: "Worktype_StatusA"},
			},
		},
	}

	configMap := map[string]interface{}{
		"destination_status_ids": []interface{}{statusId1, statusId2},
	}

	resolver := WorktypeStatusArrayRefResolver("destination_status_ids")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := configMap["destination_status_ids"].([]interface{})
	expected1 := fmt.Sprintf("${%s.%s.id}", ResourceType, "Worktype_StatusA")

	if arr[0] != expected1 {
		t.Errorf("expected %s, got %s", expected1, arr[0])
	}
	if arr[1] != statusId2 {
		t.Errorf("expected unresolved ID %s, got %s", statusId2, arr[1])
	}
}

func TestUnitWorktypeStatusArrayRefResolver_NilArray(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{},
		},
	}

	configMap := map[string]interface{}{
		"destination_status_ids": nil,
	}

	resolver := WorktypeStatusArrayRefResolver("destination_status_ids")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if configMap["destination_status_ids"] != nil {
		t.Errorf("nil array was modified: %v", configMap["destination_status_ids"])
	}
}

func TestUnitWorktypeStatusArrayRefResolver_EmptyArray(t *testing.T) {
	exporters := map[string]*resourceExporter.ResourceExporter{
		ResourceType: {
			SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{},
		},
	}

	configMap := map[string]interface{}{
		"destination_status_ids": []interface{}{},
	}

	resolver := WorktypeStatusArrayRefResolver("destination_status_ids")
	err := resolver(configMap, exporters, "test_label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := configMap["destination_status_ids"].([]interface{})
	if len(arr) != 0 {
		t.Errorf("empty array was modified: %v", arr)
	}
}
