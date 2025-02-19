package outbound

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, dataSourceOutboundMessagingcampaign())
	l.RegisterResource(ResourceType, ResourceOutboundMessagingCampaign())
	l.RegisterExporter(ResourceType, OutboundMessagingcampaignExporter())
}
