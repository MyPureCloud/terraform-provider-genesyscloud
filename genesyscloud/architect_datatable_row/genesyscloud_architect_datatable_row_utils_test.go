package architect_datatable_row

import "testing"

// TestExtractFilterPatterns checks that only this resource type's filter entries are extracted with the prefix stripped.
func TestExtractFilterPatterns(t *testing.T) {
	tests := []struct {
		name     string
		filter   []string
		expected []string
	}{
		{
			name:     "empty filter",
			filter:   nil,
			expected: nil,
		},
		{
			name:     "single matching pattern",
			filter:   []string{ResourceType + "::Prioridades_ATR_AGENTES"},
			expected: []string{"Prioridades_ATR_AGENTES"},
		},
		{
			name: "mixed resource types",
			filter: []string{
				"genesyscloud_architect_datatable::SomeTable",
				ResourceType + "::Prioridades_ATR_AGENTES",
				"genesyscloud_user::someuser",
			},
			expected: []string{"Prioridades_ATR_AGENTES"},
		},
		{
			name:     "prefix present but empty pattern is skipped",
			filter:   []string{ResourceType + "::"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFilterPatterns(ResourceType, tt.filter)
			if len(got) != len(tt.expected) {
				t.Fatalf("extractFilterPatterns() = %v, want %v", got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("extractFilterPatterns()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// TestTableMatchesFilter locks in DEVTOOLING-1789: unanchored patterns must keep the table; only anchored patterns for other tables may skip.
func TestTableMatchesFilter(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		patterns  []string
		want      bool
	}{
		// Regression cases: must keep the table (true).
		{
			// Reported case: bare row key that itself contains underscores.
			name:      "bare row key with underscores is kept",
			tableName: "Prioridades",
			patterns:  []string{"ATR_AGENTES"},
			want:      true,
		},
		{
			name:      "bare row key without underscore is kept",
			tableName: "Prioridades",
			patterns:  []string{"somekey"},
			want:      true,
		},
		{
			name:      "row key matching another table name is kept",
			tableName: "Prioridades",
			patterns:  []string{"OtherTable_key"},
			want:      true,
		},
		{
			// The ".*" workaround must still work.
			name:      "leading wildcard is kept",
			tableName: "Prioridades",
			patterns:  []string{".*ATR_AGENTES"},
			want:      true,
		},
		{
			name:      "alternation regex is kept",
			tableName: "Prioridades",
			patterns:  []string{"(Prioridades|Other)_key"},
			want:      true,
		},
		{
			name:      "full literal label spanning table and key is kept",
			tableName: "Prioridades",
			patterns:  []string{"Prioridades_ATR_AGENTES"},
			want:      true,
		},
		{
			name:      "partial table name is kept",
			tableName: "Prioridades",
			patterns:  []string{"Prio"},
			want:      true,
		},
		{
			// Fragment straddling table name and row key.
			name:      "substring spanning table and key is kept",
			tableName: "Prioridades",
			patterns:  []string{"dades_ATR"},
			want:      true,
		},
		{
			name:      "character class regex is kept",
			tableName: "Prioridades",
			patterns:  []string{"row_[0-9]+"},
			want:      true,
		},
		{
			// End-anchored only ($) has no leading ^ literal, so keep.
			name:      "end anchored pattern is kept",
			tableName: "Prioridades",
			patterns:  []string{"ATR_AGENTES$"},
			want:      true,
		},
		{
			// "^" immediately followed by "(" yields no literal prefix, so keep.
			name:      "anchor followed by alternation is kept",
			tableName: "Prioridades",
			patterns:  []string{"^(Prioridades|Other)_key"},
			want:      true,
		},
		{
			// "^" immediately followed by "." yields no literal prefix, so keep.
			name:      "anchor followed by wildcard is kept",
			tableName: "Prioridades",
			patterns:  []string{"^.*ATR_AGENTES"},
			want:      true,
		},
		{
			// "^" immediately followed by "[" yields no literal prefix, so keep.
			name:      "anchor followed by character class is kept",
			tableName: "Prioridades",
			patterns:  []string{"^[PD]rioridades"},
			want:      true,
		},
		{
			name:      "anchored to this table prefix is kept",
			tableName: "Prioridades",
			patterns:  []string{"^Prioridades_ATR_AGENTES"},
			want:      true,
		},
		{
			name:      "anchored partial table name is kept",
			tableName: "Prioridades",
			patterns:  []string{"^Prio"},
			want:      true,
		},
		{
			name:      "any one matching pattern keeps the table",
			tableName: "Prioridades",
			patterns:  []string{"^Other_key", "somekey"},
			want:      true,
		},
		{
			// include_filter_resources_by_id passes "<tableGuid>/<key>" through
			// this pre-filter; it is never ^-anchored, so all tables are kept
			// and FilterResourceById does the actual id matching.
			name:      "by-id pattern (guid/key) is kept",
			tableName: "Prioridades",
			patterns:  []string{"c1d2e3f4-0000-0000-0000-000000000000/row_0000"},
			want:      true,
		},
		{
			// Space in name: anchored to sanitized prefix "My_Table_".
			name:      "anchored to sanitized prefix (space in name) is kept",
			tableName: "My Table",
			patterns:  []string{"^My_Table_row"},
			want:      true,
		},
		{
			// Leading digit: sanitizer prefixes "_" -> "_123_Table_".
			name:      "anchored to sanitized prefix (leading digit) is kept",
			tableName: "123 Table",
			patterns:  []string{"^_123_Table_row"},
			want:      true,
		},
		{
			// Same table, anchored against the raw form.
			name:      "anchored to raw prefix (leading digit) is kept",
			tableName: "123 Table",
			patterns:  []string{"^123"},
			want:      true,
		},

		// Optimization cases: may safely skip the table (false).
		{
			name:      "anchored to a different table is skipped",
			tableName: "Prioridades",
			patterns:  []string{"^OtherTable_key"},
			want:      false,
		},
		{
			name:      "anchored to a different table (all anchored) is skipped",
			tableName: "Prioridades",
			patterns:  []string{"^DecoyTable_row1", "^AnotherTable_row2"},
			want:      false,
		},
		{
			name:      "no patterns skips",
			tableName: "Prioridades",
			patterns:  []string{},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tableMatchesFilter(tt.tableName, tt.patterns)
			if got != tt.want {
				t.Errorf("tableMatchesFilter(%q, %v) = %v, want %v", tt.tableName, tt.patterns, got, tt.want)
			}
		})
	}
}
