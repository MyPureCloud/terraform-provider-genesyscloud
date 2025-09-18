package routing_wrapupcode

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_routing_wrapupcode_init_test.go file is used to initialize the data sources and resources
used in testing the routing_wrapupcode resource.
*/

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() resource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

// Temporary variables for backward compatibility with old SDKv2 tests (will be removed in task 9)
var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

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

	frameworkResources[ResourceType] = NewRoutingWrapupcodeFrameworkResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[ResourceType] = NewRoutingWrapupcodeFrameworkDataSource
}

// initTestResources initializes all test resources and data sources (Framework-only).
func initTestResources() {
	// Framework-only initialization
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	// Temporary initialization for backward compatibility with old SDKv2 tests (will be removed in task 9)
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	// Framework resources only
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the routing_wrapupcode package
	initTestResources()

	// Run the test suite for the routing_wrapupcode package
	m.Run()
}
