package architect_schedulegroups

import (
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_architect_schedulegroups_init_test.go file is used to initialize the data sources and resources
   used in testing the architect_schedulegroups resource.
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

	providerResources[resourceName] = ResourceArchitectSchedulegroups()
	providerResources["genesyscloud_architect_schedules"] = architectSchedules.ResourceArchitectSchedules()
	providerResources["genesyscloud_auth_division"] = genesyscloud.ResourceAuthDivision()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceArchitectSchedulegroups()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
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
