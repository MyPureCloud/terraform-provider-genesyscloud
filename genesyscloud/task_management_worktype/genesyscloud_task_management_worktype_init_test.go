package task_management_worktype

import (
	"sync"
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	routingLanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	script "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/scripts"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_task_management_worktype_init_test.go file is used to initialize the data sources and resources
   used in testing the task_management_worktype resource.
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

	providerResources[ResourceType] = ResourceTaskManagementWorktype()
	providerResources[workbin.ResourceType] = workbin.ResourceTaskManagementWorkbin()
	providerResources[workitemSchema.ResourceType] = workitemSchema.ResourceTaskManagementWorkitemSchema()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
	providerResources[script.ResourceType] = script.ResourceScript()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceTaskManagementWorktype()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
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
	// Run setup function before starting the test suite for the task_management_worktype package
	initTestResources()

	// Run the test suite for the task_management_worktype package
	m.Run()
}
