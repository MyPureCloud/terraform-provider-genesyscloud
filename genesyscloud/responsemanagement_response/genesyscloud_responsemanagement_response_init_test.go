package responsemanagement_response

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	respmanagementLibrary "terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	respManagementRespAsset "terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
	"testing"
)

/*
   The genesyscloud_responsemanagement_response_init_test.go file is used to initialize the data sources and resources
   used in testing the responsemanagement_response resource.
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

	providerResources[resourceName] = ResourceResponsemanagementResponse()
	providerResources["genesyscloud_responsemanagement_responseasset"] = respManagementRespAsset.ResourceResponseManagementResponseAsset()
	providerResources["genesyscloud_responsemanagement_library"] = respmanagementLibrary.ResourceResponsemanagementLibrary()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceResponsemanagementResponse()
}

// initTestresources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the responsemanagement_response package
	initTestResources()

	// Run the test suite for the responsemanagement_response package
	m.Run()
}
