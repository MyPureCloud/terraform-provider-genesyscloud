package routing_language

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

/*
The genesyscloud_routing_language_init_test.go file is used to initialize the data sources and resources
used in testing the routing_language resource.
*/

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() resource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests (Framework-only)
func (r *registerTestInstance) registerTestResources() {
	// SDKv2 resources removed - Framework-only migration
}

// registerTestDataSources registers all data sources used in the tests (Framework-only)
func (r *registerTestInstance) registerTestDataSources() {
	// SDKv2 data sources removed - Framework-only migration
}

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[ResourceType] = NewFrameworkRoutingLanguageResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[ResourceType] = NewFrameworkRoutingLanguageDataSource
}

// initTestResources initializes all test resources and data sources (Framework-only).
func initTestResources() {
	// Framework-only initialization
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	// Framework resources only
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for routing_language package
	initTestResources()

	// Run the test suite for the routing_language package
	m.Run()
}
