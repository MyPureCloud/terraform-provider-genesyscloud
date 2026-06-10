package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
		DataSources: []ResourcePermissions{
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
	if len(decoded.DataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(decoded.DataSources))
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

	resources := []ResourcePermissions{
		{
			ResourceType: "genesyscloud_test_resource",
			ResourceName: "test_resource",
			Permissions:  []string{"test:permission:view"},
			Scopes:       []string{"test-scope"},
			Endpoints:    []string{"GET /api/v2/test"},
		},
	}
	dataSources := []ResourcePermissions{
		{
			ResourceType: "genesyscloud_test_resource",
			ResourceName: "test_resource",
			Permissions:  []string{"test:permission:view"},
			Scopes:       []string{"test-scope"},
			Endpoints:    []string{"GET /api/v2/test"},
		},
	}

	err := writePermissionsJSON(resources, dataSources, tempDir, "test_permissions", "1.0.0")
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
	if len(decoded.DataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(decoded.DataSources))
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
			name: "custom_api_client Do with typed response",
			content: `result, resp, err := customapi.Do[platformclientv2.Guide](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/scripts/", nil, nil)`,
			expected: []string{"GET /api/v2/scripts/{scriptId}"},
		},
		{
			name: "custom_api_client DoNoResponse",
			content: `resp, err := customapi.DoNoResponse(ctx, p.customApiClient, customapi.MethodDelete, "/api/v2/scripts/" + id, nil, nil)`,
			expected: []string{"DELETE /api/v2/scripts/{scriptId}"},
		},
		{
			name: "custom_api_client DoRaw",
			content: `raw, resp, err := customapi.DoRaw(ctx, p.customApiClient, customapi.MethodPost, "/api/v2/scripts/published", body, nil)`,
			expected: []string{"POST /api/v2/scripts/published"},
		},
		{
			name: "custom_api_client DoWithAcceptHeader",
			content: `raw, resp, err := customapi.DoWithAcceptHeader(ctx, p.customApiClient, customapi.MethodGet, "/api/v2/scripts", nil, nil, "text/csv")`,
			expected: []string{"GET /api/v2/scripts"},
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
		{
			name: "Mixed SDK and custom_api_client",
			content: `p.scriptsApi.GetScripts(100, 1, "", "", "", "", "", "", "", "")
resp, err := customapi.DoNoResponse(ctx, p.customApiClient, customapi.MethodDelete, "/api/v2/scripts/" + scriptId, nil, nil)`,
			expected: []string{
				"GET /api/v2/scripts",
				"DELETE /api/v2/scripts/{scriptId}",
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


func TestNormalizeRawPath(t *testing.T) {
	swaggerPaths := map[string]bool{
		"/api/v2/scripts":                              true,
		"/api/v2/scripts/{scriptId}":                   true,
		"/api/v2/authorization/divisions/{divisionId}": true,
		"/api/v2/routing/queues/{queueId}":             true,
		"/api/v2/knowledge/categories/{categoryId}":    true,
		"/api/v2/responses/{responseId}":               true,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v2/scripts", "/api/v2/scripts"},
		{"/api/v2/scripts/", "/api/v2/scripts/{scriptId}"},
		{"/api/v2/authorization/divisions/", "/api/v2/authorization/divisions/{divisionId}"},
		{"/api/v2/routing/queues/", "/api/v2/routing/queues/{queueId}"},
		{"/api/v2/knowledge/categories/", "/api/v2/knowledge/categories/{categoryId}"},
		{"/api/v2/responses/", "/api/v2/responses/{responseId}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeRawPath(tt.input, swaggerPaths)
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

func TestParseSourcesComment(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single source",
			content:  "<!-- sources\ngenesyscloud/scripts/genesyscloud_scripts_proxy.go\n-->\n* [GET /api/v2/scripts](https://example.com)\n",
			expected: []string{"genesyscloud/scripts/genesyscloud_scripts_proxy.go"},
		},
		{
			name:     "multiple sources",
			content:  "<!-- sources\ngenesyscloud/foo/foo_proxy.go\ngenesyscloud/bar/bar_proxy.go\n-->\n",
			expected: []string{"genesyscloud/foo/foo_proxy.go", "genesyscloud/bar/bar_proxy.go"},
		},
		{
			name:     "no comment",
			content:  "* [GET /api/v2/scripts](https://example.com)\n",
			expected: nil,
		},
		{
			name:     "empty comment",
			content:  "<!-- sources\n-->\n",
			expected: nil,
		},
		{
			name:     "with extra whitespace",
			content:  "<!--  sources \n  genesyscloud/foo/proxy.go  \n-->\n",
			expected: []string{"genesyscloud/foo/proxy.go"},
		},
		{
			name:     "NO_SOURCES sentinel",
			content:  "<!-- sources\nNO_SOURCES\n-->\n* Some note about this resource.\n",
			expected: []string{"NO_SOURCES"},
		},
		{
			name:     "with inline comments",
			content:  "<!-- sources\ngenesyscloud/foo/foo_proxy.go\ngenesyscloud/foo/foo_utils.go # TODO: extract API calls to proxy\n-->\n",
			expected: []string{"genesyscloud/foo/foo_proxy.go", "genesyscloud/foo/foo_utils.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSourcesComment(tt.content)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d sources, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("Source[%d]: expected %q, got %q", i, exp, result[i])
				}
			}
		})
	}
}

func TestUpdateApisMdFromProxy(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	examplesDir := filepath.Join(tempDir, "examples")
	os.MkdirAll(examplesDir, 0755)

	// Create a proxy file
	proxyDir := filepath.Join("genesyscloud", "scripts")
	os.MkdirAll(proxyDir, 0755)
	proxyFile := filepath.Join(proxyDir, "genesyscloud_scripts_proxy.go")
	proxyContent := `package scripts
func (p *proxy) doStuff() {
	p.scriptsApi.DeleteScript(id)
	p.scriptsApi.GetScripts(100, 1, "", "", "", "", "", "", "", "")
}
`
	os.WriteFile(proxyFile, []byte(proxyContent), 0644)

	// Create apis.md with sources comment and existing endpoints
	apisMdFile := filepath.Join(examplesDir, "apis.md")
	existingContent := "<!-- sources\n" + proxyFile + "\n-->\n" +
		"* [POST /api/v2/scripts](https://developer.genesys.cloud/devapps/api-explorer#post--api-v2-scripts)\n" +
		"* [GET /api/v2/scripts/{scriptId}](https://developer.genesys.cloud/devapps/api-explorer#get--api-v2-scripts--scriptId-)\n"
	os.WriteFile(apisMdFile, []byte(existingContent), 0644)

	opMap := map[string]APIEndpoint{
		"deletescript": {Method: "DELETE", Path: "/api/v2/scripts/{scriptId}"},
		"getscripts":   {Method: "GET", Path: "/api/v2/scripts"},
	}

	var errors []string
	updateApisMdFromProxy("genesyscloud_script", apisMdFile, opMap, false, &errors)

	result, err := os.ReadFile(apisMdFile)
	if err != nil {
		t.Fatalf("Failed to read apis.md: %v", err)
	}

	// Verify sources comment is preserved
	sources := parseSourcesComment(string(result))
	if len(sources) != 1 || sources[0] != proxyFile {
		t.Errorf("Sources comment not preserved correctly: %v", sources)
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

func TestUpdateApisMdFromProxy_NoSourcesSkips(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	examplesDir := filepath.Join(tempDir, "examples")
	os.MkdirAll(examplesDir, 0755)

	// Create apis.md with NO_SOURCES
	apisMdFile := filepath.Join(examplesDir, "apis.md")
	content := "<!-- sources\nNO_SOURCES\n-->\n* This resource has no standard APIs.\n"
	os.WriteFile(apisMdFile, []byte(content), 0644)

	opMap := map[string]APIEndpoint{
		"getscripts": {Method: "GET", Path: "/api/v2/scripts"},
	}

	var errors []string
	updateApisMdFromProxy("genesyscloud_tf_export", apisMdFile, opMap, false, &errors)

	// File should be unchanged
	result, _ := os.ReadFile(apisMdFile)
	if string(result) != content {
		t.Errorf("NO_SOURCES file should not be modified, got: %s", string(result))
	}
	if len(errors) != 0 {
		t.Errorf("Expected no errors for NO_SOURCES, got: %v", errors)
	}
}

func TestUpdateApisMdFromProxy_MissingComment(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	examplesDir := filepath.Join(tempDir, "examples")
	os.MkdirAll(examplesDir, 0755)

	// Create apis.md without sources comment
	apisMdFile := filepath.Join(examplesDir, "apis.md")
	os.WriteFile(apisMdFile, []byte("* [GET /api/v2/test](https://example.com)\n"), 0644)

	opMap := map[string]APIEndpoint{}
	var errors []string
	updateApisMdFromProxy("genesyscloud_test_resource", apisMdFile, opMap, false, &errors)

	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0], "missing a <!-- sources --> comment") {
		t.Errorf("Error should mention missing comment, got: %s", errors[0])
	}
}

func TestUpdateApisMdFromProxy_InvalidSourceFile(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	examplesDir := filepath.Join(tempDir, "examples")
	os.MkdirAll(examplesDir, 0755)

	// Create apis.md pointing to non-existent source
	apisMdFile := filepath.Join(examplesDir, "apis.md")
	content := "<!-- sources\ngenesyscloud/missing/missing_proxy.go\n-->\n"
	os.WriteFile(apisMdFile, []byte(content), 0644)

	opMap := map[string]APIEndpoint{}
	var errors []string
	updateApisMdFromProxy("genesyscloud_test_resource", apisMdFile, opMap, false, &errors)

	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0], "does not exist") {
		t.Errorf("Error should mention file not existing, got: %s", errors[0])
	}
}

