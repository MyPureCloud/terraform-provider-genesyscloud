package journey_outcome

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	journeySegment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_segment"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
	err       error
)

/*
   The genesyscloud_journey_outcome_init_test.go file is used to initialize the data sources and resources
   used in testing the journey_outcome resource.
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

	providerResources[ResourceType] = ResourceJourneyOutcome()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[architect_schedules.ResourceType] = architect_schedules.ResourceArchitectSchedules()
	providerResources[architect_schedulegroups.ResourceType] = architect_schedulegroups.ResourceArchitectSchedulegroups()
	providerResources[architect_flow.ResourceType] = architect_flow.ResourceArchitectFlow()
	providerResources[journeySegment.ResourceType] = journeySegment.ResourceJourneySegment()

}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceJourneyOutcome()
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
