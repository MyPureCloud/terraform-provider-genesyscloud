package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPermissionsDataStructure(t *testing.T) {
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

	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	var decoded PermissionsData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

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

func TestWritePermissionsJSON(t *testing.T) {
	tempDir := t.TempDir()

	testData := []ResourcePermissions{
		{
			ResourceType: "genesyscloud_test_resource",
			ResourceName: "test_resource",
			Permissions:  []string{"test:permission:view"},
			Scopes:       []string{"test-scope"},
			Endpoints:    []string{"GET /api/v2/test"},
		},
	}

	err := writePermissionsJSON(testData, tempDir, "test_permissions", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to write permissions JSON: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "test_permissions-1.0.0.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected file not created: %s", expectedPath)
	}

	fileData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	var decoded PermissionsData
	if err := json.Unmarshal(fileData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal generated file: %v", err)
	}
	if decoded.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", decoded.Version)
	}
	if len(decoded.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(decoded.Resources))
	}
}

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

func TestBuildOperationIdMap(t *testing.T) {
	spec := &SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/scripts": {
				"get":  {OperationId: "getScripts"},
				"post": {OperationId: "postScripts"},
			},
			"/api/v2/scripts/{scriptId}": {
				"get":    {OperationId: "getScript"},
				"delete": {OperationId: "deleteScript"},
			},
		},
	}

	opMap := buildOperationIdMap(spec)

	if len(opMap) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(opMap))
	}

	ep, ok := opMap["getscripts"]
	if !ok {
		t.Fatal("Expected 'getscripts' in map")
	}
	if ep.Method != "GET" || ep.Path != "/api/v2/scripts" {
		t.Errorf("Expected GET /api/v2/scripts, got %s %s", ep.Method, ep.Path)
	}

	ep, ok = opMap["deletescript"]
	if !ok {
		t.Fatal("Expected 'deletescript' in map")
	}
	if ep.Method != "DELETE" || ep.Path != "/api/v2/scripts/{scriptId}" {
		t.Errorf("Expected DELETE /api/v2/scripts/{scriptId}, got %s %s", ep.Method, ep.Path)
	}
}

func TestDetectEndpointsFromProxy(t *testing.T) {
	opMap := map[string]APIEndpoint{
		"getscripts":             {Method: "GET", Path: "/api/v2/scripts"},
		"postscriptspublished":   {Method: "POST", Path: "/api/v2/scripts/published"},
		"getscript":              {Method: "GET", Path: "/api/v2/scripts/{scriptId}"},
		"postscriptexport":       {Method: "POST", Path: "/api/v2/scripts/{scriptId}/export"},
		"getscriptsuploadstatus": {Method: "GET", Path: "/api/v2/scripts/uploads/{uploadId}/status"},
	}

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "SDK method calls with Api suffix",
			content:  `p.scriptsApi.GetScripts(pageSize, pageNum, "", scriptName, "", "", "", "", "", "")`,
			expected: []string{"GET /api/v2/scripts"},
		},
		{
			name:     "SDK method calls with lowercase api",
			content:  `a.api.GetScript(id)`,
			expected: []string{"GET /api/v2/scripts/{scriptId}"},
		},
		{
			name: "Raw HTTP call with BasePath",
			content: `fullPath := p.scriptsApi.Configuration.BasePath + "/api/v2/scripts/" + scriptId
r, _ := http.NewRequestWithContext(ctx, http.MethodDelete, fullPath, nil)`,
			expected: []string{"DELETE /api/v2/scripts/{scriptsId}"},
		},
		{
			name: "Multiple SDK calls",
			content: `p.scriptsApi.GetScripts(100, 1, "", "", "", "", "", "", "", "")
p.scriptsApi.PostScriptsPublished("0", *publishScriptBody)
p.scriptsApi.PostScriptExport(scriptId, body)`,
			expected: []string{
				"GET /api/v2/scripts",
				"POST /api/v2/scripts/published",
				"POST /api/v2/scripts/{scriptId}/export",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := detectEndpointsFromProxy(tt.content, opMap)
			detected := make(map[string]bool)
			for _, ep := range endpoints {
				detected[ep.Method+" "+ep.Path] = true
			}
			for _, exp := range tt.expected {
				if !detected[exp] {
					t.Errorf("Expected endpoint %s not detected. Got: %v", exp, endpoints)
				}
			}
		})
	}
}

