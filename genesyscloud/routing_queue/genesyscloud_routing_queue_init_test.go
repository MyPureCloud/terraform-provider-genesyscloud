package routing_queue

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	architectFlow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	"terraform-provider-genesyscloud/genesyscloud/group"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	routingWrapupcode "terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_routing_queue_init_test.go file is used to initialize the data sources and resources
   used in testing the routing_queue resource.
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

	providerResources[resourceName] = ResourceRoutingQueue()
	providerResources["genesyscloud_user"] = user.ResourceUser()
	providerResources["genesyscloud_routing_skill"] = routingSkill.ResourceRoutingSkill()
	providerResources["genesyscloud_group"] = group.ResourceGroup()
	providerResources["genesyscloud_routing_wrapupcode"] = routingWrapupcode.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_flow"] = architectFlow.ResourceArchitectFlow()
	providerResources["genesyscloud_routing_skill_group"] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources["genesyscloud_architect_user_prompt"] = architect_user_prompt.ResourceArchitectUserPrompt()
	providerResources["genesyscloud_auth_division"] = authDivision.ResourceAuthDivision()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceRoutingQueue()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
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
	// Run setup function before starting the test suite for the routing_queue package
	initTestResources()

	// Run the test suite for the routing_queue package
	m.Run()
}
