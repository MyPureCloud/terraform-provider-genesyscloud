package task_management_worktype_flow_onattributechange_rule

import (
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	worktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypeStatus "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_task_management_onattributechange_rule_init_test.go file is used to initialize the data sources and resources
   used in testing the task_management_worktype_flow_onattributechange_rule resource.
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

	providerResources[ResourceType] = ResourceTaskManagementOnAttributeChangeRule()
	providerResources[worktype.ResourceType] = worktype.ResourceTaskManagementWorktype()
	providerResources[workbin.ResourceType] = workbin.ResourceTaskManagementWorkbin()
	providerResources[worktypeStatus.ResourceType] = worktypeStatus.ResourceTaskManagementWorktypeStatus()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceTaskManagementOnAttributeChangeRule()
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
	// Run setup function before starting the test suite for the task_management_worktype_flow_onattributechange_rule package
	initTestResources()

	// Run the test suite for the task_management_worktype_flow_onattributechange_rule package
	m.Run()
}
