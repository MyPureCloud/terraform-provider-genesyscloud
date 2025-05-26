package outbound_contact_list

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func TestContactListBuildSdkOutboundContactListContactPhoneNumberColumnSlice(t *testing.T) {
	// Create two phone number columns
	phoneNumberColumns := []map[string]interface{}{
		{
			"column_name":          "phone1",
			"type":                 "home",
			"callable_time_column": "call_time1",
		},
		{
			"column_name":          "phone2",
			"type":                 "mobile",
			"callable_time_column": "call_time2",
		},
	}

	// Create a new schema.Set with both phone number columns
	phoneSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"column_name": {
				Type: schema.TypeString,
			},
			"type": {
				Type: schema.TypeString,
			},
			"callable_time_column": {
				Type: schema.TypeString,
			},
		},
	}), []interface{}{})

	// Add both columns to the set
	for _, column := range phoneNumberColumns {
		phoneSet.Add(column)
	}

	// Call the function being tested
	result := buildSdkOutboundContactListContactPhoneNumberColumnSlice(phoneSet)

	// Verify the result has exactly 2 elements
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(*result) != 2 {
		t.Errorf("Expected slice with 2 elements, got %d elements", len(*result))
	}

	// Verify the contents of both elements
	phoneNumbers := *result
	expectedNames := map[string]bool{"phone1": false, "phone2": false}
	expectedTypes := map[string]bool{"home": false, "mobile": false}
	expectedCallTimes := map[string]bool{"call_time1": false, "call_time2": false}

	for _, phone := range phoneNumbers {
		if phone.ColumnName != nil {
			expectedNames[*phone.ColumnName] = true
		}
		if phone.VarType != nil {
			expectedTypes[*phone.VarType] = true
		}
		if phone.CallableTimeColumn != nil {
			expectedCallTimes[*phone.CallableTimeColumn] = true
		}
	}

	// Verify all expected values were found
	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected to find column_name %s", name)
		}
	}
	for phoneType, found := range expectedTypes {
		if !found {
			t.Errorf("Expected to find type %s", phoneType)
		}
	}
	for callTime, found := range expectedCallTimes {
		if !found {
			t.Errorf("Expected to find callable_time_column %s", callTime)
		}
	}
}

func TestContactListBuildSdkOutboundContactListContactEmailAddressColumnSlice(t *testing.T) {
	// Create test data with two email columns
	emailColumns := []interface{}{
		map[string]interface{}{
			"column_name":             "email1",
			"type":                    "email",
			"contactable_time_column": "time1",
		},
		map[string]interface{}{
			"column_name":             "email2",
			"type":                    "email",
			"contactable_time_column": "time2",
		},
	}

	// Create a new schema set with the test data
	emailColumnSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"column_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"contactable_time_column": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}), emailColumns)

	// Call the function being tested
	result := buildSdkOutboundContactListContactEmailAddressColumnSlice(emailColumnSet)

	// Verify the result
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(*result) != 2 {
		t.Errorf("Expected slice with 2 elements, got %d", len(*result))
	}

	// Verify the contents of each element
	for i, expected := range emailColumns {
		expectedMap := expected.(map[string]interface{})
		actual := (*result)[i]

		if *actual.ColumnName != expectedMap["column_name"] {
			t.Errorf("Element %d: expected column_name %v, got %v",
				i, expectedMap["column_name"], *actual.ColumnName)
		}

		if *actual.VarType != expectedMap["type"] {
			t.Errorf("Element %d: expected type %v, got %v",
				i, expectedMap["type"], *actual.VarType)
		}

		if *actual.ContactableTimeColumn != expectedMap["contactable_time_column"] {
			t.Errorf("Element %d: expected contactable_time_column %v, got %v",
				i, expectedMap["contactable_time_column"], *actual.ContactableTimeColumn)
		}
	}
}
func TestContactListBuildSdkOutboundContactListContactEmailAddressColumnSliceEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    []interface{}
		expected int
	}{
		{
			name: "Missing optional field",
			input: []interface{}{
				map[string]interface{}{
					"column_name": "email1",
					"type":        "email",
				},
			},
			expected: 1,
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: 0,
		},
		{
			name: "Nil values",
			input: []interface{}{
				map[string]interface{}{
					"column_name":             nil,
					"type":                    nil,
					"contactable_time_column": nil,
				},
			},
			expected: 1,
		},
	}

	schemaRes := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"column_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"contactable_time_column": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			emailColumnSet := schema.NewSet(schema.HashResource(schemaRes), tc.input)

			// Use defer to catch any panics
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test case %s panicked: %v", tc.name, r)
				}
			}()

			result := buildSdkOutboundContactListContactEmailAddressColumnSlice(emailColumnSet)

			// Check if result is nil when expected
			if tc.expected == 0 {
				if result != nil && len(*result) != 0 {
					t.Errorf("Expected empty result, got %d elements", len(*result))
				}
				return
			}

			// Verify non-nil results
			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if len(*result) != tc.expected {
				t.Errorf("Expected slice with %d elements, got %d", tc.expected, len(*result))
			}

			// For non-empty results, verify the structure exists
			if len(*result) > 0 {
				actual := (*result)[0]

				// Only verify non-nil fields
				if actual.ColumnName != nil {
					if *actual.ColumnName == "" {
						t.Error("Expected non-empty column_name")
					}
				}

				if actual.VarType != nil {
					if *actual.VarType == "" {
						t.Error("Expected non-empty type")
					}
				}

				// ContactableTimeColumn is optional, no need to verify
			}
		})
	}
}

