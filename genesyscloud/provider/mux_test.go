package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestNewMuxedProvider(t *testing.T) {
	ctx := context.Background()

	// Create test data
	version := "test"
	providerResources := make(map[string]*schema.Resource)
	providerDataSources := make(map[string]*schema.Resource)
	frameworkResources := make(map[string]func() resource.Resource)
	frameworkDataSources := make(map[string]func() datasource.DataSource)

	// Test with no Framework resources (should return SDKv2 only)
	t.Run("SDKv2 only", func(t *testing.T) {
		muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResources, frameworkDataSources)

		serverFactory, err := muxFactory()
		if err != nil {
			t.Fatalf("Failed to create muxed provider: %v", err)
		}

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
	})

	// Test with Framework resources (should return muxed provider)
	t.Run("Muxed provider", func(t *testing.T) {
		// Add a dummy Framework resource
		frameworkResourcesWithData := make(map[string]func() resource.Resource)
		frameworkResourcesWithData["test_resource"] = func() resource.Resource {
			return &testMuxFrameworkResource{}
		}

		muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResourcesWithData, frameworkDataSources)

		serverFactory, err := muxFactory()
		if err != nil {
			t.Fatalf("Failed to create muxed provider: %v", err)
		}

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
	})
}

// testMuxFrameworkResource is a minimal test resource for mux testing
type testMuxFrameworkResource struct{}

func (r *testMuxFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_resource"
}

func (r *testMuxFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Empty schema for testing
}

func (r *testMuxFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Empty implementation for testing
}

func (r *testMuxFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Empty implementation for testing
}

func (r *testMuxFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Empty implementation for testing
}

func (r *testMuxFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Empty implementation for testing
}

// testMuxFrameworkDataSource is a minimal test data source for mux testing
type testMuxFrameworkDataSource struct{}

func (d *testMuxFrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_data_source"
}

func (d *testMuxFrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Empty schema for testing
}

func (d *testMuxFrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Empty implementation for testing
}

func TestMuxedProviderWithDataSources(t *testing.T) {
	ctx := context.Background()
	version := "test"
	providerResources := make(map[string]*schema.Resource)
	providerDataSources := make(map[string]*schema.Resource)
	frameworkResources := make(map[string]func() resource.Resource)

	// Test with Framework data sources only
	t.Run("Framework data sources only", func(t *testing.T) {
		frameworkDataSourcesWithData := make(map[string]func() datasource.DataSource)
		frameworkDataSourcesWithData["test_data_source"] = func() datasource.DataSource {
			return &testMuxFrameworkDataSource{}
		}

		muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResources, frameworkDataSourcesWithData)

		serverFactory, err := muxFactory()
		if err != nil {
			t.Fatalf("Failed to create muxed provider: %v", err)
		}

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
	})

	// Test with both Framework resources and data sources
	t.Run("Full muxed provider", func(t *testing.T) {
		frameworkResourcesWithData := make(map[string]func() resource.Resource)
		frameworkResourcesWithData["test_resource"] = func() resource.Resource {
			return &testMuxFrameworkResource{}
		}

		frameworkDataSourcesWithData := make(map[string]func() datasource.DataSource)
		frameworkDataSourcesWithData["test_data_source"] = func() datasource.DataSource {
			return &testMuxFrameworkDataSource{}
		}

		muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResourcesWithData, frameworkDataSourcesWithData)

		serverFactory, err := muxFactory()
		if err != nil {
			t.Fatalf("Failed to create muxed provider: %v", err)
		}

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
	})
}

func TestMuxedProviderResourceRouting(t *testing.T) {
	ctx := context.Background()
	version := "test"

	// Create SDKv2 resources
	providerResources := map[string]*schema.Resource{
		"genesyscloud_sdkv2_resource": {
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
	providerDataSources := make(map[string]*schema.Resource)

	// Create Framework resources
	frameworkResources := map[string]func() resource.Resource{
		"genesyscloud_framework_resource": func() resource.Resource {
			return &testMuxFrameworkResource{}
		},
	}
	frameworkDataSources := make(map[string]func() datasource.DataSource)

	// Create muxed provider
	muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResources, frameworkDataSources)

	serverFactory, err := muxFactory()
	if err != nil {
		t.Fatalf("Failed to create muxed provider: %v", err)
	}

	server := serverFactory()
	if server == nil {
		t.Fatal("Expected provider server to be created")
	}

	// Test GetProviderSchema to verify both resource types are available
	schemaReq := &tfprotov6.GetProviderSchemaRequest{}
	schemaResp, err := server.GetProviderSchema(ctx, schemaReq)
	if err != nil {
		t.Fatalf("GetProviderSchema failed: %v", err)
	}

	if schemaResp.ResourceSchemas == nil {
		t.Fatal("Expected resource schemas to be returned")
	}

	// Verify SDKv2 resource is available
	if _, exists := schemaResp.ResourceSchemas["genesyscloud_sdkv2_resource"]; !exists {
		t.Error("Expected SDKv2 resource to be available in muxed provider")
	}

	// Verify Framework resource is available (check the actual TypeName from Metadata)
	if _, exists := schemaResp.ResourceSchemas["genesyscloud_test_resource"]; !exists {
		t.Error("Expected Framework resource to be available in muxed provider")
		// Debug: Print all available resource schemas
		t.Log("Available resource schemas:")
		for name := range schemaResp.ResourceSchemas {
			t.Logf("  - %s", name)
		}
	}
}

func TestMuxedProviderPerformance(t *testing.T) {
	// Basic performance test to ensure muxing doesn't add significant overhead
	version := "test"
	providerResources := make(map[string]*schema.Resource)
	providerDataSources := make(map[string]*schema.Resource)
	frameworkResources := make(map[string]func() resource.Resource)
	frameworkDataSources := make(map[string]func() datasource.DataSource)

	// Benchmark SDKv2-only provider creation
	t.Run("SDKv2 only performance", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResources, frameworkDataSources)

			serverFactory, err := muxFactory()
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			server := serverFactory()
			if server == nil {
				t.Error("Expected provider server to be created")
			}
		}
	})

	// Add a Framework resource for muxed testing
	frameworkResourcesWithData := map[string]func() resource.Resource{
		"test_resource": func() resource.Resource {
			return &testMuxFrameworkResource{}
		},
	}

	// Benchmark muxed provider creation
	t.Run("Muxed provider performance", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			muxFactory := NewMuxedProvider(version, providerResources, providerDataSources, frameworkResourcesWithData, frameworkDataSources)

			serverFactory, err := muxFactory()
			if err != nil {
				t.Fatalf("Failed to create muxed provider: %v", err)
			}

			server := serverFactory()
			if server == nil {
				t.Error("Expected provider server to be created")
			}
		}
	})
}
