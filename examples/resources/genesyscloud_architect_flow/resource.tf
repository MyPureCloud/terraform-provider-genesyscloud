resource "genesyscloud_architect_flow" "test_flow1" {
  name   = "flow name"
  type = "INBOUNDCALL"
  files = {
      main.yaml = "Flow configuration yaml file goes here. "
  }
  debug = false
  force_unlock = true
  location = "dev"
  recreate = true
}