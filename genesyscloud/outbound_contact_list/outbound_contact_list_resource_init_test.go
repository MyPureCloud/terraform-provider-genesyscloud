package outbound_contact_list

import (
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	ob_attempt_limit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initialise_test_resources() (map[string]*schema.Resource,map[string]*schema.Resource) {
	providerDataSources = make(map[string]*schema.Resource)
    providerResources = make(map[string]*schema.Resource)
	providerResources["genesyscloud_outbound_contact_list"] = ResourceOutboundContactList()
	providerDataSources["genesyscloud_outbound_contact_list"] =  DataSourceOutboundContactList()
	providerResources["genesyscloud_outbound_attempt_limit"] = ob_attempt_limit.ResourceOutboundAttemptLimit()
	providerDataSources["genesyscloud_outbound_attempt_limit"] =  ob_attempt_limit.DataSourceOutboundAttemptLimit()

	providerDataSources["genesyscloud_auth_division_home"] =  gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_auth_division_home"] =  gcloud.DataSourceAuthDivisionHome()
	//providerResources["genesyscloud_outbound_contact_list"] = ob.ResourceOutboundContactList()
	return providerResources,providerDataSources
}


func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initialise_test_resources()

	// Run the test suite for outbound ruleset
	m.Run()
}


// main lo unna funtion  is called by individual packages

// main has struct wit those function
// ind packages have interface 


// outbound lo una function should be called by ob_ruleset

