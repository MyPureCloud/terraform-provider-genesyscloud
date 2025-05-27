package outbound_messagingcampaign

import (
	obCallableTimeset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDigitalRuleSet "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDnclist "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	responseManagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	responseManagementResponse "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"sync"
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

	providerResources[ResourceType] = ResourceOutboundMessagingcampaign()
	providerResources[obContactList.ResourceType] = obContactList.ResourceOutboundContactList()
	providerResources[obContactListFilter.ResourceType] = obContactListFilter.ResourceOutboundContactlistfilter()
	providerResources[obCallableTimeset.ResourceType] = obCallableTimeset.ResourceOutboundCallabletimeset()
	providerResources[obDnclist.ResourceType] = obDnclist.ResourceOutboundDncList()
	providerResources[obDigitalRuleSet.ResourceType] = obDigitalRuleSet.ResourceOutboundDigitalruleset()
	providerResources[responseManagementLibrary.ResourceType] = responseManagementLibrary.ResourceResponsemanagementLibrary()
	providerResources[responseManagementResponse.ResourceType] = responseManagementResponse.ResourceResponsemanagementResponse()
	providerResources[routingEmailRoute.ResourceType] = routingEmailRoute.ResourceRoutingEmailRoute()
	providerResources[routingEmailDomain.ResourceType] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceOutboundMessagingcampaign()
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
