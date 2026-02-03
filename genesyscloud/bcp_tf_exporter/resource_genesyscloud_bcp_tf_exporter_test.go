package bcp_tf_exporter

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/stretchr/testify/assert"
)

func setupMockExporters() (map[string]*resourceExporter.ResourceExporter, func()) {
	mockExporters := map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_user": {
			GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				return resourceExporter.ResourceIDMetaMap{
					"user1": {BlockLabel: "test_user", OriginalLabel: "Test User"},
				}, nil
			},
		},
		"genesyscloud_group": {
			GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				return resourceExporter.ResourceIDMetaMap{
					"group1": {BlockLabel: "test_group", OriginalLabel: "Test Group"},
				}, nil
			},
			RefAttrs: map[string]*resourceExporter.RefAttrSettings{
				"owner_id": {RefType: "genesyscloud_user"},
			},
		},
		"genesyscloud_flow": {
			GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				return resourceExporter.ResourceIDMetaMap{
					"flow1": {BlockLabel: "test_flow", OriginalLabel: "Test Flow"},
				}, nil
			},
		},
	}

	original := resourceExporter.GetResourceExporters()
	resourceExporter.SetRegisterExporter(mockExporters)
	return mockExporters, func() { resourceExporter.SetRegisterExporter(original) }
}

func TestBcpTfExporter_Basic(t *testing.T) {
	_, cleanup := setupMockExporters()
	defer cleanup()
	tempDir := t.TempDir()
	filename := "test_export.json"

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"directory": tempDir,
		"filename":  filename,
	})

	diags := createBcpTfExporter(context.Background(), d, &provider.ProviderMeta{})
	assert.False(t, diags.HasError())

	filePath := filepath.Join(tempDir, filename)
	assert.FileExists(t, filePath)
	assert.Equal(t, filePath, d.Id())
}

func TestBcpTfExporter_WithIncludeFilter(t *testing.T) {
	_, cleanup := setupMockExporters()
	defer cleanup()

	tempDir := t.TempDir()
	filename := "filtered_export.json"

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"directory":                tempDir,
		"filename":                 filename,
		"include_filter_resources": []interface{}{"genesyscloud_user", "genesyscloud_group"},
	})

	diags := createBcpTfExporter(context.Background(), d, &provider.ProviderMeta{})
	assert.False(t, diags.HasError())

	filePath := filepath.Join(tempDir, filename)
	assert.FileExists(t, filePath)

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var exportData BcpExportData
	err = json.Unmarshal(data, &exportData)
	assert.NoError(t, err)

	// Should only contain included resource types
	for resourceType := range exportData {
		assert.Contains(t, []string{"genesyscloud_user", "genesyscloud_group"}, resourceType)
	}
}

func TestBcpTfExporter_WithExcludeFilter(t *testing.T) {
	_, cleanup := setupMockExporters()
	defer cleanup()

	tempDir := t.TempDir()
	filename := "excluded_export.json"

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"directory":                tempDir,
		"filename":                 filename,
		"exclude_filter_resources": []interface{}{"genesyscloud_flow", "genesyscloud_architect_datatable"},
	})

	diags := createBcpTfExporter(context.Background(), d, &provider.ProviderMeta{})
	assert.False(t, diags.HasError())

	filePath := filepath.Join(tempDir, filename)
	assert.FileExists(t, filePath)

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var exportData BcpExportData
	err = json.Unmarshal(data, &exportData)
	assert.NoError(t, err)

	// Should not contain excluded resource types
	assert.NotContains(t, exportData, "genesyscloud_flow")
	assert.NotContains(t, exportData, "genesyscloud_architect_datatable")
}

func TestBcpTfExporter_NoFilters(t *testing.T) {
	_, cleanup := setupMockExporters()
	defer cleanup()

	tempDir := t.TempDir()
	filename := "all_export.json"

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"directory": tempDir,
		"filename":  filename,
	})

	diags := createBcpTfExporter(context.Background(), d, &provider.ProviderMeta{})
	assert.False(t, diags.HasError())

	filePath := filepath.Join(tempDir, filename)
	assert.FileExists(t, filePath)

	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var exportData BcpExportData
	err = json.Unmarshal(data, &exportData)
	assert.NoError(t, err)

	// Should contain all resource types
	assert.Len(t, exportData, 3)
	assert.Contains(t, exportData, "genesyscloud_user")
	assert.Contains(t, exportData, "genesyscloud_group")
	assert.Contains(t, exportData, "genesyscloud_flow")
}

