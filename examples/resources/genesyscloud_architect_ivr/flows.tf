
resource "genesyscloud_flow" "open-hours" {
  name              = "OpenHours_Flow"
  type              = "INBOUNDCALL"
  file_content_hash = filesha256("${local.working_dir}/openhours_flow.yaml")
  filepath          = "${local.working_dir}/openhours_flow.yaml"
}

resource "genesyscloud_flow" "closed-hours" {
  name              = "ClosedHours_Flow"
  type              = "INBOUNDCALL"
  file_content_hash = filesha256("${local.working_dir}/closedhours_flow.yaml")
  filepath          = "${local.working_dir}/closedhours_flow.yaml"
}

resource "genesyscloud_flow" "holiday-hours" {
  name              = "Holiday_Flow"
  type              = "INBOUNDCALL"
  file_content_hash = filesha256("${local.working_dir}/holiday_flow.yaml")
  filepath          = "${local.working_dir}/holiday_flow.yaml"
}