func TestProcessDocsFolder_PlaceholderReplacement(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	resourceName := "test_resource"
	docContent := "# Resource\n\n**No APIs**\n\nSome footer content\n"
	os.WriteFile(filepath.Join(docsFolder, resourceName+".md"), []byte(docContent), 0644)

	resourceExamplesDir := filepath.Join(examplesFolder, "genesyscloud_"+resourceName)
	os.MkdirAll(resourceExamplesDir, 0755)
	os.WriteFile(filepath.Join(resourceExamplesDir, "apis.md"), []byte("* [GET /api/v2/test](https://example.com)\n"), 0644)

	missingExamples := []string{}
	processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	result, _ := os.ReadFile(filepath.Join(docsFolder, resourceName+".md"))
	resultStr := string(result)
	if strings.Contains(resultStr, "**No APIs**") {
		t.Error("Placeholder **No APIs** was not replaced")
	}
	if !strings.Contains(resultStr, "GET /api/v2/test") {
		t.Error("Expected API content not found in doc file")
	}
	if !strings.Contains(resultStr, "Some footer content") {
		t.Error("Footer content should be preserved")
	}
}

func TestProcessDocsFolder_NotesAppended(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	resourceName := "noted_resource"
	os.WriteFile(filepath.Join(docsFolder, resourceName+".md"), []byte("# Resource\n\n**No APIs**\n"), 0644)

	resourceExamplesDir := filepath.Join(examplesFolder, "genesyscloud_"+resourceName)
	os.MkdirAll(resourceExamplesDir, 0755)
	os.WriteFile(filepath.Join(resourceExamplesDir, "apis.md"), []byte("* [GET /api/v2/noted](https://example.com)\n"), 0644)
	os.WriteFile(filepath.Join(resourceExamplesDir, "notes.md"), []byte("## Export Behavior\n\nThis resource has special export behavior."), 0644)

	missingExamples := []string{}
	processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	result, _ := os.ReadFile(filepath.Join(docsFolder, resourceName+".md"))
	resultStr := string(result)
	if !strings.Contains(resultStr, "## Export Behavior") {
		t.Error("Notes content was not appended to doc file")
	}
	if !strings.Contains(resultStr, "special export behavior") {
		t.Error("Full notes content not present")
	}
	apiIdx := strings.Index(resultStr, "GET /api/v2/noted")
	notesIdx := strings.Index(resultStr, "## Export Behavior")
	if notesIdx < apiIdx {
		t.Error("Notes should appear after API content")
	}
}

