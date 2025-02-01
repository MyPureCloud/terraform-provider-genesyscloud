package testrunner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestGetRootDir(t *testing.T) {
	// Call getRootDir
	rootDir := getRootDir()

	// Verify the returned path exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		t.Errorf("getRootDir() returned non-existent directory: %s", rootDir)
	}

	// Verify main.go exists in the returned directory
	mainGoPath := filepath.Join(rootDir, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("main.go not found in returned directory: %s", mainGoPath)
	}

	// Optionally verify the path contains expected child directories
	if _, err := os.Stat(filepath.Join(rootDir, "genesyscloud")); os.IsNotExist(err) {
		t.Errorf("Expected root directory to contain 'terraform-provider-genesyscloud' directory, got: %s", rootDir)
	}

	if _, err := os.Stat(filepath.Join(rootDir, "test")); os.IsNotExist(err) {
		t.Errorf("Expected root directory to contain 'test' directory, got: %s", rootDir)
	}
}

func TestGetTestDataPath(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
		want     string
	}{
		{
			name:     "single element path",
			elements: []string{"test1"},
			want:     filepath.Join(getRootDir(), "test", "data", "test1"),
		},
		{
			name:     "multiple element path",
			elements: []string{"test1", "test2", "test3"},
			want:     filepath.Join(getRootDir(), "test", "data", "test1", "test2", "test3"),
		},
		{
			name:     "empty elements",
			elements: []string{},
			want:     filepath.Join(getRootDir(), "test", "data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTestDataPath(tt.elements...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateFullPathId(t *testing.T) {
	tests := []struct {
		name          string
		resourceType  string
		resourceLabel string
		expected      string
	}{
		{
			name:          "standard path",
			resourceType:  "aws_instance",
			resourceLabel: "test",
			expected:      "aws_instance.test.id",
		},
		{
			name:          "empty values",
			resourceType:  "",
			resourceLabel: "",
			expected:      "..id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFullPathId(tt.resourceType, tt.resourceLabel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateTestProvider(t *testing.T) {
	schemas := map[string]*schema.Schema{
		"test_field": {
			Type:     schema.TypeString,
			Required: true,
		},
	}

	provider := GenerateTestProvider("test_resource", schemas, nil)

	assert.NotNil(t, provider)
	assert.NotNil(t, provider.ResourcesMap["test_resource"])
	assert.Equal(t, schemas, provider.ResourcesMap["test_resource"].Schema)
}

func TestGenerateTestDiff(t *testing.T) {
	// Create a mock provider
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"test_resource": &schema.Resource{
				Schema: map[string]*schema.Schema{
					"simple_attr": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
					},
					"list_attr": &schema.Schema{
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name         string
		resourceName string
		oldValue     map[string]string
		newValue     map[string]string
		wantErr      bool
	}{
		{
			name:         "Simple attribute change",
			resourceName: "test_resource",
			oldValue: map[string]string{
				"simple_attr": "old",
			},
			newValue: map[string]string{
				"simple_attr": "new",
			},
			wantErr: false,
		},
		{
			name:         "List attribute change",
			resourceName: "test_resource",
			oldValue: map[string]string{
				"list_attr.#": "1",
				"list_attr.0": "old_item",
			},
			newValue: map[string]string{
				"list_attr.#": "2",
				"list_attr.0": "new_item1",
				"list_attr.1": "new_item2",
			},
			wantErr: false,
		},
		{
			name:         "Invalid resource name",
			resourceName: "invalid_resource",
			oldValue:     map[string]string{},
			newValue:     map[string]string{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := GenerateTestDiff(provider, tt.resourceName, tt.oldValue, tt.newValue)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateTestDiff() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateTestDiff() error = %v", err)
				return
			}

			// Verify the diff was generated
			if diff == nil {
				t.Error("GenerateTestDiff() returned nil diff")
				return
			}

			// For simple attribute change, verify the diff contains the change
			if tt.name == "Simple attribute change" {
				if attr, ok := diff.Attributes["simple_attr"]; !ok {
					t.Error("Expected diff for simple_attr but found none")
				} else {
					if attr.Old != "old" || attr.New != "new" {
						t.Errorf("Unexpected diff values for simple_attr: got old=%v, new=%v", attr.Old, attr.New)
					}
				}
			}

			// For list attribute change, verify the diff contains the changes
			if tt.name == "List attribute change" {
				if attr, ok := diff.Attributes["list_attr.#"]; !ok {
					t.Error("Expected diff for list_attr.# but found none")
				} else {
					if attr.Old != "1" || attr.New != "2" {
						t.Errorf("Unexpected diff values for list_attr.#: got old=%v, new=%v", attr.Old, attr.New)
					}
				}
			}
		})
	}
}
