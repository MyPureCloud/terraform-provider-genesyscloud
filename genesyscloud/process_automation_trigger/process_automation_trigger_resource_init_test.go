package process_automation_trigger

import (
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	//ob_ruleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initialise_test_resources() (map[string]*schema.Resource,map[string]*schema.Resource) {
	providerDataSources = make(map[string]*schema.Resource)
    providerResources = make(map[string]*schema.Resource)
	providerResources["genesyscloud_processautomation_trigger"] = resourceProcessAutomationTrigger()
	providerDataSources["genesyscloud_processautomation_trigger"] =  dataSourceProcessAutomationTrigger()
	providerDataSources["genesyscloud_auth_division_home"] =  gcloud.DataSourceAuthDivisionHome()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()

	return providerResources,providerDataSources
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initialise_test_resources()

	// Run the test suite
	m.Run()
}
