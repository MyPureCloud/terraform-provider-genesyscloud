// Package provider implements shared provider metadata management for muxed providers.
//
// This file manages the sharing of provider metadata between SDKv2 and Framework providers
// in a muxed environment. This sharing is crucial for:
//   - Avoiding duplicate authentication
//   - Sharing API client configurations
//   - Maintaining consistent provider state
//   - Optimizing resource usage
//
// The shared metadata system uses thread-safe operations to ensure data consistency
// when both providers are accessing the same metadata concurrently.
package provider

import (
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

// SharedProviderMeta provides thread-safe access to provider metadata that can be
// shared between SDKv2 and Framework providers in a muxed environment.
//
// This struct uses a read-write mutex to ensure thread-safe access to the shared
// provider metadata. Multiple readers can access the metadata concurrently, but
// writers have exclusive access.
//
// Fields:
//   - mutex: RWMutex for thread-safe access control
//   - meta: Shared provider metadata containing configuration and client information
type SharedProviderMeta struct {
	mutex sync.RWMutex
	meta  *ProviderMeta
}

var sharedMeta *SharedProviderMeta

func init() {
	sharedMeta = &SharedProviderMeta{}
}

// SetSharedProviderMeta sets the shared provider metadata in a thread-safe manner.
// This is typically called by the SDKv2 provider during its configuration phase
// to make its metadata available to the Framework provider.
//
// Parameters:
//   - meta: Provider metadata to be shared between providers
func SetSharedProviderMeta(meta *ProviderMeta) {
	sharedMeta.mutex.Lock()
	defer sharedMeta.mutex.Unlock()
	sharedMeta.meta = meta
}

// GetSharedProviderMeta retrieves the shared provider metadata in a thread-safe manner.
// This allows the Framework provider to access metadata that was configured by the SDKv2 provider.
//
// Returns:
//   - *ProviderMeta: Shared provider metadata, or nil if not available
func GetSharedProviderMeta() *ProviderMeta {
	sharedMeta.mutex.RLock()
	defer sharedMeta.mutex.RUnlock()
	return sharedMeta.meta
}

// GetSharedClientConfig gets the shared client configuration
func GetSharedClientConfig() *platformclientv2.Configuration {
	meta := GetSharedProviderMeta()
	if meta != nil {
		return meta.ClientConfig
	}
	return nil
}

// IsSharedMetaAvailable checks if shared metadata is available
func IsSharedMetaAvailable() bool {
	sharedMeta.mutex.RLock()
	defer sharedMeta.mutex.RUnlock()
	return sharedMeta.meta != nil
}

// FrameworkProviderMeta creates or retrieves provider metadata for the Framework provider.
// This function implements a fallback strategy:
//  1. First, try to use shared metadata from the SDKv2 provider (preferred)
//  2. If shared metadata is not available, create new metadata
//
// Using shared metadata is preferred because it:
//   - Avoids duplicate authentication
//   - Reuses existing API client configurations
//   - Maintains consistency between providers
//   - Reduces resource usage
//
// Parameters:
//   - version: Provider version string
//   - clientConfig: Genesys Cloud API client configuration
//   - domain: Genesys Cloud domain/region
//
// Returns:
//   - *ProviderMeta: Provider metadata (shared or newly created)
func FrameworkProviderMeta(version string, clientConfig *platformclientv2.Configuration, domain string) *ProviderMeta {
	// Try to use shared meta first (from SDKv2 provider)
	if sharedMeta := GetSharedProviderMeta(); sharedMeta != nil {
		return sharedMeta
	}

	// Create new meta if shared is not available
	return &ProviderMeta{
		Version:      version,
		ClientConfig: clientConfig,
		Domain:       domain,
	}
}
