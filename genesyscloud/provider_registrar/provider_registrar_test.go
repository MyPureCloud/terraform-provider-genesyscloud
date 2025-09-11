package provider_registrar

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

// Mock Framework resource for testing
type mockFrameworkResource struct{}

func (m *mockFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_framework_resource"
}

func (m *mockFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Empty schema for testing
}

func (m *mockFrameworkResource) Create(_ context.Context, _ resource.CreateRequest, _ *resource.CreateResponse) {
	// Empty implementation for testing
}

func (m *mockFrameworkResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// Empty implementation for testing
}

func (m *mockFrameworkResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Empty implementation for testing
}

func (m *mockFrameworkResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Empty implementation for testing
}

// Mock Framework data source for testing
type mockFrameworkDataSource struct{}

func (m *mockFrameworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_framework_datasource"
}

func (m *mockFrameworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Empty schema for testing
}

func (m *mockFrameworkDataSource) Read(_ context.Context, _ datasource.ReadRequest, _ *datasource.ReadResponse) {
	// Empty implementation for testing
}

func TestRegisterFramework(t *testing.T) {
	// Create a new RegisterInstance for testing
	regInstance := &RegisterInstance{}

	// Test Framework resource registration
	t.Run("RegisterFrameworkResource", func(t *testing.T) {
		resourceType := "genesyscloud_test_framework_resource"
		resourceFactory := func() resource.Resource {
			return &mockFrameworkResource{}
		}

		// Register the Framework resource
		regInstance.RegisterFrameworkResource(resourceType, resourceFactory)

		// Verify the resource was registered
		resources, _ := GetFrameworkResources()
		if _, exists := resources[resourceType]; !exists {
			t.Errorf("Framework resource %s was not registered", resourceType)
		}

		// Verify provider type tracking
		providerType := GetResourceProviderType(resourceType)
		if providerType != registrar.FrameworkProvider {
			t.Errorf("Expected provider type %v, got %v", registrar.FrameworkProvider, providerType)
		}
	})

	// Test Framework data source registration
	t.Run("RegisterFrameworkDataSource", func(t *testing.T) {
		dataSourceType := "genesyscloud_test_framework_datasource"
		dataSourceFactory := func() datasource.DataSource {
			return &mockFrameworkDataSource{}
		}

		// Register the Framework data source
		regInstance.RegisterFrameworkDataSource(dataSourceType, dataSourceFactory)

		// Verify the data source was registered
		_, dataSources := GetFrameworkResources()
		if _, exists := dataSources[dataSourceType]; !exists {
			t.Errorf("Framework data source %s was not registered", dataSourceType)
		}

		// Verify provider type tracking
		providerType := GetDataSourceProviderType(dataSourceType)
		if providerType != registrar.FrameworkProvider {
			t.Errorf("Expected provider type %v, got %v", registrar.FrameworkProvider, providerType)
		}
	})

	// Test SDKv2 resource registration (existing functionality)
	t.Run("RegisterSDKv2Resource", func(t *testing.T) {
		resourceType := "genesyscloud_test_sdkv2_resource"
		resource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"test_field": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		}

		// Register the SDKv2 resource
		regInstance.RegisterResource(resourceType, resource)

		// Verify the resource was registered
		resources, _ := GetProviderResources()
		if _, exists := resources[resourceType]; !exists {
			t.Errorf("SDKv2 resource %s was not registered", resourceType)
		}

		// Verify provider type tracking (should default to SDKv2)
		providerType := GetResourceProviderType(resourceType)
		if providerType != registrar.SDKv2Provider {
			t.Errorf("Expected provider type %v, got %v", registrar.SDKv2Provider, providerType)
		}
	})

	// Test SDKv2 data source registration (existing functionality)
	t.Run("RegisterSDKv2DataSource", func(t *testing.T) {
		dataSourceType := "genesyscloud_test_sdkv2_datasource"
		dataSource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"test_field": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		}

		// Register the SDKv2 data source
		regInstance.RegisterDataSource(dataSourceType, dataSource)

		// Verify the data source was registered
		_, dataSources := GetProviderResources()
		if _, exists := dataSources[dataSourceType]; !exists {
			t.Errorf("SDKv2 data source %s was not registered", dataSourceType)
		}

		// Verify provider type tracking (should default to SDKv2)
		providerType := GetDataSourceProviderType(dataSourceType)
		if providerType != registrar.SDKv2Provider {
			t.Errorf("Expected provider type %v, got %v", registrar.SDKv2Provider, providerType)
		}
	})

	// Test exporter registration
	t.Run("RegisterExporter", func(t *testing.T) {
		exporterResourceType := "genesyscloud_test_exporter"
		exporter := &resourceExporter.ResourceExporter{
			GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				return resourceExporter.ResourceIDMetaMap{
					"test-resource-1": &resourceExporter.ResourceMeta{BlockLabel: "test-resource-1"},
					"test-resource-2": &resourceExporter.ResourceMeta{BlockLabel: "test-resource-2"},
				}, nil
			},
		}

		// Register the exporter
		regInstance.RegisterExporter(exporterResourceType, exporter)

		// Verify the exporter was registered
		exporters := GetResourceExporters()
		if _, exists := exporters[exporterResourceType]; !exists {
			t.Errorf("Exporter %s was not registered", exporterResourceType)
		}
	})

	// Test provider type separation
	t.Run("ProviderTypeSeparation", func(t *testing.T) {
		sdkv2Resources, frameworkResourceFactories, sdkv2DataSources, frameworkDataSourceFactories := GetAllResourcesByProvider()

		// Verify we have both types of resources
		if len(sdkv2Resources) == 0 {
			t.Error("Expected SDKv2 resources to be registered")
		}

		if len(frameworkResourceFactories) == 0 {
			t.Error("Expected Framework resources to be registered")
		}

		if len(sdkv2DataSources) == 0 {
			t.Error("Expected SDKv2 data sources to be registered")
		}

		if len(frameworkDataSourceFactories) == 0 {
			t.Error("Expected Framework data sources to be registered")
		}
	})
}

// Test concurrent registration (thread safety)
func TestRegisterFrameworkConcurrency(t *testing.T) {
	regInstance := &RegisterInstance{}

	// Test concurrent Framework resource registration
	t.Run("ConcurrentFrameworkResourceRegistration", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				resourceType := fmt.Sprintf("genesyscloud_concurrent_test_%d", index)
				resourceFactory := func() resource.Resource {
					return &mockFrameworkResource{}
				}
				regInstance.RegisterFrameworkResource(resourceType, resourceFactory)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all resources were registered
		resources, _ := GetFrameworkResources()
		if len(resources) < 10 {
			t.Errorf("Expected at least 10 Framework resources, got %d", len(resources))
		}
	})
}
