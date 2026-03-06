package dependent_consumers

import (
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestBuildDependsMap(t *testing.T) {
	resources := resourceExporter.ResourceIDMetaMap{
		"id1": {BlockLabel: "genesyscloud_user::::id1"},
		"id2": {BlockLabel: "genesyscloud_queue::::id2"},
		"id3": {BlockLabel: "genesyscloud_flow::::id3"},
	}
	dependsMap := make(map[string][]string)
	
	result := buildDependsMap(resources, dependsMap, "main-id")
	
	assert.NotNil(t, result)
	assert.Contains(t, result, "main-id")
	assert.Len(t, result["main-id"], 3)
	
	// Verify all expected dependencies are present
	deps := result["main-id"]
	assert.Contains(t, deps, "genesyscloud_user.id1")
	assert.Contains(t, deps, "genesyscloud_queue.id2")
	assert.Contains(t, deps, "genesyscloud_flow.id3")
}

func TestBuildDependsMap_ExcludesSelf(t *testing.T) {
	resources := resourceExporter.ResourceIDMetaMap{
		"id1": {BlockLabel: "genesyscloud_user::::id1"},
		"id2": {BlockLabel: "genesyscloud_queue::::id2"},
	}
	dependsMap := make(map[string][]string)
	
	result := buildDependsMap(resources, dependsMap, "id1")
	
	assert.Len(t, result["id1"], 1)
	assert.Equal(t, "genesyscloud_queue.id2", result["id1"][0])
}

func TestBuildDependsMap_EmptyResources(t *testing.T) {
	resources := resourceExporter.ResourceIDMetaMap{}
	dependsMap := make(map[string][]string)
	
	result := buildDependsMap(resources, dependsMap, "main-id")
	
	assert.NotNil(t, result)
	assert.Contains(t, result, "main-id")
	assert.Empty(t, result["main-id"])
}

func TestBuildDependsMap_Deterministic(t *testing.T) {
	resources := resourceExporter.ResourceIDMetaMap{
		"z-id": {BlockLabel: "genesyscloud_user::::z-id"},
		"a-id": {BlockLabel: "genesyscloud_queue::::a-id"},
		"m-id": {BlockLabel: "genesyscloud_flow::::m-id"},
	}
	
	// Run multiple times to verify consistent ordering
	results := make([][]string, 5)
	for i := 0; i < 5; i++ {
		dependsMap := make(map[string][]string)
		result := buildDependsMap(resources, dependsMap, "main-id")
		results[i] = result["main-id"]
	}
	
	// All results should be identical
	for i := 1; i < 5; i++ {
		assert.Equal(t, results[0], results[i])
	}
}

func TestGetResourceType(t *testing.T) {
	dependentConsumerMap := SetDependentObjectMaps()
	
	tests := []struct {
		name         string
		varType      string
		expectedType string
		expectedOk   bool
	}{
		{"valid queue", "QUEUE", "genesyscloud_routing_queue", true},
		{"valid user", "USER", "genesyscloud_user", true},
		{"valid flow", "BOTFLOW", "genesyscloud_flow", true},
		{"invalid type", "INVALID_TYPE", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := platformclientv2.Dependency{
				VarType: &tt.varType,
			}
			
			resourceType, ok := getResourceType(consumer, dependentConsumerMap)
			
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedType, resourceType)
		})
	}
}

func TestGetResourceFilter(t *testing.T) {
	varType := "QUEUE"
	id := "test-id-123"
	name := "Test Queue"
	
	consumer := platformclientv2.Dependency{
		VarType: &varType,
		Id:      &id,
		Name:    &name,
	}
	
	result := getResourceFilter(consumer, "genesyscloud_routing_queue")
	
	assert.Equal(t, "genesyscloud_routing_queue::::test-id-123", result)
}

