package routing_queue_conditional_group_routing

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/group"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	"testing"
)

/*
The genesyscloud_routing_queue_conditional_group_routing_init_test.go file is used to initialize the data sources and resources
used in testing the routing_queue_conditional_group_routing resource.
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

	providerResources[resourceName] = ResourceRoutingQueueConditionalGroupRouting()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_routing_skill_group"] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources["genesyscloud_user"] = gcloud.ResourceUser()
	providerResources["genesyscloud_group"] = group.ResourceGroup()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for routing_queue_conditional_group_routing package
	initTestResources()

	// Run the test suite for the routing_queue_conditional_group_routing package
	m.Run()
}
