package outbound_campaign

import (
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obResponseSet "terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	outboundRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_outbound_campaign_init_test.go file is used to initialize the data sources and resources
   used in testing the outbound_campaign resource.
*/

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

var (
	sdkConfig *platformclientv2.Configuration
	authErr   error
)

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceOutboundCampaign()
	providerResources["genesyscloud_outbound_contact_list"] = outboundContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_flow"] = flow.ResourceArchitectFlow()
	providerResources["genesyscloud_outbound_callanalysisresponseset"] = obResponseSet.ResourceOutboundCallanalysisresponseset()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_telephony_providers_edges_site"] = telephonyProvidersEdgesSite.ResourceSite()
	providerResources["genesyscloud_outbound_dnclist"] = obDnclist.ResourceOutboundDncList()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_outbound_contactlistfilter"] = obContactListFilter.ResourceOutboundContactlistfilter()
	providerResources["genesyscloud_outbound_ruleset"] = outboundRuleset.ResourceOutboundRuleset()
	providerResources["genesyscloud_outbound_callabletimeset"] = obCallableTimeset.ResourceOutboundCallabletimeset()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceOutboundCampaign()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk for package outbound_campaign: %v", authErr)
	}
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the outbound_campaign package
	initTestResources()

	// Run the test suite for the outbound_campaign package
	m.Run()
}
