package validators

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestValidateCSVFormatWithConfig(t *testing.T) {
	tests := []struct {
		name          string
		csv           any
		opts          ValidateCSVOptions
		expectedError bool
		errorMessage  string
	}{
		{
			name: "Valid CSV with required columns",
			csv: `id,name,value
1,test1,val1
2,test2,val2
3,test3,val3`,
			opts: ValidateCSVOptions{
				RequiredColumns: []string{"id", "name"},
				SampleSize:      10,
			},
			expectedError: false,
		},
		{
			name: "Missing required column",
			csv: `id,value
1,val1
2,val2`,
			opts: ValidateCSVOptions{
				RequiredColumns: []string{"id", "name"},
				SampleSize:      10,
			},
			expectedError: true,
			errorMessage:  "required column 'name' not found in CSV",
		},
		{
			name: "Inconsistent number of fields",
			csv: `id,name,value
1,test1
2,test2,val2`,
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "error reading line 1: record on line 2: wrong number of fields",
		},
		{
			name: "Empty CSV",
			csv:  "",
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "failed to read CSV headers:",
		},
		{
			name: "CSV exceeds max row count",
			csv: `id,name
1,test1
2,test2
3,test3`,
			opts: ValidateCSVOptions{
				MaxRowCount: 2,
				SampleSize:  10,
			},
			expectedError: true,
			errorMessage:  "CSV file exceeds maximum allowed rows of 2",
		},
		{
			name: "Valid CSV with sampling",
			csv: `id,name
1,test1
2,test2
3,test3
4,test4
5,test5`,
			opts: ValidateCSVOptions{
				SampleSize:   2,
				SkipInterval: 2,
			},
			expectedError: false,
		},
		{
			name: "Invalid CSV format",
			csv: `id,name
1,"unclosed quote
2,test2`,
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "parse error on line",
		},
		{
			name: "Non-string input",
			csv:  123,
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "expected type of",
		},
		{
			name: "Large CSV with sampling",
			csv:  generateLargeCSV(1000), // Helper function to generate large CSV
			opts: ValidateCSVOptions{
				SampleSize:   100,
				SkipInterval: 100,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateFunc := ValidateCSVFormatWithConfig(tt.opts)
			path := cty.Path{cty.GetAttrStep{Name: "test_csv"}}
			diags := validateFunc(tt.csv, path)

			if tt.expectedError {
				if !diags.HasError() {
					t.Error("expected error but got none")
				} else if tt.errorMessage != "" {
					// Get the error message from diagnostics
					var found bool
					for _, diag := range diags {
						if strings.Contains(diag.Detail, tt.errorMessage) ||
							strings.Contains(diag.Summary, tt.errorMessage) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error message containing '%s', got diagnostics: %v",
							tt.errorMessage, diags)
					}
				}
			} else if diags.HasError() {
				t.Errorf("unexpected errors: %v", diags)
			}
		})
	}
}

// Helper function to generate large CSV for testing
func generateLargeCSV(rows int) string {
	var builder strings.Builder
	builder.WriteString("id,name,value\n")

	for i := 1; i <= rows; i++ {
		builder.WriteString(fmt.Sprintf("%d,test%d,value%d\n", i, i, i))
	}

	return builder.String()
}

// Test specific edge cases
func TestValidateCSVFormatWithConfigEdgeCases(t *testing.T) {
	t.Run("Zero skip interval defaults to 1000", func(t *testing.T) {
		csv := generateLargeCSV(2000)
		opts := ValidateCSVOptions{
			SampleSize:   10,
			SkipInterval: 0, // Should default to 1000
		}

		validateFunc := ValidateCSVFormatWithConfig(opts)
		diags := validateFunc(csv, cty.Path{cty.GetAttrStep{Name: "test_csv"}})

		if diags.HasError() {
			t.Errorf("unexpected validation failure: %v", diags)
		}
	})

	t.Run("CSV with quoted fields", func(t *testing.T) {
		csv := `id,name,description
1,"Smith, John","Description, with comma"
2,"Jones, Bob","Another, description"`
		opts := ValidateCSVOptions{
			SampleSize: 10,
		}

		validateFunc := ValidateCSVFormatWithConfig(opts)
		diags := validateFunc(csv, cty.Path{cty.GetAttrStep{Name: "test_csv"}})

		if diags.HasError() {
			t.Errorf("unexpected validation failure: %v", diags)
		}
	})
}

func BenchmarkValidateCSVFormatWithConfig(b *testing.B) {
	smallCSV := generateLargeCSV(100)
	mediumCSV := generateLargeCSV(10_000)
	largeCSV := generateLargeCSV(1_000_000)
	xlargeCSV := generateLargeCSV(10_000_000)

	opts := ValidateCSVOptions{
		RequiredColumns: []string{"id", "name", "value"},
		SampleSize:      100,
		SkipInterval:    1000,
	}

	validateFunc := ValidateCSVFormatWithConfig(opts)
	path := cty.Path{cty.GetAttrStep{Name: "test_csv"}}

	b.Run("Small CSV (100 rows)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validateFunc(smallCSV, path)
		}
	})

	b.Run("Medium CSV (10_000 rows)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validateFunc(mediumCSV, path)
		}
	})

	b.Run("Large CSV (1_000_000 rows)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validateFunc(largeCSV, path)
		}
	})

	b.Run("X-Large CSV (10_000_000 rows)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validateFunc(xlargeCSV, path)
		}
	})
}
