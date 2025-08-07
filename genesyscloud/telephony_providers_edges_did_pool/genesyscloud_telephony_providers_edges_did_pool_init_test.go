package telephony_providers_edges_did_pool

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_telephony_providers_edges_did_pool_init_test.go file is used to initialize the data sources and resources
   used in testing the telephony_providers_edges_did_pool resource.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

// providerDataSources holds a map of all registered data sources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

var sdkConfig *platformclientv2.Configuration

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceTelephonyDidPool()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceDidPool()
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
	var err error
	sdkConfig, err = provider.AuthorizeSdk()
	if err != nil {
		log.Println("telephony_providers_edges_did_pool.TestMain: ", err.Error())
		sdkConfig = platformclientv2.GetDefaultConfiguration()
	}

	// Run setup function before starting the test suite for telephony_providers_edges_did_pool package
	initTestResources()

	// Run the test suite for the telephony_providers_edges_did_pool package
	m.Run()
}
