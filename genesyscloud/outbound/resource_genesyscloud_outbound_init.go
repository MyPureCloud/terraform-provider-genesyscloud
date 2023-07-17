package outbound

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	
	l.RegisterDataSource("genesyscloud_outbound_callabletimeset", dataSourceOutboundCallabletimeset())
	l.RegisterDataSource("genesyscloud_outbound_attempt_limit", DataSourceOutboundAttemptLimit())
	l.RegisterDataSource("genesyscloud_outbound_callanalysisresponseset", dataSourceOutboundCallAnalysisResponseSet())
	l.RegisterDataSource("genesyscloud_outbound_campaign", dataSourceOutboundCampaign())
	l.RegisterDataSource("genesyscloud_outbound_campaignrule", dataSourceOutboundCampaignRule())
	
	l.RegisterDataSource("genesyscloud_outbound_messagingcampaign", dataSourceOutboundMessagingcampaign())
	l.RegisterDataSource("genesyscloud_outbound_contactlistfilter", dataSourceOutboundContactListFilter())
	l.RegisterDataSource("genesyscloud_outbound_sequence", dataSourceOutboundSequence())
	l.RegisterDataSource("genesyscloud_outbound_dnclist", dataSourceOutboundDncList())

	l.RegisterResource("genesyscloud_outbound_callabletimeset", resourceOutboundCallabletimeset())
	l.RegisterResource("genesyscloud_outbound_campaignrule", resourceOutboundCampaignRule())
	//l.RegisterResource("genesyscloud_outbound_attempt_limit", ResourceOutboundAttemptLimit())
	l.RegisterResource("genesyscloud_outbound_callanalysisresponseset", resourceOutboundCallAnalysisResponseSet())
	l.RegisterResource("genesyscloud_outbound_campaign", resourceOutboundCampaign())
	l.RegisterResource("genesyscloud_outbound_contactlistfilter", resourceOutboundContactListFilter())
	//l.RegisterResource("genesyscloud_outbound_contact_list", ResourceOutboundContactList())
	l.RegisterResource("genesyscloud_outbound_messagingcampaign", resourceOutboundMessagingCampaign())
	l.RegisterResource("genesyscloud_outbound_sequence", resourceOutboundSequence())
	l.RegisterResource("genesyscloud_outbound_settings", ResourceOutboundSettings())
	l.RegisterResource("genesyscloud_outbound_wrapupcodemappings", resourceOutboundWrapUpCodeMappings())
	l.RegisterResource("genesyscloud_outbound_dnclist", resourceOutboundDncList())

	//l.RegisterExporter("genesyscloud_outbound_attempt_limit", OutboundAttemptLimitExporter())
	l.RegisterExporter("genesyscloud_outbound_callanalysisresponseset", OutboundCallAnalysisResponseSetExporter())
	l.RegisterExporter("genesyscloud_outbound_callabletimeset", OutboundCallableTimesetExporter())
	l.RegisterExporter("genesyscloud_outbound_campaign", OutboundCampaignExporter())
	//l.RegisterExporter("genesyscloud_outbound_contact_list", OutboundContactListExporter())
	l.RegisterExporter("genesyscloud_outbound_contactlistfilter", OutboundContactListFilterExporter())
	l.RegisterExporter("genesyscloud_outbound_messagingcampaign", OutboundMessagingcampaignExporter())
	l.RegisterExporter("genesyscloud_outbound_sequence", OutboundSequenceExporter())
	l.RegisterExporter("genesyscloud_outbound_dnclist", OutboundDncListExporter())
	l.RegisterExporter("genesyscloud_outbound_campaignrule", OutboundCampaignRuleExporter())
	l.RegisterExporter("genesyscloud_outbound_settings", OutboundSettingsExporter())
	l.RegisterExporter("genesyscloud_outbound_wrapupcodemappings", OutboundWrapupCodeMappingsExporter())
		
}