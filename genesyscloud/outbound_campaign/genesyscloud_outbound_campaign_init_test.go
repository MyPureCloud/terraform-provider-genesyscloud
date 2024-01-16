package outbound_campaign

import (
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/outbound"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	outboundRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

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
	providerResources[resourceName] = ResourceOutboundCampaign()
	providerResources["genesyscloud_outbound_contact_list"] = outboundContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
	providerResources["genesyscloud_outbound_callanalysisresponseset"] = outbound.ResourceOutboundCallAnalysisResponseSet()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_telephony_providers_edges_site"] = telephonyProvidersEdgesSite.ResourceSite()
	providerResources["genesyscloud_outbound_dnclist"] = outbound.ResourceOutboundDncList()
	providerResources["genesyscloud_routing_queue"] = gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_outbound_contactlistfilter"] = outbound.ResourceOutboundContactListFilter()
	providerResources["genesyscloud_outbound_ruleset"] = outboundRuleset.ResourceOutboundRuleset()
	providerResources["genesyscloud_outbound_callabletimeset"] = outbound.ResourceOutboundCallabletimeset()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources[resourceName] = DataSourceOutboundCampaign()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = gcloud.AuthorizeSdk()
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
