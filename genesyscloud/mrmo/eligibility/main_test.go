package eligibility

import (
	"testing"
)

// TestUnit_Evaluate_InvalidInput verifies the public Evaluate entry point rejects
// empty input with an error diagnostic (mirrors exporter input validation).
func TestUnit_Evaluate_InvalidInput(t *testing.T) {
	resp, diags := Evaluate(EvaluateInput{ResourceType: ""}, Config{})

	if !diags.HasError() {
		t.Fatal("expected an error diagnostic for empty ResourceType")
	}
	if resp != nil {
		t.Errorf("expected nil response on error, got %+v", resp)
	}
}

// TestUnit_Evaluate_ValidInputShape verifies Evaluate returns a populated
// EvaluateOutput for a syntactically valid input. We use a type we know is not
// registered so the result is deterministic (TYPE_NOT_SUPPORTED) without needing
// the real registry to contain any particular type.
func TestUnit_Evaluate_ValidInputShape(t *testing.T) {
	resp, diags := Evaluate(EvaluateInput{ResourceType: "genesyscloud_this_type_does_not_exist"}, Config{})

	if diags.HasError() {
		t.Fatalf("did not expect an error diagnostic, got %v", diags)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if resp.Result.ResourceType != "genesyscloud_this_type_does_not_exist" {
		t.Errorf("unexpected ResourceType %q", resp.Result.ResourceType)
	}
	if resp.Result.Status != StatusNotApplicable || resp.Result.ReasonCode != ReasonTypeNotSupported {
		t.Errorf("expected NotApplicable/TYPE_NOT_SUPPORTED, got %s/%s",
			resp.Result.Status, resp.Result.ReasonCode)
	}
}

// TestUnit_EvaluateAll_Subset verifies the public EvaluateAll entry point returns a
// report for the requested subset. Both requested types are unregistered, so both
// resolve to TYPE_NOT_SUPPORTED and are flagged.
func TestUnit_EvaluateAll_PublicSubset(t *testing.T) {
	resp, diags := EvaluateAll(EvaluateAllInput{
		ResourceTypes: []string{"genesyscloud_fake_a", "genesyscloud_fake_b"},
	}, Config{})

	if diags.HasError() {
		t.Fatalf("did not expect an error diagnostic, got %v", diags)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if resp.Report.TotalEvaluated != 2 {
		t.Errorf("expected TotalEvaluated=2, got %d", resp.Report.TotalEvaluated)
	}
	if len(resp.Report.Flagged) != 2 {
		t.Errorf("expected 2 flagged (both unsupported), got %d", len(resp.Report.Flagged))
	}
	if resp.Report.CountByStatus[StatusNotApplicable] != 2 {
		t.Errorf("expected 2 NotApplicable, got %d", resp.Report.CountByStatus[StatusNotApplicable])
	}
}