func TestContactListFlattenSdkOutboundContactListContactEmailAddressColumnSlice(t *testing.T) {
	// Test case 1: Empty input
	emptyResult := flattenSdkOutboundContactListContactEmailAddressColumnSlice([]platformclientv2.Emailcolumn{})
	if emptyResult != nil {
		t.Errorf("Expected nil for empty input, got %v", emptyResult)
	}

	// Helper function to create string pointer
	strPtr := func(s string) *string {
		return &s
	}

	// Test case 2: Single email column with all fields
	singleColumn := []platformclientv2.Emailcolumn{
		{
			ColumnName:            strPtr("email_col"),
			VarType:               strPtr("email"),
			ContactableTimeColumn: strPtr("time_col"),
		},
	}

	result := flattenSdkOutboundContactListContactEmailAddressColumnSlice(singleColumn)

	if result == nil {
		t.Fatal("Expected non-nil result for single column")
	}

	if result.Len() != 1 {
		t.Errorf("Expected 1 item in set, got %d", result.Len())
	}

	// Convert set to slice to check values
	resultList := result.List()
	if len(resultList) != 1 {
		t.Fatal("Expected 1 item in result list")
	}

	resultMap := resultList[0].(map[string]interface{})

	expectedValues := map[string]string{
		"column_name":             "email_col",
		"type":                    "email",
		"contactable_time_column": "time_col",
	}

	for key, expectedVal := range expectedValues {
		if val, ok := resultMap[key]; !ok || val != expectedVal {
			t.Errorf("Expected %s to be %s, got %v", key, expectedVal, val)
		}
	}

	// Test case 3: Multiple columns with partial fields
	multipleColumns := []platformclientv2.Emailcolumn{
		{
			ColumnName: strPtr("email1"),
			VarType:    strPtr("email"),
		},
		{
			ColumnName:            strPtr("email2"),
			VarType:               strPtr("email"),
			ContactableTimeColumn: strPtr("time2"),
		},
	}

	multiResult := flattenSdkOutboundContactListContactEmailAddressColumnSlice(multipleColumns)

	if multiResult == nil {
		t.Fatal("Expected non-nil result for multiple columns")
	}

	if multiResult.Len() != 2 {
		t.Errorf("Expected 2 items in set, got %d", multiResult.Len())
	}
}

func TestContactListBuildSdkOutboundContactListColumnDataTypeSpecifications(t *testing.T) {

	// Helper functions for creating pointers
	stringPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	// Test cases
	tests := []struct {
		name                         string
		columnDataTypeSpecifications []interface{}
		expected                     *[]platformclientv2.Columndatatypespecification
	}{
		{
			name:                         "nil input",
			columnDataTypeSpecifications: nil,
			expected:                     nil,
		},
		{
			name:                         "empty input",
			columnDataTypeSpecifications: []interface{}{},
			expected:                     nil,
		},
		{
			name: "single specification with all fields",
			columnDataTypeSpecifications: []interface{}{
				map[string]interface{}{
					"column_name":      "test_column",
					"column_data_type": "text",
					"min":              1,
					"max":              100,
					"max_length":       50,
				},
			},
			expected: &[]platformclientv2.Columndatatypespecification{
				{
					ColumnName:     stringPtr("test_column"),
					ColumnDataType: stringPtr("text"),
					Min:            intPtr(1),
					Max:            intPtr(100),
					MaxLength:      intPtr(50),
				},
			},
		},
		{
			name: "multiple specifications with partial fields",
			columnDataTypeSpecifications: []interface{}{
				map[string]interface{}{
					"column_name":      "col1",
					"column_data_type": "number",
				},
				map[string]interface{}{
					"column_name":      "col2",
					"column_data_type": "text",
					"max_length":       20,
				},
			},
			expected: &[]platformclientv2.Columndatatypespecification{
				{
					ColumnName:     stringPtr("col1"),
					ColumnDataType: stringPtr("number"),
				},
				{
					ColumnName:     stringPtr("col2"),
					ColumnDataType: stringPtr("text"),
					MaxLength:      intPtr(20),
				},
			},
		},
		{
			name: "empty column_data_type should be omitted",
			columnDataTypeSpecifications: []interface{}{
				map[string]interface{}{
					"column_name":      "col1",
					"column_data_type": "",
				},
			},
			expected: &[]platformclientv2.Columndatatypespecification{
				{
					ColumnName: stringPtr("col1"),
				},
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSdkOutboundContactListColumnDataTypeSpecifications(tt.columnDataTypeSpecifications)

			// Handle nil cases
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			// Compare lengths
			if len(*result) != len(*tt.expected) {
				t.Errorf("expected length %d, got length %d", len(*tt.expected), len(*result))
				return
			}

			// Compare individual specifications
			for i := range *result {
				actual := (*result)[i]
				expected := (*tt.expected)[i]

				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("specification at index %d differs.\nexpected: %+v\ngot: %+v", i, expected, actual)
				}
			}
		})
	}
}

