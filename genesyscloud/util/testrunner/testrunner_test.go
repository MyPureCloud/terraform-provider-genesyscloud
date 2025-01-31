package testrunner

import (
	"os"
	"path/filepath"
	"testing"

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
