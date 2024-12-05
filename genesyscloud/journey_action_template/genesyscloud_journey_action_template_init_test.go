package journey_action_template

import (
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/journey_action_map"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
	err       error
)

/*
   The genesyscloud_journey_action_template_init_test.go file is used to initialize the data sources and resources
   used in testing the journey_action_template resource.
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

	providerResources[ResourceType] = ResourceJourneyActionTemplate()
	providerResources[journey_action_map.ResourceType] = journey_action_map.ResourceJourneyActionMap()
	providerResources["genesyscloud_journey_segment"] = gcloud.ResourceJourneySegment()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceJourneyActionTemplate()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	if sdkConfig, err = provider.AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the journey_action_template package
	initTestResources()

	// Run the test suite for the journey_action_template package
	m.Run()
}
