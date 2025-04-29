package validators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	testrunner "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			csvContent: `field1,'field2',"field3"
value1,"mixed'quotes",value3
value4,value5,value6,value7,,
`,
			opts: ValidateCSVOptions{
				SampleSize: 10,
			},
			expectedError: true,
			errorMessage:  "wrong number of fields",
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

func TestFileContentHashChanged(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp(testrunner.GetTestDataPath(), "test-content-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial content
	initialContent := []byte("initial content")
	if err := os.WriteFile(tmpFile.Name(), initialContent, 0644); err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}

	tests := []struct {
		name         string
		setupFunc    func() error
		expectedDiff bool
	}{
		{
			name: "content_unchanged",
			setupFunc: func() error {
				// No changes to file
				return nil
			},
			expectedDiff: false,
		},
		{
			name: "content_changed",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte("changed content"), 0644)
			},
			expectedDiff: true,
		},
		{
			name: "content_unchanged_again",
			setupFunc: func() error {
				// No changes to file
				return nil
			},
			expectedDiff: false,
		},
		{
			name: "content_changed_again",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte("changed content again"), 0644)
			},
			expectedDiff: true,
		},
		{
			name: "final_content_changed",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte("final changed content"), 0644)
			},
			expectedDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := testrunner.GenerateTestProvider("test_resource",
				map[string]*schema.Schema{
					"filepath": {
						Type:     schema.TypeString,
						Required: true,
					},
					"file_content_hash": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
				customdiff.ComputedIf("file_content_hash", ValidateFileContentHashChanged("filepath", "file_content_hash")),
			)

			// Pre calculate hash
			priorHash, err := files.HashFileContent(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to calculate hash: %v", err)
			}

			// Run setup for this test case
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			diff, err := testrunner.GenerateTestDiff(
				provider,
				"test_resource",
				map[string]string{
					"filepath":          tmpFile.Name(),
					"file_content_hash": priorHash,
				},
				map[string]string{
					"filepath": tmpFile.Name(),
				},
			)

			if err != nil {
				t.Fatalf("Diff failed with error: %s", err)
			}

			if tt.expectedDiff {
				if diff == nil {
					t.Error("Expected a diff when file content changes, got nil")
				} else if !diff.Attributes["file_content_hash"].NewComputed {
					t.Error("file_content_hash is not marked as NewComputed when file content changes")
				}
			} else {
				if diff != nil && diff.Attributes["file_content_hash"].NewComputed {
					t.Error("Expected no diff when file content unchanged, but file_content_hash was marked as NewComputed")
				}
			}
		})
	}
}

func TestValidateCSVWithColumns(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp(testrunner.GetTestDataPath(), "test-csv-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name          string
		setupFunc     func() error
		oldValues     map[string]string
		newValues     map[string]string
		expectedDiff  bool
		expectedError bool
		errorMessage  string
	}{
		{
			name: "valid_csv_with_columns",
			setupFunc: func() error {
				content := "header1,header2,header3\nvalue1,value2,value3"
				return os.WriteFile(tmpFile.Name(), []byte(content), 0644)
			},
			oldValues: map[string]string{
				"filepath":       tmpFile.Name() + ".old",
				"column_names.#": "2",
				"column_names.0": "old_header1",
				"column_names.1": "old_header2",
			},
			newValues: map[string]string{
				"filepath":       tmpFile.Name(),
				"column_names.#": "3",
				"column_names.0": "header1",
				"column_names.1": "header2",
				"column_names.2": "header3",
			},
			expectedDiff:  true,
			expectedError: false,
		},
		{
			name: "missing_required_column",
			setupFunc: func() error {
				content := "header1,header3\nvalue1,value3"
				return os.WriteFile(tmpFile.Name(), []byte(content), 0644)
			},
			oldValues: map[string]string{
				"filepath":       tmpFile.Name() + ".old",
				"column_names.#": "2",
				"column_names.0": "old_header1",
				"column_names.1": "old_header2",
			},
			newValues: map[string]string{
				"filepath":       tmpFile.Name(),
				"column_names.#": "3",
				"column_names.0": "header1",
				"column_names.1": "header2",
				"column_names.2": "header3",
			},
			expectedDiff:  false,
			expectedError: true,
			errorMessage:  "missing required columns: [header2]",
		},
		{
			name: "empty_file",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte(""), 0644)
			},
			oldValues: map[string]string{
				"filepath":       tmpFile.Name() + ".old",
				"column_names.#": "1",
				"column_names.0": "old_header",
			},
			newValues: map[string]string{
				"filepath":       tmpFile.Name(),
				"column_names.#": "2",
				"column_names.0": "header1",
				"column_names.1": "header2",
			},
			expectedDiff:  false,
			expectedError: true,
			errorMessage:  "failed to read CSV headers",
		},
		{
			name: "file_with_only_headers",
			setupFunc: func() error {
				content := "header1,header2\n"
				return os.WriteFile(tmpFile.Name(), []byte(content), 0644)
			},
			oldValues: map[string]string{
				"filepath":       tmpFile.Name() + ".old",
				"column_names.#": "1",
				"column_names.0": "old_header",
			},
			newValues: map[string]string{
				"filepath":       tmpFile.Name(),
				"column_names.#": "2",
				"column_names.0": "header1",
				"column_names.1": "header2",
			},
			expectedDiff:  true,
			expectedError: false,
		},
		{
			name: "case_sensitive_headers",
			setupFunc: func() error {
				content := "Header1,HEADER2\nvalue1,value2"
				return os.WriteFile(tmpFile.Name(), []byte(content), 0644)
			},
			oldValues: map[string]string{
				"filepath":       tmpFile.Name() + ".old",
				"column_names.#": "1",
				"column_names.0": "old_header",
			},
			newValues: map[string]string{
				"filepath":       tmpFile.Name(),
				"column_names.#": "2",
				"column_names.0": "header1",
				"column_names.1": "header2",
			},
			expectedDiff:  false,
			expectedError: true,
			errorMessage:  "missing required column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &schema.Resource{
				Schema: map[string]*schema.Schema{
					"filepath": {
						Type:     schema.TypeString,
						Required: true,
					},
					"column_names": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			}

			provider := testrunner.GenerateTestProvider("test_resource", resource.Schema, ValidateCSVWithColumns("filepath", "column_names"))

			// Run setup for this test case
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			diff, err := testrunner.GenerateTestDiff(
				provider,
				"test_resource",
				tt.oldValues,
				tt.newValues,
			)

			// Check for expected error
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error for '%s' check but got none", tt.name)
				} else if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message containing '%s', got: %v for '%s' check", tt.errorMessage, err, tt.name)
				}
			} else if err != nil {
				t.Errorf("Unexpected error for '%s' check: %v", tt.name, err)
			}

			// Check for expected diff
			if tt.expectedDiff {
				if diff == nil {
					t.Errorf("Expected a diff for '%s' check but got nil", tt.name)
				}
			} else {
				if diff != nil {
					t.Errorf("Expected no diff for '%s' check but got one", tt.name)
				}
			}
		})
	}
}
