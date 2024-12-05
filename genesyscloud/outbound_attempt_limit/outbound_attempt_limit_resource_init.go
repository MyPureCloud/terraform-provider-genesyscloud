package outbound_attempt_limit

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(ResourceType, DataSourceOutboundAttemptLimit())
	regInstance.RegisterResource(ResourceType, ResourceOutboundAttemptLimit())
	regInstance.RegisterExporter(ResourceType, OutboundAttemptLimitExporter())
}
