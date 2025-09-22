package outbound_campaignrule

import (
	"sync"
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	flow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	obResponseSet "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	outboundCampaign "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	outboundContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	outboundSequence "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_outbound_campaignrule_init_test.go file is used to initialize the data sources and resources
   used in testing the outbound_campaignrule resource.
*/

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() resource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	datasourceMapMutex          sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceOutboundCampaignrule()
	providerResources[outboundCampaign.ResourceType] = outboundCampaign.ResourceOutboundCampaign()
	providerResources[outboundContactList.ResourceType] = outboundContactList.ResourceOutboundContactList()
	providerResources[flow.ResourceType] = flow.ResourceArchitectFlow()
	providerResources[obResponseSet.ResourceType] = obResponseSet.ResourceOutboundCallanalysisresponseset()
	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[edgeSite.ResourceType] = edgeSite.ResourceSite()
	providerResources[outboundSequence.ResourceType] = outboundSequence.ResourceOutboundSequence()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceOutboundCampaignrule()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[routingWrapupcode.ResourceType] = routingWrapupcode.NewRoutingWrapupcodeFrameworkResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[routingWrapupcode.ResourceType] = routingWrapupcode.NewRoutingWrapupcodeFrameworkDataSource
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the outbound_campaignrule package
	initTestResources()

	// Run the test suite for the outbound_campaignrule package
	m.Run()
}
