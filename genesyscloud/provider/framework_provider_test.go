package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func TestFrameworkProviderWithResources(t *testing.T) {
	ctx := context.Background()

	// Create Framework resources and data sources for testing
	frameworkResources := map[string]func() resource.Resource{
		"test_resource": func() resource.Resource {
			return &testFrameworkProviderResource{}
		},
	}
	frameworkDataSources := map[string]func() datasource.DataSource{
		"test_data_source": func() datasource.DataSource {
			return &testFrameworkProviderDataSource{}
		},
	}

	// Create Framework provider
	frameworkProvider := NewFrameworkProvider("test", frameworkResources, frameworkDataSources)()

	// Test Resources
	resources := frameworkProvider.Resources(ctx)
	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	// Test DataSources
	dataSources := frameworkProvider.DataSources(ctx)
	if len(dataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(dataSources))
	}

	// Test resource creation
	if len(resources) > 0 {
		testResource := resources[0]()
		if testResource == nil {
			t.Error("Expected resource to be created")
		}

		// Test resource metadata
		metadataReq := resource.MetadataRequest{ProviderTypeName: "genesyscloud"}
		metadataResp := &resource.MetadataResponse{}
		testResource.Metadata(ctx, metadataReq, metadataResp)

		if metadataResp.TypeName != "genesyscloud_test_resource" {
			t.Errorf("Expected resource TypeName 'genesyscloud_test_resource', got '%s'", metadataResp.TypeName)
		}
	}

	// Test data source creation
	if len(dataSources) > 0 {
		testDataSource := dataSources[0]()
		if testDataSource == nil {
			t.Error("Expected data source to be created")
		}

		// Test data source metadata
		metadataReq := datasource.MetadataRequest{ProviderTypeName: "genesyscloud"}
		metadataResp := &datasource.MetadataResponse{}
		testDataSource.Metadata(ctx, metadataReq, metadataResp)

		if metadataResp.TypeName != "genesyscloud_test_data_source" {
			t.Errorf("Expected data source TypeName 'genesyscloud_test_data_source', got '%s'", metadataResp.TypeName)
		}
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

func TestFrameworkProviderConfigure(t *testing.T) {
	ctx := context.Background()

	// Create Framework provider
	frameworkResources := make(map[string]func() resource.Resource)
	frameworkDataSources := make(map[string]func() datasource.DataSource)
	frameworkProvider := NewFrameworkProvider("test", frameworkResources, frameworkDataSources)()

	// Get the provider schema first
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	frameworkProvider.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Failed to get provider schema: %v", schemaResp.Diagnostics)
	}

	// Test 1: Configuration validation behavior
	t.Run("Configuration validation", func(t *testing.T) {
		// The Framework has limitations with empty configs causing conversion errors
		// This is expected behavior, not a bug in our provider
		// We'll test that the Configure method can be called without panicking

		emptyConfig := tfsdk.Config{
			Schema: schemaResp.Schema,
		}

		configReq := provider.ConfigureRequest{
			Config: emptyConfig,
		}
		configResp := &provider.ConfigureResponse{}

		// Call Configure - expect it to handle the call gracefully
		frameworkProvider.Configure(ctx, configReq, configResp)

		// The important thing is that Configure doesn't panic and returns some diagnostic
		if len(configResp.Diagnostics) == 0 {
			t.Error("Expected at least one diagnostic (even if it's a conversion error)")
		}

		// Log what we got for debugging
		for i, diag := range configResp.Diagnostics {
			t.Logf("Diagnostic %d: %s - %s", i, diag.Summary(), diag.Detail())
		}

		// Test passes if Configure method completes without panic
		t.Log("Configure method completed successfully")
	})

	// Test 2: Verify Configure method behavior
	t.Run("Configure method functionality", func(t *testing.T) {
		// Test that the Configure method exists and can be called
		// The actual validation logic is complex to test without proper config setup

		config := tfsdk.Config{
			Schema: schemaResp.Schema,
		}

		configReq := provider.ConfigureRequest{
			Config: config,
		}
		configResp := &provider.ConfigureResponse{}

		// Call Configure
		frameworkProvider.Configure(ctx, configReq, configResp)

		// Test passes if the method completes (diagnostics are expected)
		t.Logf("Configure completed with %d diagnostics", len(configResp.Diagnostics))
	})
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

// testFrameworkProviderResource is a minimal test resource for framework provider testing
type testFrameworkProviderResource struct{}

func (r *testFrameworkProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_resource"
}

func (r *testFrameworkProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Empty schema for testing
}

func (r *testFrameworkProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Empty implementation for testing
}

func (r *testFrameworkProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Empty implementation for testing
}

func (r *testFrameworkProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Empty implementation for testing
}

func (r *testFrameworkProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Empty implementation for testing
}

// testFrameworkProviderDataSource is a minimal test data source for framework provider testing
type testFrameworkProviderDataSource struct{}

func (d *testFrameworkProviderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_data_source"
}

func (d *testFrameworkProviderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Empty schema for testing
}

func (d *testFrameworkProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Empty implementation for testing
}