func TestDetectHTTPMethodNearPath(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		path     string
		expected string
	}{
		{
			name:     "http.MethodDelete",
			content:  `r, _ := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v2/scripts/" + id, nil)`,
			path:     `"/api/v2/scripts/"`,
			expected: "DELETE",
		},
		{
			name:     "http.MethodPost",
			content:  "action := http.MethodPost\nfullPath := p.Configuration.BasePath + \"/api/v2/authorization/divisions/\"",
			path:     `"/api/v2/authorization/divisions/"`,
			expected: "POST",
		},
		{
			name:     "string literal GET",
			content:  "method := \"GET\"\nurl := \"/api/v2/test\"",
			path:     `"/api/v2/test"`,
			expected: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectHTTPMethodNearPath(tt.content, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNormalizeRawPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v2/scripts", "/api/v2/scripts"},
		{"/api/v2/scripts/", "/api/v2/scripts/{scriptsId}"},
		{"/api/v2/authorization/divisions/", "/api/v2/authorization/divisions/{divisionsId}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeRawPath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}


func TestReadNotesFile(t *testing.T) {
	tempDir := t.TempDir()

	// No notes.md - should return empty
	result := readNotesFile(tempDir)
	if result != "" {
		t.Errorf("Expected empty string for missing notes.md, got %q", result)
	}

	// With notes.md
	notesContent := "## Export Behavior\n\nSome notes here.\n"
	os.WriteFile(filepath.Join(tempDir, "notes.md"), []byte(notesContent), 0644)

	result = readNotesFile(tempDir)
	if result != "## Export Behavior\n\nSome notes here." {
		t.Errorf("Expected trimmed notes content, got %q", result)
	}
}

func TestUpdateApisMdFromProxy(t *testing.T) {
	tempDir := t.TempDir()
	examplesDir := filepath.Join(tempDir, "examples")
	os.MkdirAll(examplesDir, 0755)

	// Create an existing apis.md with unsorted endpoints
	existingContent := `* [POST /api/v2/scripts](https://developer.genesys.cloud/devapps/api-explorer#post--api-v2-scripts)
* [GET /api/v2/scripts/{scriptId}](https://developer.genesys.cloud/devapps/api-explorer#get--api-v2-scripts--scriptId-)
`
	apisMdFile := filepath.Join(examplesDir, "apis.md")
	os.WriteFile(apisMdFile, []byte(existingContent), 0644)

	// Create a fake proxy file that references additional endpoints
	proxyFile := filepath.Join(tempDir, "proxy.go")
	proxyContent := `package script
func (p *proxy) doStuff() {
	p.scriptsApi.DeleteScript(id)
	p.scriptsApi.GetScripts(100, 1, "", "", "", "", "", "", "", "")
}
`
	os.WriteFile(proxyFile, []byte(proxyContent), 0644)

	// Create a .source file pointing to the proxy (relative to examplesDir)
	relPath, _ := filepath.Rel(examplesDir, proxyFile)
	os.WriteFile(filepath.Join(examplesDir, ".source"), []byte(relPath), 0644)

	opMap := map[string]APIEndpoint{
		"deletescript": {Method: "DELETE", Path: "/api/v2/scripts/{scriptId}"},
		"getscripts":   {Method: "GET", Path: "/api/v2/scripts"},
	}

	updateApisMdFromProxy("genesyscloud_script", examplesDir, apisMdFile, opMap)

	result, err := os.ReadFile(apisMdFile)
	if err != nil {
		t.Fatalf("Failed to read apis.md: %v", err)
	}

	endpoints := parseAPIEndpoints(string(result))
	if len(endpoints) != 4 {
		t.Fatalf("Expected 4 endpoints, got %d: %s", len(endpoints), string(result))
	}

	// Verify sorted order: path first, then method
	expected := []struct{ method, path string }{
		{"get", "/api/v2/scripts"},
		{"post", "/api/v2/scripts"},
		{"delete", "/api/v2/scripts/{scriptId}"},
		{"get", "/api/v2/scripts/{scriptId}"},
	}
	for i, exp := range expected {
		if endpoints[i].Method != exp.method || endpoints[i].Path != exp.path {
			t.Errorf("Endpoint[%d]: expected %s %s, got %s %s", i, exp.method, exp.path, endpoints[i].Method, endpoints[i].Path)
		}
	}
}

func TestBuildAnchor(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/api/v2/scripts", "get--api-v2-scripts"},
		{"DELETE", "/api/v2/scripts/{scriptId}", "delete--api-v2-scripts--scriptId-"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			result := buildAnchor(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
