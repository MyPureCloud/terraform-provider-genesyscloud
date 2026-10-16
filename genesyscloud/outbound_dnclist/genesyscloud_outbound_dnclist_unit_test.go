//go:build unit
// +build unit

package outbound_dnclist

import "testing"

// TestUnitShouldReconcileEntries verifies the entries-reconciliation decision
// across the feature-toggle states (DEVTOOLING-1816). The key safety property
// is that with the unmanaged-entries toggle ON, an omitted or empty entries
// block must NOT trigger reconciliation (which would delete all phone numbers),
// while a populated block still does.
func TestUnitShouldReconcileEntries(t *testing.T) {
	cases := []struct {
		name            string
		unmanagedToggle bool
		hasChange       bool
		entryCount      int
		want            bool
	}{
		// Toggle OFF: original behavior keyed on HasChange.
		{"off_change_populated", false, true, 2, true},
		{"off_change_omitted_or_empty", false, true, 0, true},
		{"off_no_change", false, false, 2, false},

		// Toggle ON: keyed on populated config only; never on absence.
		{"on_omitted_or_empty_no_reconcile", true, true, 0, false},
		{"on_populated_reconciles", true, true, 2, true},
		{"on_populated_even_without_haschange", true, false, 2, true},
		{"on_empty_even_with_haschange", true, true, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldReconcileEntries(tc.unmanagedToggle, tc.hasChange, tc.entryCount)
			if got != tc.want {
				t.Errorf("shouldReconcileEntries(toggle=%v, hasChange=%v, count=%d) = %v, want %v",
					tc.unmanagedToggle, tc.hasChange, tc.entryCount, got, tc.want)
			}
		})
	}
}