func TestProcessDocsFolder_PermissionsInjected(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	resourceName := "perms_resource"
	os.WriteFile(filepath.Join(docsFolder, resourceName+".md"), []byte("# Resource\n\n**No APIs**\n"), 0644)

	resourceExamplesDir := filepath.Join(examplesFolder, "genesyscloud_"+resourceName)
	os.MkdirAll(resourceExamplesDir, 0755)
	os.WriteFile(filepath.Join(resourceExamplesDir, "apis.md"), []byte("* [GET /api/v2/perms](https://example.com)\n* [POST /api/v2/perms](https://example.com)\n"), 0644)

	swaggerSpec := &SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/perms": {
				"get": {
					XIninRequiresPermissions: PermissionsSpec{
						Permissions: []string{"perms:resource:view"},
					},
					Security: []map[string][]string{
						{"PureCloud OAuth": []string{"perms:readonly"}},
					},
				},
				"post": {
					XIninRequiresPermissions: PermissionsSpec{
						Permissions: []string{"perms:resource:add"},
					},
					Security: []map[string][]string{
						{"PureCloud OAuth": []string{"perms"}},
					},
				},
			},
		},
	}

	missingExamples := []string{}
	allPerms := processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		swaggerSpec:     swaggerSpec,
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	result, _ := os.ReadFile(filepath.Join(docsFolder, resourceName+".md"))
	resultStr := string(result)
	if !strings.Contains(resultStr, "## Permissions and Scopes") {
		t.Error("Permissions section not injected")
	}
	if !strings.Contains(resultStr, "`perms:resource:view`") {
		t.Error("Permission 'perms:resource:view' not found")
	}
	if !strings.Contains(resultStr, "`perms:resource:add`") {
		t.Error("Permission 'perms:resource:add' not found")
	}
	if !strings.Contains(resultStr, "`perms:readonly`") {
		t.Error("Scope 'perms:readonly' not found")
	}
	if !strings.Contains(resultStr, "`perms`") {
		t.Error("Scope 'perms' not found")
	}

	if len(allPerms) != 1 {
		t.Fatalf("Expected 1 resource permission entry, got %d", len(allPerms))
	}
	if allPerms[0].ResourceType != "genesyscloud_perms_resource" {
		t.Errorf("Expected resource type 'genesyscloud_perms_resource', got '%s'", allPerms[0].ResourceType)
	}
	if len(allPerms[0].Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(allPerms[0].Permissions))
	}
	if len(allPerms[0].Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(allPerms[0].Scopes))
	}
}

