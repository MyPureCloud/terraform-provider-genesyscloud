package orgauthorization_pairing

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_orgauthorization_pairing_init_test.go file is used to initialize the data sources and resources
   used in testing the orgauthorization_pairing resource.
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

	providerResources[ResourceType] = ResourceOrgauthorizationPairing()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[group.ResourceType] = group.ResourceGroup()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the orgauthorization_pairing package
	initTestResources()
	m.Run()
}
