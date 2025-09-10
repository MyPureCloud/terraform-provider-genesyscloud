package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestFrameworkProvider(t *testing.T) {
	ctx := context.Background()

	// Create empty Framework resources and data sources for testing
	frameworkResources := make(map[string]func() resource.Resource)
	frameworkDataSources := make(map[string]func() datasource.DataSource)

	// Create Framework provider
	frameworkProvider := NewFrameworkProvider("test", frameworkResources, frameworkDataSources)()

	// Test Metadata
	metadataReq := provider.MetadataRequest{}
	metadataResp := &provider.MetadataResponse{}
	frameworkProvider.Metadata(ctx, metadataReq, metadataResp)

	if metadataResp.TypeName != "genesyscloud" {
		t.Errorf("Expected TypeName 'genesyscloud', got '%s'", metadataResp.TypeName)
	}

	if metadataResp.Version != "test" {
		t.Errorf("Expected Version 'test', got '%s'", metadataResp.Version)
	}

	// Test Schema
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	frameworkProvider.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
	}

	// Check for required attributes
	requiredAttrs := []string{
		"access_token",
		"oauthclient_id",
		"oauthclient_secret",
		"aws_region",
	}

	for _, attr := range requiredAttrs {
		if _, exists := schemaResp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute '%s' to be defined in schema", attr)
		}
	}

	// Test Resources (should be empty initially)
	resources := frameworkProvider.Resources(ctx)
	if len(resources) != 0 {
		t.Errorf("Expected 0 resources initially, got %d", len(resources))
	}

	// Test DataSources (should be empty initially)
	dataSources := frameworkProvider.DataSources(ctx)
	if len(dataSources) != 0 {
		t.Errorf("Expected 0 data sources initially, got %d", len(dataSources))
	}
}

func TestFrameworkProviderServer(t *testing.T) {
	ctx := context.Background()

	// Create empty Framework resources and data sources for testing
	frameworkResources := make(map[string]func() resource.Resource)
	frameworkDataSources := make(map[string]func() datasource.DataSource)

	// Create Framework provider server
	frameworkProvider := NewFrameworkProvider("test", frameworkResources, frameworkDataSources)

	// Create provider server
	serverFactory := providerserver.NewProtocol6(frameworkProvider())
	server := serverFactory()
	if server == nil {
		t.Error("Expected provider server to be created")
	}

	// Test GetProviderSchema
	schemaReq := &tfprotov6.GetProviderSchemaRequest{}
	schemaResp, err := server.GetProviderSchema(ctx, schemaReq)
	if err != nil {
		t.Errorf("GetProviderSchema failed: %v", err)
	}

	if schemaResp.Provider == nil {
		t.Error("Expected provider schema to be returned")
	}
}

func TestGetStringValue(t *testing.T) {
	// Test cases for getStringValue helper function
	testCases := []struct {
		name     string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "environment variable set",
			envVar:   "TEST_VAR",
			envValue: "test_value",
			expected: "test_value",
		},
		{
			name:     "environment variable not set",
			envVar:   "NONEXISTENT_VAR",
			envValue: "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				t.Setenv(tc.envVar, tc.envValue)
			}

			// Create null types.String to test environment fallback
			var nullString = types.StringNull()
			result := getStringValue(nullString, tc.envVar)

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
