package provider

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

func TestFrameworkProviderMeta(t *testing.T) {
	version := "test"
	config := platformclientv2.GetDefaultConfiguration()
	domain := "test.domain.com"

	// Test creating new meta when shared is not available
	meta := FrameworkProviderMeta(version, config, domain)
	if meta == nil {
		t.Error("Expected provider meta to be created")
	}

	if meta.Version != version {
		t.Errorf("Expected version '%s', got '%s'", version, meta.Version)
	}

	if meta.ClientConfig != config {
		t.Error("Expected client config to match")
	}

	if meta.Domain != domain {
		t.Errorf("Expected domain '%s', got '%s'", domain, meta.Domain)
	}
}

func TestSharedProviderMeta(t *testing.T) {
	// Test setting and getting shared provider meta
	originalMeta := &ProviderMeta{
		Version: "shared_test",
		Domain:  "shared.domain.com",
	}

	SetSharedProviderMeta(originalMeta)

	retrievedMeta := GetSharedProviderMeta()
	if retrievedMeta == nil {
		t.Error("Expected shared provider meta to be retrieved")
	}

	if retrievedMeta.Version != originalMeta.Version {
		t.Errorf("Expected version '%s', got '%s'", originalMeta.Version, retrievedMeta.Version)
	}

	if retrievedMeta.Domain != originalMeta.Domain {
		t.Errorf("Expected domain '%s', got '%s'", originalMeta.Domain, retrievedMeta.Domain)
	}

	// Test IsSharedMetaAvailable
	if !IsSharedMetaAvailable() {
		t.Error("Expected shared meta to be available")
	}

	// Test GetSharedClientConfig
	config := GetSharedClientConfig()
	if config != originalMeta.ClientConfig {
		t.Error("Expected shared client config to match")
	}

	// Test FrameworkProviderMeta with shared meta available
	newVersion := "new_test"
	newConfig := platformclientv2.GetDefaultConfiguration()
	newDomain := "new.domain.com"

	sharedMeta := FrameworkProviderMeta(newVersion, newConfig, newDomain)
	if sharedMeta != originalMeta {
		t.Error("Expected to use shared meta when available")
	}

	// Clean up - reset shared meta
	SetSharedProviderMeta(nil)
	if IsSharedMetaAvailable() {
		t.Error("Expected shared meta to be cleared")
	}
}

func TestSharedProviderMetaConcurrency(t *testing.T) {
	// Test concurrent access to shared provider meta
	done := make(chan bool, 10)

	// Start multiple goroutines that set and get shared meta
	for i := 0; i < 10; i++ {
		go func(id int) {
			meta := &ProviderMeta{
				Version: "concurrent_test",
				Domain:  "concurrent.domain.com",
			}

			SetSharedProviderMeta(meta)
			retrievedMeta := GetSharedProviderMeta()

			if retrievedMeta == nil {
				t.Errorf("Goroutine %d: Expected shared provider meta to be retrieved", id)
			}

			if IsSharedMetaAvailable() != true {
				t.Errorf("Goroutine %d: Expected shared meta to be available", id)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Clean up
	SetSharedProviderMeta(nil)
}

func TestFrameworkUtilityFunctions(t *testing.T) {
	// Test utility functions that might be added to framework_utils.go

	// Test provider meta creation with nil config
	meta := FrameworkProviderMeta("test", nil, "test.domain.com")
	if meta == nil {
		t.Error("Expected provider meta to be created even with nil config")
	}

	if meta.ClientConfig != nil {
		t.Error("Expected client config to be nil when passed nil")
	}

	// Test with empty values
	emptyMeta := FrameworkProviderMeta("", nil, "")
	if emptyMeta == nil {
		t.Error("Expected provider meta to be created with empty values")
	}

	if emptyMeta.Version != "" {
		t.Error("Expected empty version to be preserved")
	}

	if emptyMeta.Domain != "" {
		t.Error("Expected empty domain to be preserved")
	}
}
