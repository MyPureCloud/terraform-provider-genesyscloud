package outbound_messagingcampaign

import (
	"sync"
	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_outbound_messagingcampaign_init_test.go file is used to initialize the data sources and resources
   used in testing the outbound_messagingcampaign resource.
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
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceOutboundMessagingcampaign()
	providerResources["genesyscloud_outbound_contact_list"] = obContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_outbound_contactlistfilter"] = obContactListFilter.ResourceOutboundContactlistfilter()
	providerResources["genesyscloud_outbound_callabletimeset"] = obCallableTimeset.ResourceOutboundCallabletimeset()
	providerResources["genesyscloud_outbound_dnclist"] = obDnclist.ResourceOutboundDncList()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceOutboundMessagingcampaign()
	// TODO: Add references
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the outbound_messagingcampaign package
	initTestResources()

	// Run the test suite for the outbound_messagingcampaign package
	m.Run()
}