func TestProcessDocsFolder_IgnoredResourcesRemoved(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	ignoredName := "ignored_res"
	normalName := "normal_res"
	os.WriteFile(filepath.Join(docsFolder, ignoredName+".md"), []byte("should be deleted"), 0644)
	os.WriteFile(filepath.Join(docsFolder, normalName+".md"), []byte("# Resource\n\n**No APIs**\n"), 0644)

	normalExamplesDir := filepath.Join(examplesFolder, "genesyscloud_"+normalName)
	os.MkdirAll(normalExamplesDir, 0755)
	os.WriteFile(filepath.Join(normalExamplesDir, "apis.md"), []byte("* [GET /api/v2/normal](https://example.com)\n"), 0644)

	missingExamples := []string{}
	processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		ignoredExamples: []string{"genesyscloud_" + ignoredName},
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	if _, err := os.Stat(filepath.Join(docsFolder, ignoredName+".md")); !os.IsNotExist(err) {
		t.Error("Ignored resource doc file should have been deleted")
	}

	result, err := os.ReadFile(filepath.Join(docsFolder, normalName+".md"))
	if err != nil {
		t.Fatal("Normal resource doc file should still exist")
	}
	if strings.Contains(string(result), "**No APIs**") {
		t.Error("Normal resource placeholder should have been replaced")
	}
}

func TestProcessDocsFolder_MissingExamplesTracked(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	resourceName := "no_examples"
	os.WriteFile(filepath.Join(docsFolder, resourceName+".md"), []byte("# Resource\n\n**No APIs**\n"), 0644)

	missingExamples := []string{}
	processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	if len(missingExamples) != 1 || missingExamples[0] != resourceName {
		t.Errorf("Expected missingExamples to contain '%s', got %v", resourceName, missingExamples)
	}
}