func TestBcpTfExporter_JSONStructure(t *testing.T) {
	// This test validates the basic structure with empty dependencies
	// since full resource reading is complex to mock in unit tests
	mockExporters := map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_user": {
			GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				return resourceExporter.ResourceIDMetaMap{
					"user1": {BlockLabel: "test_user", OriginalLabel: "Test User"},
				}, nil
			},
		},
		"genesyscloud_group": {
			GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				return resourceExporter.ResourceIDMetaMap{
					"group1": {BlockLabel: "test_group", OriginalLabel: "Test Group"},
				}, nil
			},
			RefAttrs: map[string]*resourceExporter.RefAttrSettings{
				"owner_id": {RefType: "genesyscloud_user"},
			},
		},
	}

	original := resourceExporter.GetResourceExporters()
	resourceExporter.SetRegisterExporter(mockExporters)
	defer resourceExporter.SetRegisterExporter(original)

	tempDir := t.TempDir()
	filename := "structure_test.json"

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"directory":                tempDir,
		"filename":                 filename,
		"include_filter_resources": []interface{}{"genesyscloud_user", "genesyscloud_group"},
	})

	diags := createBcpTfExporter(context.Background(), d, &provider.ProviderMeta{})
	assert.False(t, diags.HasError())

	filePath := filepath.Join(tempDir, filename)
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var exportData BcpExportData
	err = json.Unmarshal(data, &exportData)
	assert.NoError(t, err)

	// Verify exact JSON structure matches mock data
	assert.Len(t, exportData, 2)
	assert.Contains(t, exportData, "genesyscloud_user")
	assert.Contains(t, exportData, "genesyscloud_group")

	// Test user resource (no dependencies)
	userResources := exportData["genesyscloud_user"]
	assert.Len(t, userResources, 1)
	user := userResources[0]
	assert.Equal(t, "user1", user.ID)
	assert.Equal(t, "test_user", user.Name)
	assert.Equal(t, user.Dependencies, BcpResourceDependency{
		AsProviderResourceList: []string{},
		AsObjectMap:            map[string][]string{},
	})

	// Test group resource - will have empty dependencies since resource reading fails in test
	groupResources := exportData["genesyscloud_group"]
	assert.Len(t, groupResources, 1)
	group := groupResources[0]
	assert.Equal(t, "group1", group.ID)
	assert.Equal(t, "test_group", group.Name)
	// Should have empty dependencies since resource reading fails in test
	assert.Equal(t, group.Dependencies, BcpResourceDependency{
		AsProviderResourceList: []string{},
		AsObjectMap:            map[string][]string{},
	})
}

func TestBcpTfExporter_Read(t *testing.T) {
	tempDir := t.TempDir()
	filename := "read_test.json"
	filePath := filepath.Join(tempDir, filename)

	// Create test file
	testData := BcpExportData{
		"genesyscloud_user": []BcpResource{
			{ID: "test-id", Name: "test-name", Dependencies: BcpResourceDependency{}},
		},
	}
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{})
	d.SetId(filePath)

	diags := readBcpTfExporter(context.Background(), d, nil)
	assert.False(t, diags.HasError())
	assert.Equal(t, filePath, d.Id())
}

func TestBcpTfExporter_ReadMissingFile(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{})
	d.SetId("/nonexistent/file.json")

	diags := readBcpTfExporter(context.Background(), d, nil)
	assert.False(t, diags.HasError())
	assert.Empty(t, d.Id())
}

func TestBcpTfExporter_Delete(t *testing.T) {
	tempDir := t.TempDir()
	filename := "delete_test.json"
	filePath := filepath.Join(tempDir, filename)

	// Create test file
	err := os.WriteFile(filePath, []byte("{}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{})
	d.SetId(filePath)

	diags := deleteBcpTfExporter(context.Background(), d, nil)
	assert.False(t, diags.HasError())
	assert.NoFileExists(t, filePath)
}

func TestBcpTfExporter_FilterExporters_IncludeOnly(t *testing.T) {
	mockExporters := map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_user":  {},
		"genesyscloud_group": {},
		"genesyscloud_flow":  {},
	}

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"include_filter_resources": []interface{}{"genesyscloud_user", "genesyscloud_group"},
	})

	filtered := filterExporters(context.Background(), mockExporters, d)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, "genesyscloud_user")
	assert.Contains(t, filtered, "genesyscloud_group")
	assert.NotContains(t, filtered, "genesyscloud_flow")
}

func TestBcpTfExporter_FilterExporters_ExcludeOnly(t *testing.T) {
	mockExporters := map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_user":  {},
		"genesyscloud_group": {},
		"genesyscloud_flow":  {},
	}

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{
		"exclude_filter_resources": []interface{}{"genesyscloud_flow"},
	})

	filtered := filterExporters(context.Background(), mockExporters, d)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, "genesyscloud_user")
	assert.Contains(t, filtered, "genesyscloud_group")
	assert.NotContains(t, filtered, "genesyscloud_flow")
}

func TestBcpTfExporter_FilterExporters_NoFilters(t *testing.T) {
	mockExporters := map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_user":  {},
		"genesyscloud_group": {},
		"genesyscloud_flow":  {},
	}

	d := schema.TestResourceDataRaw(t, ResourceBcpTfExporter().Schema, map[string]interface{}{})

	filtered := filterExporters(context.Background(), mockExporters, d)

	assert.Len(t, filtered, 3)
	assert.Contains(t, filtered, "genesyscloud_user")
	assert.Contains(t, filtered, "genesyscloud_group")
	assert.Contains(t, filtered, "genesyscloud_flow")
}

func TestBcpTfExporter_ExtractGUIDsFromValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "single string GUID",
			input:    "12345678-1234-1234-1234-123456789012",
			expected: []string{"12345678-1234-1234-1234-123456789012"},
		},
		{
			name:     "array of GUIDs",
			input:    []interface{}{"87654321-4321-4321-4321-210987654321", "11111111-2222-3333-4444-555555555555"},
			expected: []string{"87654321-4321-4321-4321-210987654321", "11111111-2222-3333-4444-555555555555"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "non-string value",
			input:    123,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGUIDsFromValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBcpTfExporter_IsValidGUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid GUID",
			input:    "12345678-1234-1234-1234-123456789012",
			expected: true,
		},
		{
			name:     "invalid GUID - too short",
			input:    "12345678-1234-1234-1234-12345678901",
			expected: false,
		},
		{
			name:     "invalid GUID - missing hyphens",
			input:    "123456781234123412341234567890123",
			expected: false,
		},
		{
			name:     "invalid GUID - wrong hyphen positions",
			input:    "1234567-81234-1234-1234-123456789012",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGUID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
