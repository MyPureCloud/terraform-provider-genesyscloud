resource "genesyscloud_speechandtextanalytics_topic" "example_topic" {
  name        = "Example Topic"
  description = "A sample topic"
}

resource "genesyscloud_speechandtextanalytics_program" "example_program" {
  name        = "Example Program"
  description = "Example Speech & Text Analytics Program"

  topic_ids = [genesyscloud_speechandtextanalytics_topic.example_topic.id]

  tags = ["terraform", "example"]
}
