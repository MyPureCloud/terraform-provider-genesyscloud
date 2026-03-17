package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestPermissionsDataStructure verifies the JSON structure is correct
func TestPermissionsDataStructure(t *testing.T) {
	// Create a sample permissions data
	testData := PermissionsData{
		Version: "test",
		Resources: []ResourcePermissions{
			{
				ResourceType: "genesyscloud_test_resource",
				ResourceName: "test_resource",
				Permissions:  []string{"test:permission:view"},
				Scopes:       []string{"test-scope"},
				Endpoints:    []string{"GET /api/v2/test"},
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Unmarshal back
	var decoded PermissionsData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	// Verify structure
	if decoded.Version != "test" {
		t.Errorf("Expected version 'test', got '%s'", decoded.Version)
	}

	if len(decoded.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(decoded.Resources))
	}

	resource := decoded.Resources[0]
	if resource.ResourceType != "genesyscloud_test_resource" {
		t.Errorf("Expected resource type 'genesyscloud_test_resource', got '%s'", resource.ResourceType)
	}

	if len(resource.Permissions) != 1 || resource.Permissions[0] != "test:permission:view" {
		t.Errorf("Permissions not correctly preserved")
	}

	if len(resource.Scopes) != 1 || resource.Scopes[0] != "test-scope" {
		t.Errorf("Scopes not correctly preserved")
	}

	if len(resource.Endpoints) != 1 || resource.Endpoints[0] != "GET /api/v2/test" {
		t.Errorf("Endpoints not correctly preserved")
	}
}

// TestWritePermissionsJSON verifies the file writing functionality
func TestWritePermissionsJSON(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create test data
	testData := []ResourcePermissions{
		{
			ResourceType: "genesyscloud_test_resource",
			ResourceName: "test_resource",
			Permissions:  []string{"test:permission:view"},
			Scopes:       []string{"test-scope"},
			Endpoints:    []string{"GET /api/v2/test"},
		},
	}

	// Write to file
	filename := "test_permissions"
	version := "1.0.0"
	err := writePermissionsJSON(testData, tempDir, filename, version)
	if err != nil {
		t.Fatalf("Failed to write permissions JSON: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tempDir, "test_permissions-1.0.0.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected file not created: %s", expectedPath)
	}

	// Read and verify content
	fileData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	var decoded PermissionsData
	if err := json.Unmarshal(fileData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal generated file: %v", err)
	}

	// Verify version
	if decoded.Version != version {
		t.Errorf("Expected version '%s', got '%s'", version, decoded.Version)
	}

	// Verify resources
	if len(decoded.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(decoded.Resources))
	}
}

// TestParseAPIEndpoints verifies endpoint parsing from markdown
func TestParseAPIEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "dash format",
			content: `- [GET /api/v2/test](https://example.com)
- [POST /api/v2/test](https://example.com)`,
			expected: 2,
		},
		{
			name: "asterisk format",
			content: `* [GET /api/v2/test](https://example.com)
* [POST /api/v2/test](https://example.com)`,
			expected: 2,
		},
		{
			name: "mixed format",
			content: `- [GET /api/v2/test](https://example.com)
* [POST /api/v2/test](https://example.com)`,
			expected: 2,
		},
		{
			name:     "no endpoints",
			content:  "Some text without endpoints",
			expected: 0,
		},
		{
			name: "with path parameters",
			content: `- [GET /api/v2/test/{id}](https://example.com)
- [DELETE /api/v2/test/{id}/items/{itemId}](https://example.com)`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := parseAPIEndpoints(tt.content)
			if len(endpoints) != tt.expected {
				t.Errorf("Expected %d endpoints, got %d", tt.expected, len(endpoints))
			}
		})
	}
}

// TestInsertPermissionsAndScopes verifies the insertion logic
func TestInsertPermissionsAndScopes(t *testing.T) {
	tests := []struct {
		name                 string
		content              string
		permissionsAndScopes string
		expectBefore         string
	}{
		{
			name: "insert before heading",
			content: `- [GET /api/v2/test](link)

## Migration Notes

Some migration content`,
			permissionsAndScopes: "## Permissions\n\nTest permissions\n",
			expectBefore:         "## Migration Notes",
		},
		{
			name: "append when no heading",
			content: `- [GET /api/v2/test](link)
- [POST /api/v2/test](link)`,
			permissionsAndScopes: "## Permissions\n\nTest permissions\n",
			expectBefore:         "", // Should be at end
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := insertPermissionsAndScopes(tt.content, tt.permissionsAndScopes)

			// Verify permissions section is present
			if !contains(result, "## Permissions") {
				t.Error("Permissions section not found in result")
			}

			// If we expect it before a heading, verify that
			if tt.expectBefore != "" {
				permsIdx := indexOf(result, "## Permissions")
				headingIdx := indexOf(result, tt.expectBefore)
				if permsIdx == -1 || headingIdx == -1 {
					t.Error("Could not find expected sections")
				} else if permsIdx >= headingIdx {
					t.Error("Permissions section should appear before heading")
				}
			}
		})
	}
}

// Helper functions
func contains(s, substr string) bool {
	return indexOf(s, substr) != -1
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
