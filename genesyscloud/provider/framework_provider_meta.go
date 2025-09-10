package provider

import (
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// SharedProviderMeta provides thread-safe access to provider metadata
// that can be shared between SDKv2 and Framework providers
type SharedProviderMeta struct {
	mutex sync.RWMutex
	meta  *ProviderMeta
}

var sharedMeta *SharedProviderMeta

func init() {
	sharedMeta = &SharedProviderMeta{}
}

// SetSharedProviderMeta sets the shared provider metadata
func SetSharedProviderMeta(meta *ProviderMeta) {
	sharedMeta.mutex.Lock()
	defer sharedMeta.mutex.Unlock()
	sharedMeta.meta = meta
}

// GetSharedProviderMeta gets the shared provider metadata
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

// FrameworkProviderMeta creates a provider meta for Framework provider
// This can either use shared meta from SDKv2 or create its own
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
