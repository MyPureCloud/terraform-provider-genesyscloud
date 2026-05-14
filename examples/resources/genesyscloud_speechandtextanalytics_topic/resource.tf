resource "genesyscloud_speechandtextanalytics_topic" "example_topic" {
  name        = "Example Topic"
  dialect     = "en-US"
  description = "Example Speech & Text Analytics Topic"

  strictness   = "72"
  participants = "All"

  tags = ["terraform", "example"]

  phrases {
    text       = "billing"
    strictness = "72"
    sentiment  = "Neutral"
  }
}

