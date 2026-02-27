package group_greeting

import (
	"sync"
	"testing"

	groupResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	userResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
This file initializes the data sources and resources used in testing the group greeting resource.
*/

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

	providerResources[userResource.ResourceType] = userResource.ResourceUser()
	providerResources[groupResource.ResourceType] = groupResource.ResourceGroup()
	providerResources[ResourceType] = ResourceGroupGreeting()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for group_greeting package
	initTestResources()

	// Run the test suite for the group_greeting package
	m.Run()
}
