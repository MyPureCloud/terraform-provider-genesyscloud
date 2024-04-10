resource "genesyscloud_journey_outcome_predictor" "example_journey_outcome_predictor_resource" {
  outcome_id = data.genesyscloud_journey_outcome.example_outcome.id
}
