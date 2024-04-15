package architect_datatable_row

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	dt "terraform-provider-genesyscloud/genesyscloud/architect_datatable"

	"testing"
)

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerResources["genesyscloud_architect_datatable_row"] = ResourceArchitectDatatableRow()
	providerResources["genesyscloud_architect_datatable"] = dt.ResourceArchitectDatatable()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	//There are no data sources for this resource
}

// initTestResources initializes all test_data resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test_data
func TestMain(m *testing.M) {
	// Run setup function before starting the test_data suite for the package
	initTestResources()

	// Run the test_data suite for the architect_grammar_language package
	m.Run()
}
