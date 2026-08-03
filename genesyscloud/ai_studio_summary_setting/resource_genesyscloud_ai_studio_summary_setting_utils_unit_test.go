package ai_studio_summary_setting

import "testing"

func TestBuildSummarySettingPIIs_PreservesFalse(t *testing.T) {
	out := buildSummarySettingPIIs([]interface{}{
		map[string]interface{}{
			"all": false,
		},
	})

	if out == nil {
		t.Fatalf("expected non-nil Summarysettingpii")
	}
	if out.All == nil {
		t.Fatalf("expected All to be non-nil when set to false")
	}
	if *out.All != false {
		t.Fatalf("expected All=false, got %v", *out.All)
	}
}
