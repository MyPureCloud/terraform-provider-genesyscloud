package routing_utilization

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	"testing"
)

/*
The genesyscloud_routing_utilization_init_test.go file is used to initialize the data sources and resources
used in testing the routing_utilization resource.
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

	providerResources[resourceName] = ResourceRoutingUtilization()
	providerResources["genesyscloud_routing_utilization_label"] = routing_utilization_label.ResourceRoutingUtilizationLabel()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the package
	initTestResources()

	// Run the test suite for the package
	m.Run()
}
