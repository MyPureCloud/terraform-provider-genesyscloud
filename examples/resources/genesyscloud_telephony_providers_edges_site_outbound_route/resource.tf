resource "genesyscloud_telephony_providers_edges_site_outbound_route" "example_site_outbound_route" {
  site_id                 = genesyscloud_telephony_providers_edges_site.site.id
  name                    = "outboundRoute 1"
  description             = "outboundRoute description"
  classification_types    = ["International", "National"]
  external_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.example_trunkbasesettings.id]
  distribution            = "RANDOM"
  enabled                 = false
}

resource "genesyscloud_telephony_providers_edges_site_outbound_route" "example_site_outbound_route2" {
  site_id                 = genesyscloud_telephony_providers_edges_site.site.id
  name                    = "outboundRoute 2"
  description             = "outboundRoute description"
  classification_types    = ["numberList classification"]
  external_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.example_trunkbasesettings.id]
  distribution            = "SEQUENTIAL"
  enabled                 = true
}
