resource "genesyscloud_journey_segment" "test_journey_segment" {
  display_name = "journey_segment_1"
  description  = "Description of Journey Segment"
  color        = "#008000"
  scope        = "Customer"
}