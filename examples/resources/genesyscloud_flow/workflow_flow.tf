
resource "genesyscloud_flow" "workflow_flow" {
  filepath          = "${local.working_dir.flow}/workflow_flow_example.yaml"
  file_content_hash = "${local.working_dir.flow}/workflow_flow_example.yaml"
}
