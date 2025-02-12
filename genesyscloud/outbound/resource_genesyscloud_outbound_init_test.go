package outbound

import (
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	flow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	obAttemptLimit "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCallableTimeset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	outboundCampaignrule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDigitalRuleset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDnclist "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	obRuleset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	outboundSequence "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {

	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceOutboundMessagingCampaign()
	providerResources[obCallableTimeset.ResourceType] = obCallableTimeset.ResourceOutboundCallabletimeset()
	providerResources[outboundCampaignrule.ResourceType] = outboundCampaignrule.ResourceOutboundCampaignrule()
	providerResources[obAttemptLimit.ResourceType] = obAttemptLimit.ResourceOutboundAttemptLimit()
	providerResources[obContactListFilter.ResourceType] = obContactListFilter.ResourceOutboundContactlistfilter()
	providerResources[obContactList.ResourceType] = obContactList.ResourceOutboundContactList()
	providerResources[outboundSequence.ResourceType] = outboundSequence.ResourceOutboundSequence()

	// external package dependencies for outbound
	providerResources[edgeSite.ResourceType] = edgeSite.ResourceSite()
	providerResources[routingWrapupcode.ResourceType] = routingWrapupcode.ResourceRoutingWrapupCode()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[flow.ResourceType] = flow.ResourceArchitectFlow()
	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[obRuleset.ResourceType] = obRuleset.ResourceOutboundRuleset()
	providerResources[obDigitalRuleset.ResourceType] = obDigitalRuleset.ResourceOutboundDigitalruleset()
	providerResources[obDnclist.ResourceType] = obDnclist.ResourceOutboundDncList()
}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[obCallableTimeset.ResourceType] = obCallableTimeset.DataSourceOutboundCallabletimeset()
	providerDataSources[obAttemptLimit.ResourceType] = obAttemptLimit.DataSourceOutboundAttemptLimit()
	providerDataSources[outboundCampaignrule.ResourceType] = outboundCampaignrule.DataSourceOutboundCampaignrule()
	providerDataSources[obContactList.ResourceType] = obContactList.DataSourceOutboundContactList()
	providerDataSources[ResourceType] = dataSourceOutboundMessagingcampaign()
	providerDataSources[obContactListFilter.ResourceType] = obContactListFilter.DataSourceOutboundContactlistfilter()
	providerDataSources[outboundSequence.ResourceType] = outboundSequence.DataSourceOutboundSequence()

	// external package dependencies for outbound
	providerDataSources[edgeSite.ResourceType] = edgeSite.DataSourceSite()
	providerDataSources[routingWrapupcode.ResourceType] = routingWrapupcode.DataSourceRoutingWrapupCode()
	providerDataSources[routingQueue.ResourceType] = routingQueue.DataSourceRoutingQueue()
	providerDataSources[flow.ResourceType] = flow.DataSourceArchitectFlow()
	providerDataSources[location.ResourceType] = location.DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources[obRuleset.ResourceType] = obRuleset.DataSourceOutboundRuleset()

}

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

func TestMain(m *testing.M) {

	// Run setup function before starting the test suite for Outbound Package
	initTestResources()

	// Run the test suite for outbound
	m.Run()
}
