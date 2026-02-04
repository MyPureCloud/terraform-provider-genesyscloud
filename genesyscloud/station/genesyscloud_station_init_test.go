package station

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	edgePhone "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	phoneBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_station_init_test.go file is used to initialize the data sources and resources
   used in testing the station resource.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

var (
	// providerDataSources holds a map of all registered datasources
	providerDataSources map[string]*schema.Resource

	// providerResources holds a map of all registered resources
	providerResources map[string]*schema.Resource

	// frameworkResources holds a map of all registered Framework resources
	frameworkResources map[string]func() resource.Resource

	// frameworkDataSources holds a map of all registered Framework data sources
	frameworkDataSources map[string]func() datasource.DataSource

	sdkConfig *platformclientv2.Configuration
	authErr   error
)

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	datasourceMapMutex          sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

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

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[user.ResourceType] = user.NewUserFrameworkResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[user.ResourceType] = user.NewUserFrameworkDataSource
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk for the package station: %v", authErr)
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for station package
	initTestResources()

	// Run the test suite for the station package
	m.Run()
}
