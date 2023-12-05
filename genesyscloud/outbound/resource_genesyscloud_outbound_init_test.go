package outbound

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	outboundCampaignrule "terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	outboundSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
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

	providerResources["genesyscloud_outbound_callabletimeset"] = ResourceOutboundCallabletimeset()
	providerResources["genesyscloud_outbound_campaignrule"] = outboundCampaignrule.ResourceOutboundCampaignrule()
	providerResources["genesyscloud_outbound_attempt_limit"] = obAttemptLimit.ResourceOutboundAttemptLimit()
	providerResources["genesyscloud_outbound_callanalysisresponseset"] = ResourceOutboundCallAnalysisResponseSet()
	providerResources["genesyscloud_outbound_contactlistfilter"] = ResourceOutboundContactListFilter()
	providerResources["genesyscloud_outbound_contact_list"] = obContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_outbound_messagingcampaign"] = ResourceOutboundMessagingCampaign()
	providerResources["genesyscloud_outbound_sequence"] = outboundSequence.ResourceOutboundSequence()
	providerResources["genesyscloud_outbound_settings"] = ResourceOutboundSettings()

	providerResources["genesyscloud_outbound_dnclist"] = ResourceOutboundDncList()

	// external package dependencies for outbound
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_routing_queue"] = gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_outbound_ruleset"] = obRuleset.ResourceOutboundRuleset()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_outbound_callabletimeset"] = dataSourceOutboundCallabletimeset()
	providerDataSources["genesyscloud_outbound_attempt_limit"] = obAttemptLimit.DataSourceOutboundAttemptLimit()
	providerDataSources["genesyscloud_outbound_callanalysisresponseset"] = dataSourceOutboundCallAnalysisResponseSet()
	providerDataSources["genesyscloud_outbound_campaignrule"] = outboundCampaignrule.DataSourceOutboundCampaignrule()
	providerDataSources["genesyscloud_outbound_contact_list"] = obContactList.DataSourceOutboundContactList()
	providerDataSources["genesyscloud_outbound_messagingcampaign"] = dataSourceOutboundMessagingcampaign()
	providerDataSources["genesyscloud_outbound_contactlistfilter"] = dataSourceOutboundContactListFilter()
	providerDataSources["genesyscloud_outbound_sequence"] = outboundSequence.DataSourceOutboundSequence()
	providerDataSources["genesyscloud_outbound_dnclist"] = dataSourceOutboundDncList()

	// external package dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()
	providerDataSources["genesyscloud_routing_wrapupcode"] = gcloud.DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_routing_queue"] = gcloud.DataSourceRoutingQueue()
	providerDataSources["genesyscloud_flow"] = gcloud.DataSourceFlow()
	providerDataSources["genesyscloud_location"] = gcloud.DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_outbound_ruleset"] = obRuleset.DataSourceOutboundRuleset()

}

func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

func TestMain(m *testing.M) {

	// Run setup function before starting the test suite for Outbound Package
	initTestresources()

	// Run the test suite for outbound
	m.Run()
}
