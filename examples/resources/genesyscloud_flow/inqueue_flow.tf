
resource "genesyscloud_flow" "inqueue_flow" {
  filepath          = "${local.working_dir.flow}/inqueuecall_default_example.yaml"
  file_content_hash = filesha256("${local.working_dir.flow}/inqueuecall_default_example.yaml")
}
