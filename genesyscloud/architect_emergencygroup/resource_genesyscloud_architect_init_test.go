package architect_emergencygroup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"testing"
)

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
	providerResources["genesyscloud_architect_ivr"] = architect_ivr.ResourceArchitectIvrConfig()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
	providerResources["genesyscloud_architect_emergencygroup"] = ResourceArchitectEmergencyGroup()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_architect_ivr"] = architect_ivr.DataSourceArchitectIvr()
	providerDataSources["genesyscloud_flow"] = gcloud.DataSourceFlow()
	providerDataSources["genesyscloud_architect_emergencygroup"] = DataSourceArchitectEmergencyGroup()
}

// initTestResources initializes all test_data resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test_data
func TestMain(m *testing.M) {
	// Run setup function before starting the test_data suite for the architect_grammar_language package
	initTestResources()

	// Run the test_data suite for the architect_grammar_language package
	m.Run()
}
