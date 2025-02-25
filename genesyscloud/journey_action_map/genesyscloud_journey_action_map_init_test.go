package journey_action_map

import (
	"log"
	"sync"
	architectFlow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	architectSchedulegroups "terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	journeyOutcome "terraform-provider-genesyscloud/genesyscloud/journey_outcome"
	journeySegment "terraform-provider-genesyscloud/genesyscloud/journey_segment"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v152/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
	err       error
)

/*
   The genesyscloud_journey_action_map_init_test.go file is used to initialize the data sources and resources
   used in testing the journey_action_map resource.
*/

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

	providerResources[ResourceType] = ResourceJourneyActionMap()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[architectSchedules.ResourceType] = architectSchedules.ResourceArchitectSchedules()
	providerResources[architectSchedulegroups.ResourceType] = architectSchedulegroups.ResourceArchitectSchedulegroups()
	providerResources[architectFlow.ResourceType] = architectFlow.ResourceArchitectFlow()
	providerResources[journeyOutcome.ResourceType] = journeyOutcome.ResourceJourneyOutcome()
	providerResources[journeySegment.ResourceType] = journeySegment.ResourceJourneySegment()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceJourneyActionMap()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	if sdkConfig, err = provider.AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the architect_schedulegroups package
	initTestResources()

	// Run the test suite for the architect_schedulegroups package
	m.Run()
}