func TestContactListFlattenSdkOutboundContactListColumnDataTypeSpecifications(t *testing.T) {
	// Test 1: Empty input
	emptyResult := flattenSdkOutboundContactListColumnDataTypeSpecifications([]platformclientv2.Columndatatypespecification{})
	if emptyResult != nil {
		t.Errorf("Expected nil for empty input, got %v", emptyResult)
	}

	// Test 2: Single column with all fields populated
	columnName := "test_column"
	dataType := "text"
	min := int(1)
	max := int(100)
	maxLength := int(50)

	input := []platformclientv2.Columndatatypespecification{
		{
			ColumnName:     &columnName,
			ColumnDataType: &dataType,
			Min:            &min,
			Max:            &max,
			MaxLength:      &maxLength,
		},
	}

	result := flattenSdkOutboundContactListColumnDataTypeSpecifications(input)

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	// Type assert and verify all fields
	if spec, ok := result[0].(map[string]interface{}); ok {
		if spec["column_name"] != columnName {
			t.Errorf("Expected column_name %s, got %v", columnName, spec["column_name"])
		}
		if spec["column_data_type"] != dataType {
			t.Errorf("Expected column_data_type %s, got %v", dataType, spec["column_data_type"])
		}
		if spec["min"] != min {
			t.Errorf("Expected min %d, got %v", min, spec["min"])
		}
		if spec["max"] != max {
			t.Errorf("Expected max %d, got %v", max, spec["max"])
		}
		if spec["max_length"] != maxLength {
			t.Errorf("Expected max_length %d, got %v", maxLength, spec["max_length"])
		}
	} else {
		t.Error("Failed to type assert result to map[string]interface{}")
	}

	// Test 3: Column with only required fields
	minimalColumnName := "minimal_column"
	minimalDataType := "number"
	minimalInput := []platformclientv2.Columndatatypespecification{
		{
			ColumnName:     &minimalColumnName,
			ColumnDataType: &minimalDataType,
		},
	}

	minimalResult := flattenSdkOutboundContactListColumnDataTypeSpecifications(minimalInput)

	if len(minimalResult) != 1 {
		t.Errorf("Expected 1 result, got %d", len(minimalResult))
	}

	if spec, ok := minimalResult[0].(map[string]interface{}); ok {
		if spec["column_name"] != minimalColumnName {
			t.Errorf("Expected column_name %s, got %v", minimalColumnName, spec["column_name"])
		}
		if spec["column_data_type"] != minimalDataType {
			t.Errorf("Expected column_data_type %s, got %v", minimalDataType, spec["column_data_type"])
		}
		// Verify optional fields are not present
		if _, exists := spec["min"]; exists {
			t.Error("Expected min to not be present")
		}
		if _, exists := spec["max"]; exists {
			t.Error("Expected max to not be present")
		}
		if _, exists := spec["max_length"]; exists {
			t.Error("Expected max_length to not be present")
		}
	} else {
		t.Error("Failed to type assert result to map[string]interface{}")
	}
}

