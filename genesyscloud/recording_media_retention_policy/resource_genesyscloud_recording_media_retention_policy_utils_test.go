package recording_media_retention_policy

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

// Regression test: flattenCallMediaPolicyConditions must carry the duration condition through,
// otherwise exporting a call media policy silently drops it.
func TestUnitFlattenCallMediaPolicyConditionsDuration(t *testing.T) {
	target := "DURATION"
	operator := "GREATER_THAN"
	durationRange := "300000"
	mode := "Over"

	conditions := &platformclientv2.Callmediapolicyconditions{
		Duration: &platformclientv2.Durationcondition{
			DurationTarget:   &target,
			DurationOperator: &operator,
			DurationRange:    &durationRange,
			DurationMode:     &mode,
		},
	}

	flattened := flattenCallMediaPolicyConditions(conditions, nil)
	if len(flattened) != 1 {
		t.Fatalf("expected 1 flattened conditions block, got %d", len(flattened))
	}

	conditionsMap := flattened[0].(map[string]interface{})
	durationList, ok := conditionsMap["duration"].([]interface{})
	if !ok || len(durationList) != 1 {
		t.Fatalf("expected duration to be flattened into a 1-item list, got %#v", conditionsMap["duration"])
	}

	durationMap := durationList[0].(map[string]interface{})
	expected := map[string]string{
		"duration_target":   target,
		"duration_operator": operator,
		"duration_range":    durationRange,
		"duration_mode":     mode,
	}
	for key, want := range expected {
		if got := durationMap[key]; got != want {
			t.Errorf("duration.%s = %v, want %v", key, got, want)
		}
	}
}

func TestUnitFlattenCallMediaPolicyConditionsNoDuration(t *testing.T) {
	flattened := flattenCallMediaPolicyConditions(&platformclientv2.Callmediapolicyconditions{}, nil)
	if len(flattened) != 1 {
		t.Fatalf("expected 1 flattened conditions block, got %d", len(flattened))
	}
	if _, ok := flattened[0].(map[string]interface{})["duration"]; ok {
		t.Errorf("duration must not be set when the API returns none")
	}
}
