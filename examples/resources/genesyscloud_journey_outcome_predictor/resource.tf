resource "genesyscloud_journey_outcome_predictor" "example_journey_outcome_predictor_resource" {
  outcome {
    id = data.genesyscloud_journey_outcome.exampleOutcome.id
  }
}
