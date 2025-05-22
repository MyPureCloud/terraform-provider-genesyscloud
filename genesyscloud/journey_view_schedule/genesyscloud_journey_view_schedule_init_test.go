package journey_view_schedule

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	journeyViews "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_views"
)

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceJourneyViewSchedule()
	providerResources[journeyViews.ResourceType] = journeyViews.ResourceJourneyViews()
}
func (r *registerTestInstance) registerTestDataSources() {
	//There are no data sources for this resource
}

// initTestResources initializes all test resources and data sources
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when running tests
func TestMain(m *testing.M) {
	initTestResources()

	// Run the tests
	m.Run()
}
