package routing_queue

import (
	"sync"
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	responseManagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() resource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	datasourceMapMutex          sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceRoutingQueue()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
	providerResources[group.ResourceType] = group.ResourceGroup()
	providerResources[architectFlow.ResourceType] = architectFlow.ResourceArchitectFlow()
	providerResources[routingSkillGroup.ResourceType] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources[architect_user_prompt.ResourceType] = architect_user_prompt.ResourceArchitectUserPrompt()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[responseManagementLibrary.ResourceType] = responseManagementLibrary.ResourceResponsemanagementLibrary()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceRoutingQueue()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[routingWrapupcode.ResourceType] = routingWrapupcode.NewRoutingWrapupcodeFrameworkResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[routingWrapupcode.ResourceType] = routingWrapupcode.NewRoutingWrapupcodeFrameworkDataSource
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the routing_queue package
	initTestResources()

	// Run the test suite for the routing_queue package
	m.Run()
}
