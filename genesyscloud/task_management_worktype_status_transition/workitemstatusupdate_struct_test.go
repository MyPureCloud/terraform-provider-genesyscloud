package task_management_worktype_status_transition

import (
	"encoding/json"
	"testing"
)

func TestWorkitemstatusupdateMarshalJSONClearsOptionalFields(t *testing.T) {
	body := &Workitemstatusupdate{}
	body.SetField("Name", strPtr("Open"))
	body.SetField("DefaultDestinationStatusId", nil)
	body.SetField("StatusTransitionDelaySeconds", nil)
	body.SetField("StatusTransitionTime", nil)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if result["name"] != "Open" {
		t.Fatalf("expected name to be set, got %#v", result["name"])
	}
	if _, ok := result["defaultDestinationStatusId"]; !ok {
		t.Fatalf("expected defaultDestinationStatusId to be present in payload")
	}
	if result["defaultDestinationStatusId"] != nil {
		t.Fatalf("expected defaultDestinationStatusId to be null, got %#v", result["defaultDestinationStatusId"])
	}
}

func TestWorkitemstatusupdateMarshalJSONExplicitFalseAutoTerminate(t *testing.T) {
	falseValue := false
	body := &Workitemstatusupdate{}
	body.SetField("AutoTerminateWorkitem", &falseValue)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if result["autoTerminateWorkitem"] != false {
		t.Fatalf("expected autoTerminateWorkitem false, got %#v", result["autoTerminateWorkitem"])
	}
}

func strPtr(value string) *string {
	return &value
}
