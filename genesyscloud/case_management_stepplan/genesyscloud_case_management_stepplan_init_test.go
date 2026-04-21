package case_management_stepplan

import (
	"sync"
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	caseplanpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_caseplan"
	stageplanpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_stageplan"
	customerIntent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/customer_intent"
	intentCategory "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/customer_intent_category"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	userpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceCaseManagementStepplan()
	providerResources[stageplanpkg.ResourceType] = stageplanpkg.ResourceCaseManagementStageplan()
	providerResources[caseplanpkg.ResourceType] = caseplanpkg.ResourceCaseManagementCaseplan()
	providerResources[caseplanpkg.PublishResourceType] = caseplanpkg.ResourceCaseManagementCaseplanPublish()
	providerResources[caseplanpkg.CreateVersionResourceType] = caseplanpkg.ResourceCaseManagementCaseplanCreateVersion()
	providerResources[customerIntent.ResourceType] = customerIntent.ResourceCustomerIntent()
	providerResources[intentCategory.ResourceType] = intentCategory.ResourceIntentCategory()
	providerResources[workitemSchema.ResourceType] = workitemSchema.ResourceTaskManagementWorkitemSchema()
	providerResources[workbin.ResourceType] = workbin.ResourceTaskManagementWorkbin()
	providerResources[worktype.ResourceType] = worktype.ResourceTaskManagementWorktype()
	providerResources[userpkg.ResourceType] = userpkg.ResourceUser()
}

func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceCaseManagementStepplan()
	providerDataSources[stageplanpkg.ResourceType] = stageplanpkg.DataSourceCaseManagementStageplan()
	providerDataSources[caseplanpkg.ResourceType] = caseplanpkg.DataSourceCaseManagementCaseplan()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	reg := &registerTestInstance{}
	reg.registerTestResources()
	reg.registerTestDataSources()
}

func TestMain(m *testing.M) {
	initTestResources()
	m.Run()
}
