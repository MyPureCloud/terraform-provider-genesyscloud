package station

import (
	"sync"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_station_init_test.go file is used to initialize the data sources and resources
   used in testing the station.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

const (
	trueValue  = "true"
	falseValue = "false"
	nullValue  = "null"
)

// providerDataSources holds a map of all registered sites
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered sites
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_user"] = gcloud.ResourceUser()
	providerResources["genesyscloud_telephony_providers_edges_phonebasesettings"] = gcloud.ResourcePhoneBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_phone"] = edgePhone.ResourcePhone()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceStation()
	providerDataSources["genesyscloud_organizations_me"] = gcloud.DataSourceOrganizationsMe()
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestresources()

	// Run the test suite for suite for the integration package
	m.Run()
}
