package task_management_worktype

import (
	"sync"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"testing"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"

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

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceTaskManagementWorktype()
	providerResources["genesyscloud_task_management_workbin"] = workbin.ResourceTaskManagementWorkbin()
	providerResources["genesyscloud_task_management_workitem_schema"] = workitemSchema.ResourceTaskManagementWorkitemSchema()
	providerResources["genesyscloud_routing_language"] = routingLanguage.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_routing_skill"] = routingSkill.ResourceRoutingSkill()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceTaskManagementWorktype()
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
	// Run setup function before starting the test suite for the task_management_worktype package
	initTestResources()

	// Run the test suite for the task_management_worktype package
	m.Run()
}
