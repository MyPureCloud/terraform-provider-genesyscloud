package scripts

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Initializes and registers data sources and resources for the scripts test package
*/
var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResource registers the CX as Code resources used in test
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceScript()
}

// registerTestResource registers the CX as Code resources used in test
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceScript()
}

// initTestResources initializes all the data sources and resources
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain acts as a setup class that gets called before the tests cases to register resources
func TestMain(m *testing.M) {

	// Run setup function before starting the test suite for Outbound Package
	initTestResources()

	// Run the test suite for outbound
	m.Run()
}
