package outbound

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	outboundCampaignrule "terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	outboundSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
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

	providerResources["genesyscloud_outbound_callabletimeset"] = obCallableTimeset.ResourceOutboundCallabletimeset()
	providerResources["genesyscloud_outbound_campaignrule"] = outboundCampaignrule.ResourceOutboundCampaignrule()
	providerResources["genesyscloud_outbound_attempt_limit"] = obAttemptLimit.ResourceOutboundAttemptLimit()
	providerResources["genesyscloud_outbound_contactlistfilter"] = obContactListFilter.ResourceOutboundContactlistfilter()
	providerResources["genesyscloud_outbound_contact_list"] = obContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_outbound_messagingcampaign"] = ResourceOutboundMessagingCampaign()
	providerResources["genesyscloud_outbound_sequence"] = outboundSequence.ResourceOutboundSequence()

	// external package dependencies for outbound
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_flow"] = flow.ResourceArchitectFlow()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_outbound_ruleset"] = obRuleset.ResourceOutboundRuleset()
	providerResources["genesyscloud_outbound_dnclist"] = obDnclist.ResourceOutboundDncList()
}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_outbound_callabletimeset"] = obCallableTimeset.DataSourceOutboundCallabletimeset()
	providerDataSources["genesyscloud_outbound_attempt_limit"] = obAttemptLimit.DataSourceOutboundAttemptLimit()
	providerDataSources["genesyscloud_outbound_campaignrule"] = outboundCampaignrule.DataSourceOutboundCampaignrule()
	providerDataSources["genesyscloud_outbound_contact_list"] = obContactList.DataSourceOutboundContactList()
	providerDataSources["genesyscloud_outbound_messagingcampaign"] = dataSourceOutboundMessagingcampaign()
	providerDataSources["genesyscloud_outbound_contactlistfilter"] = obContactListFilter.DataSourceOutboundContactlistfilter()
	providerDataSources["genesyscloud_outbound_sequence"] = outboundSequence.DataSourceOutboundSequence()

	// external package dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()
	providerDataSources["genesyscloud_routing_wrapupcode"] = gcloud.DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_routing_queue"] = routingQueue.DataSourceRoutingQueue()
	providerDataSources["genesyscloud_flow"] = flow.DataSourceArchitectFlow()
	providerDataSources["genesyscloud_location"] = gcloud.DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_outbound_ruleset"] = obRuleset.DataSourceOutboundRuleset()

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
