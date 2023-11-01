package teams_resource

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
   The genesyscloud_teams_resource_init_test.go file is used to initialize the data sources and resources
   used in testing the teams_resource resource.
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
	providerResources["genesyscloud_teams_resource"] = ResourceTeamsResource()
	// TODO: Add references
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_teams_resource"] = DataSourceTeamsResource()
	// TODO: Add references
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestResources()
	reg_instance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the teams_resource package
	initTestresources()

	// Run the test suite for the teams_resource package
	m.Run()
}
