package telephony_providers_edges_did

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	archIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

	providerResources[didPool.ResourceType] = didPool.ResourceTelephonyDidPool()
	providerResources[archIvr.ResourceType] = archIvr.ResourceArchitectIvrConfig()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceDid()
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
		log.Println("telephony_providers_edges_did.TestMain: ", err.Error())
		sdkConfig = platformclientv2.GetDefaultConfiguration()
	}

	// Run setup function before starting the test suite for package
	initTestResources()

	// Run the test suite for the package
	m.Run()
}
