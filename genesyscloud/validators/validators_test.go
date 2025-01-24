package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCSVFormatWithConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "csv-tests")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up after tests

	tests := []struct {
		name          string
		csvContent    string
		opts          ValidateCSVOptions
		expectedError bool
		errorMessage  string
	}{
		{
			name: "Valid CSV with required columns",
			csvContent: `id,name,value
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
			csvContent: `id,value
1,val1
2,val2`,
			opts: ValidateCSVOptions{
				RequiredColumns: []string{"id", "name"},
				SampleSize:      10,
			},
			expectedError: true,
			errorMessage:  "CSV file is missing required columns: [name]",
		},
		{
			name: "Inconsistent number of fields",
			csvContent: `id,name,value
1,test1
2,test2,val2`,
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "error reading line 1: record on line 2: wrong number of fields",
		},
		{
			name:       "Empty CSV",
			csvContent: "",
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "failed to read CSV headers",
		},
		{
			name: "CSV exceeds max row count",
			csvContent: `id,name
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
			csvContent: `id,name
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
			csvContent: `id,name
1,"unclosed quote
2,test2`,
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "parse error on line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for this test
			tmpFile := filepath.Join(tmpDir, fmt.Sprintf("test-%s.csv", tt.name))
			err := os.WriteFile(tmpFile, []byte(tt.csvContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err = ValidateCSVFormatWithConfig(tmpFile, tt.opts)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("expected error message containing '%s', got: %v",
						tt.errorMessage, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
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
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "csv-edge-cases")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("Zero skip interval defaults to 1000", func(t *testing.T) {
		tmpFile := filepath.Join(tmpDir, "large.csv")
		err := os.WriteFile(tmpFile, []byte(generateLargeCSV(2000)), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		opts := ValidateCSVOptions{
			SampleSize:   10,
			SkipInterval: 0, // Should default to 1000
		}

		err = ValidateCSVFormatWithConfig(tmpFile, opts)
		if err != nil {
			t.Errorf("unexpected validation failure: %v", err)
		}
	})

	t.Run("CSV with quoted fields", func(t *testing.T) {
		csvContent := `id,name,description
1,"Smith, John","Description, with comma"
2,"Jones, Bob","Another, description"`

		tmpFile := filepath.Join(tmpDir, "quoted.csv")
		err := os.WriteFile(tmpFile, []byte(csvContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		opts := ValidateCSVOptions{
			SampleSize: 10,
		}

		err = ValidateCSVFormatWithConfig(tmpFile, opts)
		if err != nil {
			t.Errorf("unexpected validation failure: %v", err)
		}
	})
}

func BenchmarkValidateCSVFormatWithConfig(b *testing.B) {
	// Create a temporary directory for benchmark files
	tmpDir, err := os.MkdirTemp("", "csv-benchmarks")
	if err != nil {
		b.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files of different sizes
	sizes := map[string]int{
		"Small":  100,
		"Medium": 10_000,
		"Large":  100_000,
	}

	files := make(map[string]string)
	for name, size := range sizes {
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("%s.csv", name))
		err := os.WriteFile(tmpFile, []byte(generateLargeCSV(size)), 0644)
		if err != nil {
			b.Fatalf("failed to create benchmark file: %v", err)
		}
		files[name] = tmpFile
	}

	opts := ValidateCSVOptions{
		RequiredColumns: []string{"id", "name", "value"},
		SampleSize:      100,
		SkipInterval:    1000,
	}

	for name, file := range files {
		b.Run(fmt.Sprintf("%s CSV", name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ValidateCSVFormatWithConfig(file, opts)
			}
		})
	}
}
