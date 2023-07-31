package webdeployments_configuration

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_webdeployments_configuration.go file is used to initialize the data sources and resources
   used in testing the webdeployments_configuration resource.
   Please make sure you register ALL resources and data sources your test cases will use.
*/

// providerDataSources holds a map of all registered resources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered datasources.
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_webdeployments_configuration"] = ResourceWebDeploymentConfiguration()

}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_webdeployments_configuration"] = DataSourceWebDeploymentsConfiguration()

}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

// TestMain is a "setup" function called by the testing framework when run the
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for webdeployments_configuration package
	initTestresources()

	// Run the test suite for suite for the webdeployments_configuration package
	m.Run()
}
