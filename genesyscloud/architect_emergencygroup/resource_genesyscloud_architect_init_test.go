package architect_emergencygroup

import (
	"sync"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	providerResources[architect_ivr.ResourceType] = architect_ivr.ResourceArchitectIvrConfig()
	providerResources[flow.ResourceType] = flow.ResourceArchitectFlow()
	providerResources[ResourceType] = ResourceArchitectEmergencyGroup()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources[architect_ivr.ResourceType] = architect_ivr.DataSourceArchitectIvr()
	providerDataSources[flow.ResourceType] = flow.DataSourceArchitectFlow()
	providerDataSources[ResourceType] = DataSourceArchitectEmergencyGroup()
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
	// Run setup function before starting the test_data suite for the architect_grammar_language package
	initTestResources()

	// Run the test_data suite for the architect_grammar_language package
	m.Run()
}
