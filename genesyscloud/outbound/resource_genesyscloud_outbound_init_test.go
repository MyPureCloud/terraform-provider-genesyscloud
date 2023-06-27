package outbound

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"	
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	ob_ruleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	ob_contact_list "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	ob_attempt_limit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	"testing"
)

const nullValue = "null"

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initialise_test_resources() {
	providerDataSources = make(map[string]*schema.Resource)
    providerResources = make(map[string]*schema.Resource)

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

	// external dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = gcloud.DataSourceSite()
	providerResources["genesyscloud_telephony_providers_edges_site"] =  gcloud.ResourceSite()
	providerDataSources["genesyscloud_routing_wrapupcode"] =  gcloud.DataSourceRoutingWrapupcode()
    providerResources["genesyscloud_routing_wrapupcode"] =  gcloud.ResourceRoutingWrapupCode()
	providerDataSources["genesyscloud_routing_queue"] =  gcloud.DataSourceRoutingQueue()
	providerResources["genesyscloud_routing_queue"] =  gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_flow"] =  gcloud.ResourceFlow()
	providerDataSources["genesyscloud_flow"] =  gcloud.DataSourceFlow()
	providerDataSources["genesyscloud_location"] =  gcloud.DataSourceLocation()
	providerResources["genesyscloud_location"] =  gcloud.ResourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] =  gcloud.DataSourceAuthDivisionHome()
	providerResources["genesyscloud_outbound_ruleset"] = ob_ruleset.ResourceOutboundRuleset()
	providerDataSources["genesyscloud_outbound_ruleset"] =  ob_ruleset.DataSourceOutboundRuleset()

}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initialise_test_resources()

	// Run the test suite for outbound
	m.Run()
}

