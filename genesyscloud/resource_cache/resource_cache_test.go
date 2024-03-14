package resource_cache

import (
	"os"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"testing"
)

func TestUnitWithoutExporterState(t *testing.T) {
	os.Setenv("ENABLE_CX_CACHE", "")
	tfexporter_state.ActivateExporterState()
	cache := NewResourceCache[int]()
	// Test SetCache
	SetCache(cache, "key1", 10)

	// Test GetCache
	valPtr := GetCache(cache, "key1")
	if valPtr != nil {
		t.Errorf("Expected Nil Value for key 'key1', got %v", valPtr)
	}

	// Test GetCache for non-existent key
	valPtr = GetCache(cache, "nonexistent")
	if valPtr != nil {
		t.Errorf("Expected nil value from the Cache")
	}
}

func TestUnitSetCacheAndGetCache(t *testing.T) {
	os.Setenv("ENABLE_CX_CACHE", "true")
	tfexporter_state.ActivateExporterState()
	cache := NewResourceCache[int]()
	// Test SetCache
	SetCache(cache, "key1", 10)

	// Test GetCache
	valPtr := GetCache(cache, "key1")
	if *valPtr != 10 {
		t.Errorf("Expected value %d for key 'key1', got %v", 10, valPtr)
	}

	// Test GetCache for non-existent key
	valPtr = GetCache(cache, "nonexistent")
	if &valPtr == nil {
		t.Errorf("Expected key 'nonexistent' to not exist in the cache")
	}
}
