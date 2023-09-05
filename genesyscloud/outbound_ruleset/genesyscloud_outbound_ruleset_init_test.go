package outbound_ruleset

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
)

/*
   The genesyscloud_outbound_ruleset_init_test.go file is used to initialize the data sources and resources
   used in testing the outbound_ruleset resource.
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
	providerResources["genesyscloud_outbound_ruleset"] = ResourceOutboundRuleset()
	providerResources["genesyscloud_routing_queue"] = gcloud.ResourceRoutingQueue()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_outbound_ruleset"] = DataSourceOutboundRuleset()
	providerResources["genesyscloud_outbound_contact_list"] = obContactList.ResourceOutboundContactList()
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestResources()
	reg_instance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the outbound_ruleset package
	initTestresources()

	// Run the test suite for the outbound_ruleset package
	m.Run()
}