func TestProcessResource(t *testing.T) {
	varType := "QUEUE"
	id := "queue-123"
	name := "Test Queue"
	
	consumer := platformclientv2.Dependency{
		VarType: &varType,
		Id:      &id,
		Name:    &name,
	}
	
	resources := make(resourceExporter.ResourceIDMetaMap)
	architectDependencies := make(map[string][]string)
	key := "flow-123"
	
	resources, architectDependencies = processResource(consumer, "genesyscloud_routing_queue", resources, architectDependencies, key)
	
	assert.Contains(t, resources, "queue-123")
	assert.Equal(t, "genesyscloud_routing_queue::::queue-123", resources["queue-123"].BlockLabel)
	assert.Contains(t, architectDependencies, key)
	assert.Contains(t, architectDependencies[key], "queue-123")
}

func TestProcessResource_DoesNotDuplicate(t *testing.T) {
	varType := "QUEUE"
	id := "queue-123"
	name := "Test Queue"
	
	consumer := platformclientv2.Dependency{
		VarType: &varType,
		Id:      &id,
		Name:    &name,
	}
	
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["queue-123"] = &resourceExporter.ResourceMeta{BlockLabel: "existing"}
	architectDependencies := make(map[string][]string)
	key := "flow-123"
	
	resources, architectDependencies = processResource(consumer, "genesyscloud_routing_queue", resources, architectDependencies, key)
	
	// Should not overwrite existing resource
	assert.Equal(t, "existing", resources["queue-123"].BlockLabel)
}

func TestIsDependencyPresent(t *testing.T) {
	architectDependencies := map[string][]string{
		"flow1": {"dep1", "dep2", "dep3"},
		"flow2": {"dep4", "dep5"},
	}
	
	assert.True(t, isDependencyPresent(architectDependencies, "flow1", "dep2"))
	assert.False(t, isDependencyPresent(architectDependencies, "flow2", "dep2"))
	assert.False(t, isDependencyPresent(architectDependencies, "flow1", "dep6"))
	assert.False(t, isDependencyPresent(architectDependencies, "flow3", "dep1"))
}

func TestSearchForKeyValue(t *testing.T) {
	m := map[string][]string{
		"key1": {"value1", "value2", "value3"},
		"key2": {"value4", "value5"},
	}
	
	assert.True(t, searchForKeyValue(m, "key1", "value2"))
	assert.False(t, searchForKeyValue(m, "key1", "value4"))
	assert.False(t, searchForKeyValue(m, "key3", "value1"))
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}
	
	assert.True(t, stringInSlice("banana", slice))
	assert.False(t, stringInSlice("grape", slice))
	assert.False(t, stringInSlice("", slice))
}

func TestGetDependentConsumerProxy_Singleton(t *testing.T) {
	config := &platformclientv2.Configuration{}
	
	proxy1 := GetDependentConsumerProxy(config)
	proxy2 := GetDependentConsumerProxy(config)
	
	// Should return same instance
	assert.Equal(t, proxy1, proxy2)
}

func TestNewDependentConsumerProxy_ThreadSafe(t *testing.T) {
	// Reset singleton for test
	InternalProxy = nil
	
	config := &platformclientv2.Configuration{}
	done := make(chan *DependentConsumerProxy, 10)
	
	// Run concurrent initializations
	for i := 0; i < 10; i++ {
		go func() {
			proxy := newDependentConsumerProxy(config)
			done <- proxy
		}()
	}
	
	// Collect all proxies
	proxies := make([]*DependentConsumerProxy, 10)
	for i := 0; i < 10; i++ {
		proxies[i] = <-done
	}
	
	// All should be the same instance
	for i := 1; i < 10; i++ {
		assert.Equal(t, proxies[0], proxies[i])
	}
}

func TestGetDependentConsumerProxy_NilConfig(t *testing.T) {
	// Reset singleton
	InternalProxy = nil
	
	proxy := GetDependentConsumerProxy(nil)
	
	assert.NotNil(t, proxy)
	assert.NotNil(t, proxy.GetPooledClientAttr)
}
