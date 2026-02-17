package resource_register

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// Mock Framework resource for testing
type mockFrameworkResource struct{}

func (m *mockFrameworkResource) Metadata(_ context.Context, req fwresource.MetadataRequest, resp *fwresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_framework_resource"
}

func (m *mockFrameworkResource) Schema(_ context.Context, _ fwresource.SchemaRequest, resp *fwresource.SchemaResponse) {
	// Empty schema for testing
}

func (m *mockFrameworkResource) Create(_ context.Context, _ fwresource.CreateRequest, _ *fwresource.CreateResponse) {
	// Empty implementation for testing
}

func (m *mockFrameworkResource) Read(_ context.Context, _ fwresource.ReadRequest, _ *fwresource.ReadResponse) {
	// Empty implementation for testing
}

func (m *mockFrameworkResource) Update(_ context.Context, _ fwresource.UpdateRequest, _ *fwresource.UpdateResponse) {
	// Empty implementation for testing
}

func (m *mockFrameworkResource) Delete(_ context.Context, _ fwresource.DeleteRequest, _ *fwresource.DeleteResponse) {
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

// Mock registrar implementation for testing
type mockRegistrar struct {
	resources               map[string]*schema.Resource
	dataSources             map[string]*schema.Resource
	exporters               map[string]*resourceExporter.ResourceExporter
	frameworkResources      map[string]func() fwresource.Resource
	frameworkDataSources    map[string]func() datasource.DataSource
	resourceProviderTypes   map[string]ProviderType
	dataSourceProviderTypes map[string]ProviderType
}

func newMockRegistrar() *mockRegistrar {
	return &mockRegistrar{
		resources:               make(map[string]*schema.Resource),
		dataSources:             make(map[string]*schema.Resource),
		exporters:               make(map[string]*resourceExporter.ResourceExporter),
		frameworkResources:      make(map[string]func() fwresource.Resource),
		frameworkDataSources:    make(map[string]func() datasource.DataSource),
		resourceProviderTypes:   make(map[string]ProviderType),
		dataSourceProviderTypes: make(map[string]ProviderType),
	}
}

func (m *mockRegistrar) RegisterResource(resourceType string, resource *schema.Resource) {
	m.resources[resourceType] = resource
	m.resourceProviderTypes[resourceType] = SDKv2Provider
}

func (m *mockRegistrar) RegisterDataSource(dataSourceType string, datasource *schema.Resource) {
	m.dataSources[dataSourceType] = datasource
	m.dataSourceProviderTypes[dataSourceType] = SDKv2Provider
}

func (m *mockRegistrar) RegisterExporter(exporterResourceType string, resourceExporter *resourceExporter.ResourceExporter) {
	m.exporters[exporterResourceType] = resourceExporter
}

func (m *mockRegistrar) RegisterFrameworkResource(resourceType string, resourceFactory func() fwresource.Resource) {
	m.frameworkResources[resourceType] = resourceFactory
	m.resourceProviderTypes[resourceType] = FrameworkProvider
}

func (m *mockRegistrar) RegisterFrameworkDataSource(dataSourceType string, dataSourceFactory func() datasource.DataSource) {
	m.frameworkDataSources[dataSourceType] = dataSourceFactory
	m.dataSourceProviderTypes[dataSourceType] = FrameworkProvider
}

func (m *mockRegistrar) GetResourceProviderType(resourceType string) ProviderType {
	if providerType, exists := m.resourceProviderTypes[resourceType]; exists {
		return providerType
	}
	return SDKv2Provider
}

func (m *mockRegistrar) GetDataSourceProviderType(dataSourceType string) ProviderType {
	if providerType, exists := m.dataSourceProviderTypes[dataSourceType]; exists {
		return providerType
	}
	return SDKv2Provider
}

func TestFramework(t *testing.T) {
	// Test ProviderType enum
	t.Run("ProviderTypeString", func(t *testing.T) {
		if SDKv2Provider.String() != "SDKv2" {
			t.Errorf("Expected 'SDKv2', got '%s'", SDKv2Provider.String())
		}
		if FrameworkProvider.String() != "Framework" {
			t.Errorf("Expected 'Framework', got '%s'", FrameworkProvider.String())
		}
	})

	// Test Registrar interface implementation
	t.Run("RegistrarInterface", func(t *testing.T) {
		registrar := newMockRegistrar()

		// Test SDKv2 resource registration
		resourceType := "genesyscloud_test_resource"
		resource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"test_field": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		}
		registrar.RegisterResource(resourceType, resource)

		if _, exists := registrar.resources[resourceType]; !exists {
			t.Errorf("Resource %s was not registered", resourceType)
		}

		if registrar.GetResourceProviderType(resourceType) != SDKv2Provider {
			t.Errorf("Expected SDKv2Provider, got %v", registrar.GetResourceProviderType(resourceType))
		}

		// Test SDKv2 data source registration
		dataSourceType := "genesyscloud_test_datasource"
		dataSource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"test_field": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		}
		registrar.RegisterDataSource(dataSourceType, dataSource)

		if _, exists := registrar.dataSources[dataSourceType]; !exists {
			t.Errorf("Data source %s was not registered", dataSourceType)
		}

		if registrar.GetDataSourceProviderType(dataSourceType) != SDKv2Provider {
			t.Errorf("Expected SDKv2Provider, got %v", registrar.GetDataSourceProviderType(dataSourceType))
		}

		// Test Framework resource registration
		frameworkResourceType := "genesyscloud_test_framework_resource"
		frameworkResourceFactory := func() fwresource.Resource {
			return &mockFrameworkResource{}
		}
		registrar.RegisterFrameworkResource(frameworkResourceType, frameworkResourceFactory)

		if _, exists := registrar.frameworkResources[frameworkResourceType]; !exists {
			t.Errorf("Framework resource %s was not registered", frameworkResourceType)
		}

		if registrar.GetResourceProviderType(frameworkResourceType) != FrameworkProvider {
			t.Errorf("Expected FrameworkProvider, got %v", registrar.GetResourceProviderType(frameworkResourceType))
		}

		// Test Framework data source registration
		frameworkDataSourceType := "genesyscloud_test_framework_datasource"
		frameworkDataSourceFactory := func() datasource.DataSource {
			return &mockFrameworkDataSource{}
		}
		registrar.RegisterFrameworkDataSource(frameworkDataSourceType, frameworkDataSourceFactory)

		if _, exists := registrar.frameworkDataSources[frameworkDataSourceType]; !exists {
			t.Errorf("Framework data source %s was not registered", frameworkDataSourceType)
		}

		if registrar.GetDataSourceProviderType(frameworkDataSourceType) != FrameworkProvider {
			t.Errorf("Expected FrameworkProvider, got %v", registrar.GetDataSourceProviderType(frameworkDataSourceType))
		}
	})

	// Test resource and data source management functions
	t.Run("ResourceManagement", func(t *testing.T) {
		// Test SetResources and GetResources
		testResources := map[string]*schema.Resource{
			"test_resource": {
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		}
		testDataSources := map[string]*schema.Resource{
			"test_datasource": {
				Schema: map[string]*schema.Schema{
					"filter": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		}

		SetResources(testResources, testDataSources)
		resources, dataSources := GetResources()

		if len(resources) != len(testResources) {
			t.Errorf("Expected %d resources, got %d", len(testResources), len(resources))
		}

		if len(dataSources) != len(testDataSources) {
			t.Errorf("Expected %d data sources, got %d", len(testDataSources), len(dataSources))
		}

		// Test SetFrameworkResources and GetFrameworkResources
		testFrameworkResources := map[string]func() fwresource.Resource{
			"test_framework_resource": func() fwresource.Resource {
				return &mockFrameworkResource{}
			},
		}
		testFrameworkDataSources := map[string]func() datasource.DataSource{
			"test_framework_datasource": func() datasource.DataSource {
				return &mockFrameworkDataSource{}
			},
		}

		SetFrameworkResources(testFrameworkResources, testFrameworkDataSources)
		frameworkResources, frameworkDataSources := GetFrameworkResources()

		if len(frameworkResources) != len(testFrameworkResources) {
			t.Errorf("Expected %d framework resources, got %d", len(testFrameworkResources), len(frameworkResources))
		}

		if len(frameworkDataSources) != len(testFrameworkDataSources) {
			t.Errorf("Expected %d framework data sources, got %d", len(testFrameworkDataSources), len(frameworkDataSources))
		}
	})

	// Test provider type functions with default behavior
	t.Run("ProviderTypeDefaults", func(t *testing.T) {
		// Test that unknown resource types default to SDKv2Provider
		unknownResourceType := "unknown_resource_type"
		providerType := GetResourceProviderType(unknownResourceType)
		if providerType != SDKv2Provider {
			t.Errorf("Expected SDKv2Provider for unknown resource type, got %v", providerType)
		}

		// Test that unknown data source types default to SDKv2Provider
		unknownDataSourceType := "unknown_datasource_type"
		providerType = GetDataSourceProviderType(unknownDataSourceType)
		if providerType != SDKv2Provider {
			t.Errorf("Expected SDKv2Provider for unknown data source type, got %v", providerType)
		}
	})
}

// Test Framework resource and data source functionality
func TestFrameworkResourcesAndDataSources(t *testing.T) {
	t.Run("FrameworkResourceCreation", func(t *testing.T) {
		resourceFactory := func() fwresource.Resource {
			return &mockFrameworkResource{}
		}

		// Create resource instance
		resourceInstance := resourceFactory()
		if resourceInstance == nil {
			t.Error("Framework resource factory returned nil")
		}

		// Test that we can call methods on the resource
		ctx := context.Background()
		metadataReq := fwresource.MetadataRequest{ProviderTypeName: "test"}
		metadataResp := &fwresource.MetadataResponse{}
		resourceInstance.Metadata(ctx, metadataReq, metadataResp)

		expectedTypeName := "test_test_framework_resource"
		if metadataResp.TypeName != expectedTypeName {
			t.Errorf("Expected TypeName '%s', got '%s'", expectedTypeName, metadataResp.TypeName)
		}
	})

	t.Run("FrameworkDataSourceCreation", func(t *testing.T) {
		dataSourceFactory := func() datasource.DataSource {
			return &mockFrameworkDataSource{}
		}

		// Create data source instance
		dataSourceInstance := dataSourceFactory()
		if dataSourceInstance == nil {
			t.Error("Framework data source factory returned nil")
		}

		// Test that we can call methods on the data source
		ctx := context.Background()
		metadataReq := datasource.MetadataRequest{ProviderTypeName: "test"}
		metadataResp := &datasource.MetadataResponse{}
		dataSourceInstance.Metadata(ctx, metadataReq, metadataResp)

		expectedTypeName := "test_test_framework_datasource"
		if metadataResp.TypeName != expectedTypeName {
			t.Errorf("Expected TypeName '%s', got '%s'", expectedTypeName, metadataResp.TypeName)
		}
	})
}
