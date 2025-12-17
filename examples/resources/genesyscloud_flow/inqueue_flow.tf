
resource "genesyscloud_flow" "inqueue_flow" {
  filepath = "${local.working_dir.flow}/inqueuecall_default_example.yaml"
}
