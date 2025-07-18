package station

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	edgePhone "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	phoneBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_station_init_test.go file is used to initialize the data sources and resources
   used in testing the station.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

// providerDataSources holds a map of all registered sites
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered sites
var providerResources map[string]*schema.Resource

var sdkConfig *platformclientv2.Configuration
var authErr error

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[phoneBaseSettings.ResourceType] = phoneBaseSettings.ResourcePhoneBaseSettings()
	providerResources[edgePhone.ResourceType] = edgePhone.ResourcePhone()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceStation()
	providerDataSources[gcloud.DataSourceOrganizationsMeResourceType] = gcloud.DataSourceOrganizationsMe()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatal(authErr)
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