func TestProcessDocsFolder_NonexistentFolder(t *testing.T) {
	missingExamples := []string{}
	result := processDocsFolder(docsFolderProcessor{
		docsFolder:      "/nonexistent/path",
		examplesFolder:  "/also/nonexistent",
		apiDocsTag:      "**No APIs**",
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	if result != nil {
		t.Errorf("Expected nil for nonexistent folder, got %v", result)
	}
}

func TestProcessDocsFolder_PermissionsAndNotesOrdering(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	resourceName := "ordering_resource"
	os.WriteFile(filepath.Join(docsFolder, resourceName+".md"), []byte("# Resource\n\n**No APIs**\n\nEnd of doc\n"), 0644)

	resourceExamplesDir := filepath.Join(examplesFolder, "genesyscloud_"+resourceName)
	os.MkdirAll(resourceExamplesDir, 0755)
	os.WriteFile(filepath.Join(resourceExamplesDir, "apis.md"), []byte("* [GET /api/v2/ordering](https://example.com)\n"), 0644)
	os.WriteFile(filepath.Join(resourceExamplesDir, "notes.md"), []byte("## Notes\n\nImportant note."), 0644)

	swaggerSpec := &SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/ordering": {
				"get": {
					XIninRequiresPermissions: PermissionsSpec{
						Permissions: []string{"ordering:view"},
					},
					Security: []map[string][]string{
						{"PureCloud OAuth": []string{"ordering"}},
					},
				},
			},
		},
	}

	missingExamples := []string{}
	processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		swaggerSpec:     swaggerSpec,
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	result, _ := os.ReadFile(filepath.Join(docsFolder, resourceName+".md"))
	resultStr := string(result)

	apiIdx := strings.Index(resultStr, "GET /api/v2/ordering")
	permsIdx := strings.Index(resultStr, "## Permissions and Scopes")
	notesIdx := strings.Index(resultStr, "## Notes")

	if apiIdx == -1 || permsIdx == -1 || notesIdx == -1 {
		t.Fatalf("Missing content sections. APIs=%d, Perms=%d, Notes=%d", apiIdx, permsIdx, notesIdx)
	}
	if !(apiIdx < permsIdx && permsIdx < notesIdx) {
		t.Errorf("Expected ordering APIs(%d) < Permissions(%d) < Notes(%d)", apiIdx, permsIdx, notesIdx)
	}
}

func TestProcessDocsFolder_NoSwaggerSpec(t *testing.T) {
	tempDir := t.TempDir()
	docsFolder := filepath.Join(tempDir, "docs")
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(docsFolder, 0755)

	resourceName := "no_swagger"
	os.WriteFile(filepath.Join(docsFolder, resourceName+".md"), []byte("# Resource\n\n**No APIs**\n"), 0644)

	resourceExamplesDir := filepath.Join(examplesFolder, "genesyscloud_"+resourceName)
	os.MkdirAll(resourceExamplesDir, 0755)
	os.WriteFile(filepath.Join(resourceExamplesDir, "apis.md"), []byte("* [GET /api/v2/noswagger](https://example.com)\n"), 0644)

	missingExamples := []string{}
	allPerms := processDocsFolder(docsFolderProcessor{
		docsFolder:      docsFolder,
		examplesFolder:  examplesFolder,
		apiDocsTag:      "**No APIs**",
		missingExamples: &missingExamples,
		errors:          &[]string{},
	})

	if len(allPerms) != 0 {
		t.Errorf("Expected no permissions without swagger spec, got %d", len(allPerms))
	}

	result, _ := os.ReadFile(filepath.Join(docsFolder, resourceName+".md"))
	resultStr := string(result)
	if strings.Contains(resultStr, "**No APIs**") {
		t.Error("Placeholder should still be replaced even without swagger")
	}
	if strings.Contains(resultStr, "## Permissions and Scopes") {
		t.Error("Permissions section should not appear without swagger spec")
	}
	if !strings.Contains(resultStr, "GET /api/v2/noswagger") {
		t.Error("API content should still be present")
	}
}

func TestLoadCachedSpec(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "swagger_cache.json")

	// No cache file should error
	_, err := loadCachedSpec(cacheFile)
	if err == nil {
		t.Fatal("Expected error when cache file does not exist")
	}

	// Valid cache file should load
	spec := SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/test": {
				"get": {OperationId: "getTest"},
			},
		},
	}
	data, _ := json.Marshal(spec)
	os.WriteFile(cacheFile, data, 0644)

	loaded, err := loadCachedSpec(cacheFile)
	if err != nil {
		t.Fatalf("Failed to load cached spec: %v", err)
	}
	if _, ok := loaded.Paths["/api/v2/test"]; !ok {
		t.Error("Expected /api/v2/test path in loaded spec")
	}

	// Invalid JSON should error
	os.WriteFile(cacheFile, []byte("not json"), 0644)
	_, err = loadCachedSpec(cacheFile)
	if err == nil {
		t.Fatal("Expected error for invalid JSON cache file")
	}
}

