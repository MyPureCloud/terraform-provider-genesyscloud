package routing_queue_outbound_email_address

import (
	"sync"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"testing"
)

/*
The genesyscloud_routing_queue_outbound_email_address_init_test.go file is used to initialize the data sources and resources
used in testing the routing_queue_outbound_email_address resource.
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

	providerResources[ResourceType] = ResourceRoutingQueueOutboundEmailAddress()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[routingEmailRoute.ResourceType] = routingEmailRoute.ResourceRoutingEmailRoute()
	providerResources[routingEmailDomain.ResourceType] = routingEmailDomain.ResourceRoutingEmailDomain()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for routing_queue_outbound_email_address package
	initTestResources()

	// Run the test suite for the routing_queue_outbound_email_address package
	m.Run()
}
