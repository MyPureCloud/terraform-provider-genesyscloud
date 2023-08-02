package process_automation_trigger

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	//obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestResources()
	reg_instance.registerTestDataSources()
}

type registerTestInstance struct {
}

func (r *registerTestInstance) registerTestResources() {
	providerResources["genesyscloud_processautomation_trigger"] = ResourceProcessAutomationTrigger()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_processautomation_trigger"] = dataSourceProcessAutomationTrigger()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for Process Automation Trigger
	initTestresources()

	// Run the test suite
	m.Run()
}
