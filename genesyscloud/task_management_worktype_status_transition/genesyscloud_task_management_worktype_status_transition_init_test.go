package task_management_worktype_status_transition

import (
	"sync"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypestatus "terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_task_management_worktype_status_transition_init_test.go file is used to initialize the data sources and resources
   used in testing the task_management_worktype_status resource.
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

	providerResources[ResourceType] = ResourceTaskManagementWorktypeStatusTransition()
	providerResources[worktype.ResourceType] = worktype.ResourceTaskManagementWorktype()
	providerResources[workbin.ResourceType] = workbin.ResourceTaskManagementWorkbin()
	providerResources[worktypestatus.ResourceType] = worktypestatus.ResourceTaskManagementWorktypeStatus()
	providerResources[workitemSchema.ResourceType] = workitemSchema.ResourceTaskManagementWorkitemSchema()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceTaskManagementWorktypeStatusTransition()
	providerDataSources[worktype.ResourceType] = worktype.DataSourceTaskManagementWorktype()
	providerDataSources[workbin.ResourceType] = workbin.DataSourceTaskManagementWorkbin()
	providerDataSources[worktypestatus.ResourceType] = worktypestatus.DataSourceTaskManagementWorktypeStatus()
	providerDataSources[workitemSchema.ResourceType] = workitemSchema.DataSourceTaskManagementWorkitemSchema()
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
	// Run setup function before starting the test suite for the task_management_worktype_status package
	initTestResources()

	// Run the test suite for the task_management_worktype_status_transition package
	m.Run()
}
