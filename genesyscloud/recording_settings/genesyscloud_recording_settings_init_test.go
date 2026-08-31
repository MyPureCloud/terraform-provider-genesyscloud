package recording_settings

import (
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
   The genesyscloud_recording_settings_init_test.go file is used to initialize the data sources and resources
   used in testing the recording settings resource.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

var sdkConfig *platformclientv2.Configuration

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

	providerResources[ResourceType] = ResourceRecordingSettings()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceRecordingSettings()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig = provider.SdkConfigurationForTests()

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when running the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the recording_settings package
	initTestResources()

	// Run the test suite for the recording_settings package
	m.Run()
}