func TestContactListContactsExporterResolver(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	subDir := "test_subdir"

	// Mock the config map
	configMap := map[string]interface{}{
		"contact_list_id": "test-contact-list",
	}

	t.Run("successful export", func(t *testing.T) {
		// Create test proxy with our test implementation
		testProxy := &OutboundContactlistProxy{
			initiateContactListContactsExportAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (*platformclientv2.APIResponse, error) {
				// Mock the API response
				resp := &platformclientv2.APIResponse{
					StatusCode: http.StatusOK,
				}
				return resp, nil
			},
			getContactListContactsExportUrlAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (string, *platformclientv2.APIResponse, error) {
				// Mock the API response
				resp := &platformclientv2.APIResponse{
					StatusCode: http.StatusOK,
				}
				return "http://test-url.com/export", resp, nil
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		// Mock the provider meta
		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		// Mock resource info
		mockResource := resourceExporter.ResourceInfo{
			BlockLabel: "contacts_test-contact-list",
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		// Mock the file download function
		origDownloadFile := files.DownloadExportFileWithAccessToken
		files.DownloadExportFileWithAccessToken = func(directory, filename, url, accessToken string) (*platformclientv2.APIResponse, error) {
			fullPath := filepath.Join(directory, filename)
			if err := os.MkdirAll(directory, os.ModePerm); err != nil {
				return nil, err
			}
			os.WriteFile(fullPath, []byte("test content"), 0644)
			return nil, nil
		}
		defer func() { files.DownloadExportFileWithAccessToken = origDownloadFile }()

		// Test the function
		err := ContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify the filepath was set in configMap
		expectedPath := filepath.Join(subDir, "contacts_test-contact-list.csv")
		if configMap["contacts_filepath"] != expectedPath {
			t.Errorf("Expected filepath %s, got %s", expectedPath, configMap["filepath"])
		}

		if configMap["contacts_id_name"] != "inin-outbound-id" {
			t.Errorf("Expected contacts_id_name to be 'inin-outbound-id', got %s", configMap["contacts_id_name"])
		}

		// Verify computed attributes not set on configMap
		if _, exists := configMap["contacts_file_content_hash"]; exists {
			t.Errorf("Expected contacts_file_content_hash to not be in configMap")
		}
		if _, exists := configMap["contacts_record_count"]; exists {
			t.Errorf("Expected contacts_record_count to not be in configMap")
		}

		// Verify state attributes set
		if mockResource.State.Attributes["contacts_file_content_hash"] == "" {
			t.Error("Expected contacts_file_content_hash to be set")
		}
		if mockResource.State.Attributes["contacts_record_count"] == "" {
			t.Error("Expected contacts_record_count to be set")
		}
		if mockResource.State.Attributes["contacts_filepath"] == "" {
			t.Error("Expected contacts_filepath to be set")
		}
		if mockResource.State.Attributes["contacts_id_name"] == "" {
			t.Error("Expected contacts_id_name to be set")
		}

	})

	t.Run("initiate export url error", func(t *testing.T) {
		// Create test proxy with error case

		testProxy := &OutboundContactlistProxy{
			initiateContactListContactsExportAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (*platformclientv2.APIResponse, error) {
				return nil, fmt.Errorf("failed to initiate export")
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		mockResource := resourceExporter.ResourceInfo{
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		err := ContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("get export url error", func(t *testing.T) {
		// Create test proxy with error case

		testProxy := &OutboundContactlistProxy{
			initiateContactListContactsExportAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (*platformclientv2.APIResponse, error) {
				// Mock the API response
				resp := &platformclientv2.APIResponse{
					StatusCode: http.StatusOK,
				}
				return resp, nil
			},
			getContactListContactsExportUrlAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (string, *platformclientv2.APIResponse, error) {
				return "", nil, fmt.Errorf("failed to get export URL")
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		mockResource := resourceExporter.ResourceInfo{
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		err := ContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("download error", func(t *testing.T) {
		// Create test proxy
		testProxy := &OutboundContactlistProxy{
			initiateContactListContactsExportAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (*platformclientv2.APIResponse, error) {
				// Mock the API response
				resp := &platformclientv2.APIResponse{
					StatusCode: http.StatusOK,
				}
				return resp, nil
			},
			getContactListContactsExportUrlAttr: func(_ context.Context, p *OutboundContactlistProxy, contactListId string) (string, *platformclientv2.APIResponse, error) {
				// Mock the API response
				resp := &platformclientv2.APIResponse{
					StatusCode: http.StatusOK,
				}
				return "http://test-url.com/export", resp, nil
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		mockResource := resourceExporter.ResourceInfo{
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		// Mock download failure
		files.DownloadExportFileWithAccessToken = func(directory, filename, url, accessToken string) (*platformclientv2.APIResponse, error) {
			return nil, fmt.Errorf("download failed")
		}

		err := ContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	// Clean up after all tests
	internalProxy = nil
}