func TestFetchSwaggerSpec_UsesCacheOn304(t *testing.T) {
	// Create a test server that returns 304
	spec := SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/cached": {
				"get": {OperationId: "getCached"},
			},
		},
	}
	data, _ := json.Marshal(spec)

	// Write cache file
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "swagger_cache.json")
	os.WriteFile(cacheFile, data, 0644)

	// Temporarily override the cache file path
	origCacheFile := swaggerCacheFile
	// We can't easily override the const, so test loadCachedSpec directly
	_ = origCacheFile

	loaded, err := loadCachedSpec(cacheFile)
	if err != nil {
		t.Fatalf("Failed to load cached spec: %v", err)
	}
	if _, ok := loaded.Paths["/api/v2/cached"]; !ok {
		t.Error("Expected /api/v2/cached path in spec loaded from cache")
	}
}

func TestFetchSwaggerSpec_WritesCache(t *testing.T) {
	// Create a test HTTP server that serves a swagger spec
	spec := SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/fresh": {
				"post": {OperationId: "postFresh"},
			},
		},
	}
	data, _ := json.Marshal(spec)

	server := httpTestServer(t, http.StatusOK, data, "")
	defer server.Close()

	// Override cache file to temp location
	tempDir := t.TempDir()
	original := swaggerCacheFile
	// Since swaggerCacheFile is a package-level const we can't override it in the function,
	// but we can verify loadCachedSpec works with any path.
	_ = original

	// Directly test that fetchSwaggerSpec works with a real HTTP server
	// by temporarily pointing to a writable cache dir
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)
	os.MkdirAll("apidocs", 0755)

	result, err := fetchSwaggerSpec(server.URL)
	if err != nil {
		t.Fatalf("fetchSwaggerSpec failed: %v", err)
	}
	if _, ok := result.Paths["/api/v2/fresh"]; !ok {
		t.Error("Expected /api/v2/fresh in fetched spec")
	}

	// Verify cache was written
	cached, err := loadCachedSpec(filepath.Join(tempDir, "apidocs", "swagger_cache.json"))
	if err != nil {
		t.Fatalf("Cache file was not written: %v", err)
	}
	if _, ok := cached.Paths["/api/v2/fresh"]; !ok {
		t.Error("Cached spec missing expected path")
	}
}

func TestFetchSwaggerSpec_304UsesCache(t *testing.T) {
	// Pre-populate cache, then serve 304
	spec := SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/notmodified": {
				"get": {OperationId: "getNotModified"},
			},
		},
	}
	data, _ := json.Marshal(spec)

	server := httpTestServer(t, http.StatusNotModified, nil, "")
	defer server.Close()

	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)
	os.MkdirAll("apidocs", 0755)
	os.WriteFile("apidocs/swagger_cache.json", data, 0644)

	result, err := fetchSwaggerSpec(server.URL)
	if err != nil {
		t.Fatalf("fetchSwaggerSpec failed on 304: %v", err)
	}
	if _, ok := result.Paths["/api/v2/notmodified"]; !ok {
		t.Error("Expected cached spec to be returned on 304")
	}
}

// httpTestServer creates a simple test HTTP server for swagger spec testing.
func httpTestServer(t *testing.T, statusCode int, body []byte, lastModified string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if lastModified != "" {
			w.Header().Set("Last-Modified", lastModified)
		}
		w.WriteHeader(statusCode)
		if body != nil {
			w.Write(body)
		}
	}))
}

func TestStripSourcesComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes sources comment",
			input:    "<!-- sources\ngenesyscloud/foo/proxy.go\n-->\n* [GET /api/v2/foo](https://example.com)\n",
			expected: "* [GET /api/v2/foo](https://example.com)\n",
		},
		{
			name:     "no comment present",
			input:    "* [GET /api/v2/foo](https://example.com)\n",
			expected: "* [GET /api/v2/foo](https://example.com)\n",
		},
		{
			name:     "empty after stripping",
			input:    "<!-- sources\nfoo.go\n-->\n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripSourcesComment(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestWriteDocFile(t *testing.T) {
	tempDir := t.TempDir()

	// Test: placeholder gets replaced
	docPath := filepath.Join(tempDir, "resource.md")
	os.WriteFile(docPath, []byte("# Title\n\n**No APIs**\n\nFooter\n"), 0644)

	changed, err := writeDocFile(docPath, "**No APIs**", "* [GET /api/v2/test](url)\n")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !changed {
		t.Error("Expected changed=true when placeholder is replaced")
	}
	result, _ := os.ReadFile(docPath)
	if !strings.Contains(string(result), "GET /api/v2/test") {
		t.Error("Content not injected")
	}

	// Test: no change when placeholder is already gone
	changed, err = writeDocFile(docPath, "**No APIs**", "something else")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if changed {
		t.Error("Expected changed=false when placeholder not found")
	}
}

