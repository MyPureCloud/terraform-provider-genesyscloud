resource "genesyscloud_journey_view_schedule" "journey_schedule" {
  journey_view_id = genesyscloud_journey_views.journey_view.id
  frequency       = "Daily"

}
