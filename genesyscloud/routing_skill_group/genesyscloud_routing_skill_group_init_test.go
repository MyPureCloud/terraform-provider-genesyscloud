package routing_skill_group

import (
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"testing"

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

	providerResources[ResourceType] = ResourceRoutingSkillGroup()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceRoutingSkillGroup()
	providerDataSources[genesyscloud.ResourceType] = genesyscloud.DataSourceAuthDivisionHome()

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
