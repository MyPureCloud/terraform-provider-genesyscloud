package telephony_providers_edges_phone

import (
	"log"
	"sync"
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	phoneBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_telephony_providers_edges_phone_init_test.go file is used to initialize the data sources and resources
   used in testing the edges phones.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

var (
	// providerDataSources holds a map of all registered sites
	providerDataSources map[string]*schema.Resource

	// providerResources holds a map of all registered sites
	providerResources map[string]*schema.Resource

	sdkConfig *platformclientv2.Configuration
	authErr   error
)

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourcePhone()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[phoneBaseSettings.ResourceType] = phoneBaseSettings.ResourcePhoneBaseSettings()
	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[edgeSite.ResourceType] = edgeSite.ResourceSite()
	providerResources[didPool.ResourceType] = didPool.ResourceTelephonyDidPool()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourcePhone()
	providerDataSources[gcloud.DataSourceOrganizationsMeResourceType] = gcloud.DataSourceOrganizationsMe()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk for the package telephony_providers_edges_phone: %v", authErr)
	}
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestResources()

	// Run the test suite for the integration package
	m.Run()
}
