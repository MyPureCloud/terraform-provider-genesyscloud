package outbound_attempt_limit

import (
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
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

type registerTestInstance struct{
}

func (r *registerTestInstance) registerTestResources() {
	providerResources["genesyscloud_outbound_attempt_limit"] = ResourceOutboundAttemptLimit()
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_outbound_attempt_limit"] =  DataSourceOutboundAttemptLimit()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestresources()

	// Run the test suite for outbound ruleset
	m.Run()
}


// main lo unna funtion  is called by individual packages

// main has struct wit those function
// ind packages have interface 


// outbound lo una function should be called by ob_ruleset

