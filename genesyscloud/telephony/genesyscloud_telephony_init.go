package telephony

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_trunkbasesettings", dataSourceTrunkBaseSettings())
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_edge_group", dataSourceEdgeGroup())
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_trunk", dataSourceTrunk())

	l.RegisterResource("genesyscloud_telephony_providers_edges_trunkbasesettings", ResourceTrunkBaseSettings())
	l.RegisterResource("genesyscloud_telephony_providers_edges_edge_group", ResourceEdgeGroup())
	l.RegisterResource("genesyscloud_telephony_providers_edges_trunk", ResourceTrunk())

	l.RegisterExporter("genesyscloud_telephony_providers_edges_trunkbasesettings", TrunkBaseSettingsExporter())
	l.RegisterExporter("genesyscloud_telephony_providers_edges_edge_group", EdgeGroupExporter())
	l.RegisterExporter("genesyscloud_telephony_providers_edges_trunk", TrunkExporter())

}
