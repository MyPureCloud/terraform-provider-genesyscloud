package process_automation_trigger

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"

	//obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_processautomation_trigger"] = ResourceProcessAutomationTrigger()
	providerResources["genesyscloud_flow"] = architect_flow.ResourceArchitectFlow()
}

func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_processautomation_trigger"] = dataSourceProcessAutomationTrigger()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for Process Automation Trigger
	initTestResources()

	// Run the test suite
	m.Run()
}
