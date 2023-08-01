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

	l.RegisterResource("genesyscloud_outbound_callabletimeset", ResourceOutboundCallabletimeset())
	l.RegisterResource("genesyscloud_outbound_campaignrule", ResourceOutboundCampaignRule())
	l.RegisterResource("genesyscloud_outbound_callanalysisresponseset", ResourceOutboundCallAnalysisResponseSet())
	l.RegisterResource("genesyscloud_outbound_campaign", ResourceOutboundCampaign())
	l.RegisterResource("genesyscloud_outbound_contactlistfilter", ResourceOutboundContactListFilter())
	l.RegisterResource("genesyscloud_outbound_messagingcampaign", ResourceOutboundMessagingCampaign())
	l.RegisterResource("genesyscloud_outbound_sequence", ResourceOutboundSequence())
	l.RegisterResource("genesyscloud_outbound_settings", ResourceOutboundSettings())

	l.RegisterResource("genesyscloud_outbound_dnclist", ResourceOutboundDncList())

	l.RegisterExporter("genesyscloud_outbound_callanalysisresponseset", OutboundCallAnalysisResponseSetExporter())
	l.RegisterExporter("genesyscloud_outbound_callabletimeset", OutboundCallableTimesetExporter())
	l.RegisterExporter("genesyscloud_outbound_campaign", OutboundCampaignExporter())
	l.RegisterExporter("genesyscloud_outbound_contactlistfilter", OutboundContactListFilterExporter())
	l.RegisterExporter("genesyscloud_outbound_messagingcampaign", OutboundMessagingcampaignExporter())
	l.RegisterExporter("genesyscloud_outbound_sequence", OutboundSequenceExporter())
	l.RegisterExporter("genesyscloud_outbound_dnclist", OutboundDncListExporter())
	l.RegisterExporter("genesyscloud_outbound_campaignrule", OutboundCampaignRuleExporter())
	l.RegisterExporter("genesyscloud_outbound_settings", OutboundSettingsExporter())
}
