resource "genesyscloud_journey_views" "journey_view" {
  duration = "P1Y"
  name     = "Sample Journey 1"
  elements {
    id   = "ac6c61b5-1cd4-4c6e-a8a5-edb74d9117eb"
    name = "Wrap Up"
    attributes {
      type   = "Event"
      id     = "a416328b-167c-0365-d0e1-f072cd5d4ded"
      source = "Voice"
    }
    filter {
      type = "And"
      predicates {
        dimension = "mediaType"
        values    = ["VOICE"]
        operator  = "Matches"
        no_value  = false
      }
    }
  }
}