func TestExtractPermissionsAndScopes(t *testing.T) {
	spec := &SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/test": {
				"get": {
					XIninRequiresPermissions: PermissionsSpec{
						Permissions: []string{"test:view"},
					},
					Security: []map[string][]string{
						{"PureCloud OAuth": []string{"test:readonly"}},
					},
				},
			},
		},
	}

	// Endpoint exists in spec
	result := extractPermissionsAndScopes([]APIEndpoint{{Method: "get", Path: "/api/v2/test"}}, spec)
	if !strings.Contains(result, "`test:view`") {
		t.Error("Expected permission test:view")
	}
	if !strings.Contains(result, "`test:readonly`") {
		t.Error("Expected scope test:readonly")
	}

	// Endpoint not in spec — should return empty
	result = extractPermissionsAndScopes([]APIEndpoint{{Method: "get", Path: "/api/v2/missing"}}, spec)
	if result != "" {
		t.Errorf("Expected empty string for missing endpoint, got %q", result)
	}

	// Endpoint with no permissions or scopes
	specEmpty := &SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/empty": {
				"get": {},
			},
		},
	}
	result = extractPermissionsAndScopes([]APIEndpoint{{Method: "get", Path: "/api/v2/empty"}}, specEmpty)
	if result != "" {
		t.Errorf("Expected empty string for endpoint with no perms/scopes, got %q", result)
	}
}

func TestExtractResourcePermissions(t *testing.T) {
	spec := &SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/res": {
				"get": {
					XIninRequiresPermissions: PermissionsSpec{
						Permissions: []string{"res:view", "res:list"},
					},
					Security: []map[string][]string{
						{"PureCloud OAuth": []string{"res:readonly"}},
					},
				},
				"post": {
					XIninRequiresPermissions: PermissionsSpec{
						Permissions: []string{"res:add"},
					},
					Security: []map[string][]string{
						{"PureCloud OAuth": []string{"res"}},
					},
				},
			},
		},
	}

	endpoints := []APIEndpoint{
		{Method: "get", Path: "/api/v2/res"},
		{Method: "post", Path: "/api/v2/res"},
		{Method: "get", Path: "/api/v2/nonexistent"},
	}

	result := extractResourcePermissions("genesyscloud_res", "res", endpoints, spec)

	if result.ResourceType != "genesyscloud_res" {
		t.Errorf("Expected resource type 'genesyscloud_res', got %q", result.ResourceType)
	}
	if len(result.Permissions) != 3 {
		t.Errorf("Expected 3 permissions, got %d: %v", len(result.Permissions), result.Permissions)
	}
	if len(result.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d: %v", len(result.Scopes), result.Scopes)
	}
	if len(result.Endpoints) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(result.Endpoints))
	}
}

func TestCustomApiMethodToHTTP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MethodGet", "GET"},
		{"MethodPost", "POST"},
		{"MethodPut", "PUT"},
		{"MethodPatch", "PATCH"},
		{"MethodDelete", "DELETE"},
		{"MethodUnknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := customApiMethodToHTTP(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProcessAuditOnly(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	examplesFolder := filepath.Join(tempDir, "examples")
	resourceDir := filepath.Join(examplesFolder, "genesyscloud_test_res")
	os.MkdirAll(resourceDir, 0755)

	// Create proxy file
	proxyDir := filepath.Join("genesyscloud", "test_res")
	os.MkdirAll(proxyDir, 0755)
	proxyFile := filepath.Join(proxyDir, "genesyscloud_test_res_proxy.go")
	os.WriteFile(proxyFile, []byte(`package test_res
func (p *proxy) do() {
	p.testApi.GetThings()
}
`), 0644)

	// Create apis.md with sources but no endpoints
	apisMd := filepath.Join(resourceDir, "apis.md")
	os.WriteFile(apisMd, []byte("<!-- sources\n"+proxyFile+"\n-->\n"), 0644)

	opMap := map[string]APIEndpoint{
		"getthings": {Method: "GET", Path: "/api/v2/things"},
	}

	var errors []string
	processAuditOnly(docsFolderProcessor{
		examplesFolder: examplesFolder,
		opMap:          opMap,
		errors:         &errors,
	})

	// Verify endpoint was added
	result, _ := os.ReadFile(apisMd)
	if !strings.Contains(string(result), "GET /api/v2/things") {
		t.Errorf("Expected endpoint to be added, got: %s", string(result))
	}
	if len(errors) != 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
}

