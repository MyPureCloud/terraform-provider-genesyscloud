package routing_email_route

import (
	"sync"
	"testing"

	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingLanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() resource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	dataSourceMapMutex          sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceRoutingEmailRoute()
	providerResources[routingEmailDomain.ResourceType] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
	providerResources[architectFlow.ResourceType] = architectFlow.ResourceArchitectFlow()
	providerResources[routingSkillGroup.ResourceType] = routingSkillGroup.ResourceRoutingSkillGroup()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.dataSourceMapMutex.Lock()
	defer r.dataSourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceRoutingEmailRoute()
}

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[routingLanguage.ResourceType] = routingLanguage.NewFrameworkRoutingLanguageResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[routingLanguage.ResourceType] = routingLanguage.NewFrameworkRoutingLanguageDataSource
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
	// Run setup function before starting the test suite for routing_email_route package
	initTestResources()

	// Run the test suite for the routing_email_route package
	m.Run()
}
