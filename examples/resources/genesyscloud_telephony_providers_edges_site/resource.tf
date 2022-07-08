resource "genesyscloud_telephony_providers_edges_site" "site" {
  name                            = "example site"
  description                     = "example site description"
  location_id                     = genesyscloud_location.location.id
  media_model                     = "Cloud"
  media_regions_use_latency_based = true
  edge_auto_update_config {
    time_zone = "America/New_York"
    rrule     = "FREQ=WEEKLY;BYDAY=SU"
    start     = "2021-08-08T08:00:00.000000"
    end       = "2021-08-08T11:00:00.000000"
  }
  number_plans {
    name           = "numberList plan"
    classification = "numberList classification"
    match_type     = "numberList"
    numbers {
      start = "114"
      end   = "115"
    }
  }
  number_plans {
    name           = "digitLength plan"
    classification = "digitLength classification"
    match_type     = "digitLength"
    digit_length {
      start = "6"
      end   = "8"
    }
  }
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