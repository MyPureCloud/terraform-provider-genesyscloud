package outbound_ruleset

import (
	
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	"testing"

	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
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
	providerResources["genesyscloud_outbound_ruleset"] = ResourceOutboundRuleset()
	providerResources["genesyscloud_routing_queue"] =  gcloud.ResourceRoutingQueue()
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_outbound_ruleset"] =  DataSourceOutboundRuleset()
	providerResources["genesyscloud_outbound_contact_list"] = obContactList.ResourceOutboundContactList()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestresources()

	// Run the test suite for outbound ruleset
	m.Run()
}