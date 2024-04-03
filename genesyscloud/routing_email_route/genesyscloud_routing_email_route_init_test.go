package routing_email_route

import (
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_routing_email_route_init_test.go file is used to initialize the data sources and resources
used in testing the routing_email_route resource.
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

	providerResources["genesyscloud_routing_email_route"] = ResourceRoutingEmailRoute()
	providerResources["genesyscloud_routing_email_domain"] = genesyscloud.ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_queue"] = genesyscloud.ResourceRoutingQueue()
	providerResources["genesyscloud_routing_language"] = genesyscloud.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_skill"] = genesyscloud.ResourceRoutingSkill()
	providerResources["genesyscloud_flow"] = architect_flow.ResourceArchitectFlow()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for routing_email_route package
	initTestResources()

	// Run the test suite for the routing_email_route package
	m.Run()
}
