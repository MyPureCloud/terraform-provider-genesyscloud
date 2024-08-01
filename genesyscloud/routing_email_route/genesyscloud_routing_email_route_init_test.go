package routing_email_route

import (
	"sync"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"

	architectFlow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_routing_email_route_init_test.go file is used to initialize the data sources and resources
used in testing the routing_email_route resource.
*/

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	dataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceRoutingEmailRoute()
	providerResources["genesyscloud_routing_email_domain"] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_routing_language"] = routingLanguage.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_skill"] = routingSkill.ResourceRoutingSkill()
	providerResources["genesyscloud_flow"] = architectFlow.ResourceArchitectFlow()
	providerResources["genesyscloud_routing_skill_group"] = routingSkillGroup.ResourceRoutingSkillGroup()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.dataSourceMapMutex.Lock()
	defer r.dataSourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceRoutingEmailRoute()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for routing_email_route package
	initTestResources()

	// Run the test suite for the routing_email_route package
	m.Run()
}
