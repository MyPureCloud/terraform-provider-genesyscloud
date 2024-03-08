package journey_outcome_predictor

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/resource_genesyscloud_journey_outcome.go"
	"terraform-provider-genesyscloud/genesyscloud/data_source_genesyscloud_journey_outcome.go"
	"testing"
)

// providerDataSources holds a map of all registered webdeployments_configuration
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all webdeployments_configuration
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceJourneyOutcomePredictor()
	providerResources["genesyscloud_journey_outcome"] = ResourceJourneyOutcome()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerResources["genesyscloud_journey_outcome"] = dataSourceJourneyOutcome()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestResources()

	// Run the test suite for the integration package
	m.Run()
}
