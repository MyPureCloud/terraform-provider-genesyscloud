package external_contacts

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_externalcontacts_contacts_init_test.go file is used to initialize the data sources and resources
   used in testing the externalcontacts_contacts resource.

   Please make sure you register ALL resources and data sources your test cases will use.
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

	providerResources["genesyscloud_externalcontacts_contact"] = ResourceExternalContact()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_externalcontacts_contact"] = DataSourceExternalContactsContact()

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
	// Run setup function before starting the test suite for external_contacts package
	initTestresources()

	// Run the test suite for suite for the external_contacts package
	m.Run()
}
