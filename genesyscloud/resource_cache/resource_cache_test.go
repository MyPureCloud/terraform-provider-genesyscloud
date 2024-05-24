package resource_cache

import (
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"testing"
)

func TestUnitWithoutExporterState(t *testing.T) {
	cache := NewResourceCache[int]()
	// Test SetCache
	SetCache(cache, "key1", 10)

	// Test GetCacheItem
	valPtr := GetCacheItem(cache, "key1")
	if valPtr != nil {
		t.Errorf("Expected Nil Value for key 'key1', got %v", valPtr)
	}

	// Test GetCacheItem for non-existent key
	valPtr = GetCacheItem(cache, "nonexistent")
	if valPtr != nil {
		t.Errorf("Expected nil value from the Cache")
	}
}

func TestUnitSetCacheAndGetCache(t *testing.T) {
	tfexporter_state.ActivateExporterState()
	cache := NewResourceCache[int]()
	// Test SetCache
	SetCache(cache, "key1", 10)

	// Test GetCacheItem
	valPtr := GetCacheItem(cache, "key1")
	if *valPtr != 10 {
		t.Errorf("Expected value %d for key 'key1', got %v", 10, valPtr)
	}

	// Test GetCacheItem for non-existent key
	valPtr = GetCacheItem(cache, "nonexistent")
	if &valPtr == nil {
		t.Errorf("Expected key 'nonexistent' to not exist in the cache")
	}
}
