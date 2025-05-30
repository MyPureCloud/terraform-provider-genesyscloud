package journey_outcome_predictor

import (
	journeyOutcome "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_outcome"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	providerDataSources map[string]*schema.Resource
	providerResources   map[string]*schema.Resource
)

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceJourneyOutcomePredictor()
	providerResources[journeyOutcome.ResourceType] = journeyOutcome.ResourceJourneyOutcome()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for journey outcome predictor package
	initTestResources()

	// Run the test suite for the journey outcome predictor package
	m.Run()
}
