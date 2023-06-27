package outbound_attempt_limit

import (
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	"testing"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initialise_test_resources() (map[string]*schema.Resource,map[string]*schema.Resource) {
	providerDataSources = make(map[string]*schema.Resource)
    providerResources = make(map[string]*schema.Resource)
	providerDataSources["genesyscloud_outbound_attempt_limit"] =  DataSourceOutboundAttemptLimit()
	providerResources["genesyscloud_outbound_attempt_limit"] = ResourceOutboundAttemptLimit()
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

