package outbound_attempt_limit

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource("genesyscloud_outbound_attempt_limit", DataSourceOutboundAttemptLimit())
	regInstance.RegisterResource("genesyscloud_outbound_attempt_limit", ResourceOutboundAttemptLimit())
	regInstance.RegisterExporter("genesyscloud_outbound_attempt_limit", OutboundAttemptLimitExporter())
}
