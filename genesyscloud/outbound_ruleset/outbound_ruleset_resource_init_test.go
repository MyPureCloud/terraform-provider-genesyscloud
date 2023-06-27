package outbound_ruleset

import (
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	ob_contact_list "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initialise_test_resources() (map[string]*schema.Resource,map[string]*schema.Resource) {
	providerDataSources = make(map[string]*schema.Resource)
    providerResources = make(map[string]*schema.Resource)
	providerResources["genesyscloud_outbound_ruleset"] = ResourceOutboundRuleset()
	providerDataSources["genesyscloud_outbound_ruleset"] =  DataSourceOutboundRuleset()
	providerResources["genesyscloud_routing_queue"] =  gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_outbound_contact_list"] = ob_contact_list.ResourceOutboundContactList()

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

