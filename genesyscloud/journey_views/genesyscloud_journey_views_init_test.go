package journey_views

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

//var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceJourneyViews()
	providerResources["genesyscloud_user"] = genesyscloud.ResourceUser()

}

// registerTestDataSources registers all data sources used in the tests.
/* TODO:
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceGroup()
}*/

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	//TODO: providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	//TODO: regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the journey_views package
	initTestResources()

	// Run the test suite for the journey_views package
	m.Run()
}
