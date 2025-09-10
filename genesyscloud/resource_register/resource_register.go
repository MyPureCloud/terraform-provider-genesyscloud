package resource_register

import (
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProviderType represents the type of provider (SDKv2 or Framework)
type ProviderType int

const (
	SDKv2Provider ProviderType = iota
	FrameworkProvider
)

// String returns a string representation of the ProviderType
func (pt ProviderType) String() string {
	switch pt {
	case SDKv2Provider:
		return "SDKv2"
	case FrameworkProvider:
		return "Framework"
	default:
		return "Unknown"
	}
}

type Registrar interface {
	// Existing SDKv2 methods
	RegisterResource(resourceType string, resource *schema.Resource)
	RegisterDataSource(dataSourceType string, datasource *schema.Resource)
	RegisterExporter(exporterResourceType string, resourceExporter *resourceExporter.ResourceExporter)

	// NEW: Framework methods
	RegisterFrameworkResource(resourceType string, resourceFactory func() resource.Resource)
	RegisterFrameworkDataSource(dataSourceType string, dataSourceFactory func() datasource.DataSource)

	// NEW: Provider type tracking
	GetResourceProviderType(resourceType string) ProviderType
	GetDataSourceProviderType(dataSourceType string) ProviderType
}

// need this for TFexport where Resources are required for provider initialisation.
// NewGenesysCloudResourceExporter

// SDKv2 provider resources and data sources
var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

// Framework provider resources and data sources
var frameworkResources map[string]func() resource.Resource
var frameworkDataSources map[string]func() datasource.DataSource

// Provider type tracking
var resourceProviderTypes map[string]ProviderType
var dataSourceProviderTypes map[string]ProviderType

func SetResources(resources map[string]*schema.Resource, dataSources map[string]*schema.Resource) {
	providerResources = resources
	providerDataSources = dataSources
}

func GetResources() (map[string]*schema.Resource, map[string]*schema.Resource) {
	return providerResources, providerDataSources
}

// GetFrameworkResources returns the Framework resources and data sources
func GetFrameworkResources() (map[string]func() resource.Resource, map[string]func() datasource.DataSource) {
	return frameworkResources, frameworkDataSources
}

// GetResourceProviderType returns the provider type for a given resource type
func GetResourceProviderType(resourceType string) ProviderType {
	if providerType, exists := resourceProviderTypes[resourceType]; exists {
		return providerType
	}
	return SDKv2Provider // Default to SDKv2 for backward compatibility
}

// GetDataSourceProviderType returns the provider type for a given data source type
func GetDataSourceProviderType(dataSourceType string) ProviderType {
	if providerType, exists := dataSourceProviderTypes[dataSourceType]; exists {
		return providerType
	}
	return SDKv2Provider // Default to SDKv2 for backward compatibility
}

// SetFrameworkResources sets the Framework resources and data sources (for testing/export)
func SetFrameworkResources(resources map[string]func() resource.Resource, dataSources map[string]func() datasource.DataSource) {
	frameworkResources = resources
	frameworkDataSources = dataSources
}
