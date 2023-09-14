package integration

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	trueValue  = "true"
	falseValue = "false"
	nullValue  = "null"
)

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_integration"] = ResourceIntegration()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_integration"] = DataSourceIntegration()
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestresources()

	// Run the test suite for suite for the integration package
	m.Run()
}
