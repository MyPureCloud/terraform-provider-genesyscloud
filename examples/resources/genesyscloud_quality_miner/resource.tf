resource "genesyscloud_quality_miner" "example_miner" {
  name       = "Example Topic Miner"
  miner_type = "Topic"
  language   = "en-US"
}
