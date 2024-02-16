package outbound

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_outbound_attempt_limit", DataSourceOutboundAttemptLimit())
	l.RegisterDataSource("genesyscloud_outbound_callanalysisresponseset", dataSourceOutboundCallAnalysisResponseSet())
	l.RegisterDataSource("genesyscloud_outbound_messagingcampaign", dataSourceOutboundMessagingcampaign())
	l.RegisterDataSource("genesyscloud_outbound_contactlistfilter", dataSourceOutboundContactListFilter())

	l.RegisterResource("genesyscloud_outbound_callanalysisresponseset", ResourceOutboundCallAnalysisResponseSet())
	l.RegisterResource("genesyscloud_outbound_contactlistfilter", ResourceOutboundContactListFilter())
	l.RegisterResource("genesyscloud_outbound_messagingcampaign", ResourceOutboundMessagingCampaign())

	l.RegisterExporter("genesyscloud_outbound_callanalysisresponseset", OutboundCallAnalysisResponseSetExporter())
	l.RegisterExporter("genesyscloud_outbound_contactlistfilter", OutboundContactListFilterExporter())
	l.RegisterExporter("genesyscloud_outbound_messagingcampaign", OutboundMessagingcampaignExporter())
}