func TestProcessAuditOnly_SkipsIgnored(t *testing.T) {
	tempDir := t.TempDir()
	examplesFolder := filepath.Join(tempDir, "examples")
	os.MkdirAll(filepath.Join(examplesFolder, "genesyscloud_ignored"), 0755)
	os.WriteFile(filepath.Join(examplesFolder, "genesyscloud_ignored", "apis.md"), []byte("<!-- sources\nNO_SOURCES\n-->\n"), 0644)

	var errors []string
	processAuditOnly(docsFolderProcessor{
		examplesFolder:  examplesFolder,
		ignoredExamples: []string{"genesyscloud_ignored"},
		opMap:           map[string]APIEndpoint{},
		errors:          &errors,
	})

	if len(errors) != 0 {
		t.Errorf("Expected no errors for ignored resource, got: %v", errors)
	}
}

func TestProcessAuditOnly_NonexistentFolder(t *testing.T) {
	var errors []string
	result := processAuditOnly(docsFolderProcessor{
		examplesFolder: "/nonexistent",
		errors:         &errors,
	})
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestFetchSwaggerSpec_UnexpectedStatusFallsBackToCache(t *testing.T) {
	spec := SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/cached500": {
				"get": {OperationId: "getCached500"},
			},
		},
	}
	data, _ := json.Marshal(spec)

	server := httpTestServer(t, http.StatusInternalServerError, []byte("error"), "")
	defer server.Close()

	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)
	os.MkdirAll("apidocs", 0755)
	os.WriteFile("apidocs/swagger_cache.json", data, 0644)

	result, err := fetchSwaggerSpec(server.URL)
	if err != nil {
		t.Fatalf("Expected fallback to cache, got error: %v", err)
	}
	if _, ok := result.Paths["/api/v2/cached500"]; !ok {
		t.Error("Expected cached spec on unexpected status code")
	}
}

func TestFetchSwaggerSpec_UnexpectedStatusNoCacheErrors(t *testing.T) {
	server := httpTestServer(t, http.StatusInternalServerError, []byte("error"), "")
	defer server.Close()

	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)
	os.MkdirAll("apidocs", 0755)
	// No cache file

	_, err := fetchSwaggerSpec(server.URL)
	if err == nil {
		t.Fatal("Expected error when unexpected status and no cache")
	}
	if !strings.Contains(err.Error(), "unexpected status code 500") {
		t.Errorf("Expected status code in error, got: %v", err)
	}
}

func TestFetchSwaggerSpec_NetworkFailureFallsBackToCache(t *testing.T) {
	spec := SwaggerSpec{
		Paths: map[string]map[string]EndpointSpec{
			"/api/v2/offline": {
				"get": {OperationId: "getOffline"},
			},
		},
	}
	data, _ := json.Marshal(spec)

	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)
	os.MkdirAll("apidocs", 0755)
	os.WriteFile("apidocs/swagger_cache.json", data, 0644)

	// Use an unreachable URL to trigger network failure
	result, err := fetchSwaggerSpec("http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("Expected fallback to cache on network error, got: %v", err)
	}
	if _, ok := result.Paths["/api/v2/offline"]; !ok {
		t.Error("Expected cached spec on network failure")
	}
}

func TestFetchSwaggerSpec_NetworkFailureNoCacheErrors(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)
	os.MkdirAll("apidocs", 0755)
	// No cache file

	_, err := fetchSwaggerSpec("http://127.0.0.1:1")
	if err == nil {
		t.Fatal("Expected error when network fails and no cache")
	}
	if !strings.Contains(err.Error(), "no cache available") {
		t.Errorf("Expected 'no cache available' in error, got: %v", err)
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
