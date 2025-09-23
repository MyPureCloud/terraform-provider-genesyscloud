// Package resource_register provides interfaces and utilities for managing provider resources
// and data sources in a muxed provider environment.
//
// This package defines the core interfaces and types used for:
//   - Registering SDKv2 and Framework resources/data sources
//   - Tracking provider types for proper routing
//   - Managing resource exporters
//   - Providing backward compatibility during migration
//
// The package supports both SDKv2 (legacy) and Plugin Framework (modern) provider
// architectures, allowing for gradual migration while maintaining full functionality.
package resource_register

import (
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProviderType represents the type of provider implementation for resources and data sources.
// This enum is used to track which provider architecture (SDKv2 or Framework) should handle
// specific resources, enabling proper routing in the muxed provider environment.
type ProviderType int

const (
	// SDKv2Provider indicates the resource/data source uses the legacy SDKv2 architecture
	SDKv2Provider ProviderType = iota

	// FrameworkProvider indicates the resource/data source uses the modern Plugin Framework architecture
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

// Registrar interface defines the contract for registering provider resources, data sources,
// and exporters in both SDKv2 and Framework provider architectures.
//
// This interface supports the muxed provider architecture by:
//   - Maintaining separate registration methods for SDKv2 and Framework components
//   - Tracking provider types for proper routing
//   - Supporting resource exporters for Terraform state management
//
// Implementation note: The registrar must be thread-safe as it may be called
// concurrently during provider initialization.
type Registrar interface {
	// SDKv2 provider registration methods (legacy)
	RegisterResource(resourceType string, resource *schema.Resource)
	RegisterDataSource(dataSourceType string, datasource *schema.Resource)
	RegisterExporter(exporterResourceType string, resourceExporter *resourceExporter.ResourceExporter)

	// Framework provider registration methods (modern)
	RegisterFrameworkResource(resourceType string, resourceFactory func() resource.Resource)
	RegisterFrameworkDataSource(dataSourceType string, dataSourceFactory func() datasource.DataSource)

	// Provider type tracking for muxer routing
	GetResourceProviderType(resourceType string) ProviderType
	GetDataSourceProviderType(dataSourceType string) ProviderType
}

// Package-level variables for managing provider resources and data sources.
// These variables are used by the TFexport functionality and provider initialization.

// SDKv2 provider resources and data sources (legacy architecture)
var providerResources map[string]*schema.Resource   // Maps resource names to SDKv2 resource definitions
var providerDataSources map[string]*schema.Resource // Maps data source names to SDKv2 data source definitions

// Framework provider resources and data sources (modern architecture)
var frameworkResources map[string]func() resource.Resource       // Maps resource names to Framework resource factory functions
var frameworkDataSources map[string]func() datasource.DataSource // Maps data source names to Framework data source factory functions

// Provider type tracking for muxer routing
// These maps track which provider architecture should handle each resource/data source
var resourceProviderTypes map[string]ProviderType   // Maps resource names to their provider type
var dataSourceProviderTypes map[string]ProviderType // Maps data source names to their provider type

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

// GetResourceProviderType returns the provider type for a given resource type.
// This function is used by the muxer to determine which provider (SDKv2 or Framework)
// should handle a specific resource.
//
// Parameters:
//   - resourceType: The Terraform resource type name (e.g., "genesyscloud_routing_language")
//
// Returns:
//   - ProviderType: The provider type that should handle this resource
//
// Default behavior: Returns SDKv2Provider for unknown resource types to maintain
// backward compatibility during the migration process.
func GetResourceProviderType(resourceType string) ProviderType {
	if providerType, exists := resourceProviderTypes[resourceType]; exists {
		return providerType
	}
	return SDKv2Provider // Default to SDKv2 for backward compatibility
}

// GetDataSourceProviderType returns the provider type for a given data source type.
// This function is used by the muxer to determine which provider (SDKv2 or Framework)
// should handle a specific data source.
//
// Parameters:
//   - dataSourceType: The Terraform data source type name (e.g., "genesyscloud_routing_language")
//
// Returns:
//   - ProviderType: The provider type that should handle this data source
//
// Default behavior: Returns SDKv2Provider for unknown data source types to maintain
// backward compatibility during the migration process.
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
