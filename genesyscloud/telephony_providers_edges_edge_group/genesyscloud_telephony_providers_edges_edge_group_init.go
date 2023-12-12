package telephony_providers_edges_edge_group

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_edge_group", DataSourceEdgeGroup())

	l.RegisterResource("genesyscloud_telephony_providers_edges_edge_group", ResourceEdgeGroup())

	l.RegisterExporter("genesyscloud_telephony_providers_edges_edge_group", EdgeGroupExporter())

}
