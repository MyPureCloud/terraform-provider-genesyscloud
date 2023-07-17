package outbound_ruleset

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource("genesyscloud_outbound_ruleset", ResourceOutboundRuleset())
	regInstance.RegisterDataSource("genesyscloud_outbound_ruleset", DataSourceOutboundRuleset())
	regInstance.RegisterExporter("genesyscloud_outbound_ruleset", OutboundRulesetExporter())
	
}