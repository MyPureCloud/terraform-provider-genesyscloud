package routing_skill_group

import (
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	"testing"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceRoutingSkillGroup()
	providerResources["genesyscloud_auth_division"] = genesyscloud.ResourceAuthDivision()
	providerResources["genesyscloud_user"] = genesyscloud.ResourceUser()
	providerResources["genesyscloud_routing_skill"] = routingSkill.ResourceRoutingSkill()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceRoutingSkillGroup()
	providerDataSources["genesyscloud_auth_division_home"] = genesyscloud.DataSourceAuthDivisionHome()

}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for routing_skill_group package
	initTestResources()

	// Run the test suite for the routing_skill_group package
	m.Run()
}
