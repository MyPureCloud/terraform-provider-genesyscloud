package outbound

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_outbound_attempt_limit", DataSourceOutboundAttemptLimit())

}
