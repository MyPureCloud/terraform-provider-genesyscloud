// To enable this resource, set ENABLE_STANDALONE_OUTBOUND_ROUTES as an environment variable
resource "genesyscloud_telephony_providers_edges_site_outbound_routes" "site1-routes" {
  site_id = genesyscloud_telephony_providers_edges_site.site1.id
  outbound_routes {
    name                    = "outboundRoute 1"
    description             = "outboundRoute description"
    classification_types    = ["International", "National"]
    external_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base-settings1.id]
    distribution            = "RANDOM"
    enabled                 = false
  }
  outbound_routes {
    name                    = "outboundRoute 2"
    description             = "outboundRoute description"
    classification_types    = ["Network"]
    external_trunk_base_ids = [genesyscloud_telephony_providers_edges_trunkbasesettings.trunk-base-settings2.id]
    distribution            = "SEQUENTIAL"
    enabled                 = true
  }
}