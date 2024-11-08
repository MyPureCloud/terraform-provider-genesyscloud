package outbound

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, dataSourceOutboundMessagingcampaign())
	l.RegisterResource(resourceName, ResourceOutboundMessagingCampaign())
	l.RegisterExporter(resourceName, OutboundMessagingcampaignExporter())
}
