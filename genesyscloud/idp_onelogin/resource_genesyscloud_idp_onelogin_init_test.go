package idp_onelogin

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_idp_onelogin_init_test.go file is used to initialize the data sources and resources
   used in testing the idp_onelogin resource.
*/

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceIdpOnelogin()
}

// initTestResources initializes all test_data resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test_data
func TestMain(m *testing.M) {
	// Run setup function before starting the test_data suite for the idp_onelogin package
	initTestResources()

	// Run the test_data suite for the idp_onelogin package
	m.Run()
}
