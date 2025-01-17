package outbound_contact_list_contacts_bulk

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFileContentHashChanged(t *testing.T) {
	r := schema.Resource{
		Schema: map[string]*schema.Schema{
			"filepath": {
				Type:     schema.TypeString,
				Required: true,
			},
			"file_content_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

	tests := []struct {
		name          string
		filepath      string
		oldHash       string
		expectedDiff  bool
		setupTestFile func(string) error
	}{
		{
			name:         "hash changed",
			filepath:     "testdata/test.txt",
			oldHash:      "old-hash-value",
			expectedDiff: true,
			setupTestFile: func(path string) error {
				return os.WriteFile(path, []byte("new content"), 0644)
			},
		},
		{
			name:         "hash unchanged",
			filepath:     "testdata/test.txt",
			oldHash:      "", // This will be set to the actual hash after file creation
			expectedDiff: false,
			setupTestFile: func(path string) error {
				return os.WriteFile(path, []byte("test content"), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory if it doesn't exist
			err := os.MkdirAll("testdata", 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}
			defer os.RemoveAll("testdata")

			// Setup test file
			err = tt.setupTestFile(tt.filepath)
			if err != nil {
				t.Fatalf("Failed to setup test file: %v", err)
			}

			// Create ResourceDiff using the SDK's testing framework
			oldConfig := map[string]interface{}{
				"filepath":          tt.filepath,
				"file_content_hash": tt.oldHash,
			}

			d := schema.TestResourceDataRaw(t, r.Schema, oldConfig)

			// If testing for unchanged hash, calculate the actual hash
			if tt.oldHash == "" {
				hash, err := fileContentHashReader(tt.filepath)
				if err != nil {
					t.Fatalf("Failed to calculate hash: %v", err)
				}
				d.Set("file_content_hash", hash)
			}

			// Test the function
			ctx := context.Background()
			result := fileContentHashChanged(ctx, d, nil)

			if result != tt.expectedDiff {
				t.Errorf("fileContentHashChanged() = %v, want %v", result, tt.expectedDiff)
			}
		})
	}
}
