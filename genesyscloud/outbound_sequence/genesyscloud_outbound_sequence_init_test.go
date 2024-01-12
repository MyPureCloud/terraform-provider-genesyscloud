package outbound_sequence

import (
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/outbound"
	outboundCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_outbound_sequence_init_test.go file is used to initialize the data sources and resources
   used in testing the outbound_sequence resource.
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

	providerResources[resourceName] = ResourceOutboundSequence()
	providerResources["genesyscloud_outbound_campaign"] = outboundCampaign.ResourceOutboundCampaign()
	providerResources["genesyscloud_outbound_contact_list"] = outboundContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
	providerResources["genesyscloud_outbound_callanalysisresponseset"] = outbound.ResourceOutboundCallAnalysisResponseSet()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceOutboundSequence()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = gcloud.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk for the package outbound_sequence: %v", authErr)
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the outbound_sequence package
	initTestResources()

	// Run the test suite for the outbound_sequence package
	m.Run()
}
