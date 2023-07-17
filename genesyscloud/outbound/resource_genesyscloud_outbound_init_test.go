package outbound

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	ob_ruleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	ob_contact_list "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	ob_attempt_limit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	"testing"
	"sync"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct{
	resourceMapMutex sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {

	r.resourceMapMutex.Lock()

	providerResources["genesyscloud_outbound_callabletimeset"] = resourceOutboundCallabletimeset()
	providerResources["genesyscloud_outbound_campaignrule"] = resourceOutboundCampaignRule()
	providerResources["genesyscloud_outbound_attempt_limit"] = ob_attempt_limit.ResourceOutboundAttemptLimit()
	providerResources["genesyscloud_outbound_callanalysisresponseset"] = resourceOutboundCallAnalysisResponseSet()
	providerResources["genesyscloud_outbound_campaign"] = resourceOutboundCampaign()
	providerResources["genesyscloud_outbound_contactlistfilter"] = resourceOutboundContactListFilter()
	providerResources["genesyscloud_outbound_contact_list"] = ob_contact_list.ResourceOutboundContactList()
	providerResources["genesyscloud_outbound_messagingcampaign"] = resourceOutboundMessagingCampaign()
	providerResources["genesyscloud_outbound_sequence"] = resourceOutboundSequence()
	providerResources["genesyscloud_outbound_settings"] = ResourceOutboundSettings()
	providerResources["genesyscloud_outbound_wrapupcodemappings"] = resourceOutboundWrapUpCodeMappings()
	providerResources["genesyscloud_outbound_dnclist"] = resourceOutboundDncList()

	// external package dependencies for outbound
	providerResources["genesyscloud_telephony_providers_edges_site"] =  gcloud.ResourceSite()
    providerResources["genesyscloud_routing_wrapupcode"] =  gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_routing_queue"] =  gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_flow"] =  gcloud.ResourceFlow()
	providerResources["genesyscloud_location"] =  gcloud.ResourceLocation()
	providerResources["genesyscloud_outbound_ruleset"] = ob_ruleset.ResourceOutboundRuleset()

	r.resourceMapMutex.Unlock()
}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()

	providerDataSources["genesyscloud_outbound_callabletimeset"] = dataSourceOutboundCallabletimeset()
	providerDataSources["genesyscloud_outbound_attempt_limit"] = ob_attempt_limit.DataSourceOutboundAttemptLimit()
	providerDataSources["genesyscloud_outbound_callanalysisresponseset"] = dataSourceOutboundCallAnalysisResponseSet()
	providerDataSources["genesyscloud_outbound_campaign"] = dataSourceOutboundCampaign()
	providerDataSources["genesyscloud_outbound_campaignrule"] = dataSourceOutboundCampaignRule()
	providerDataSources["genesyscloud_outbound_contact_list"] = ob_contact_list.DataSourceOutboundContactList()
	providerDataSources["genesyscloud_outbound_messagingcampaign"] = dataSourceOutboundMessagingcampaign()
	providerDataSources["genesyscloud_outbound_contactlistfilter"] = dataSourceOutboundContactListFilter()
	providerDataSources["genesyscloud_outbound_sequence"] = dataSourceOutboundSequence()
	providerDataSources["genesyscloud_outbound_dnclist"] = dataSourceOutboundDncList()

	// external package dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = gcloud.DataSourceSite()
	providerDataSources["genesyscloud_routing_wrapupcode"] =  gcloud.DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_routing_queue"] =  gcloud.DataSourceRoutingQueue()
	providerDataSources["genesyscloud_flow"] =  gcloud.DataSourceFlow()
	providerDataSources["genesyscloud_location"] =  gcloud.DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] =  gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_outbound_ruleset"] =  ob_ruleset.DataSourceOutboundRuleset()
	
	r.datasourceMapMutex.Unlock()

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

