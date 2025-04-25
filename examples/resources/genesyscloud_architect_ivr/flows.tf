
# These resources won't be published in the docs, but will be used in the sanity tests to confirm
# the docs have correct examples.
resource "genesyscloud_flow" "open-hours" {
  name              = "OpenHours_Flow"
  type              = "INBOUNDCALL"
  file_content_hash = filesha256("${local.working_dir.architect_ivr}/openhours_flow.yaml")
  filepath          = "${local.working_dir.architect_ivr}/openhours_flow.yaml"
}

resource "genesyscloud_flow" "closed-hours" {
  name              = "ClosedHours_Flow"
  type              = "INBOUNDCALL"
  file_content_hash = filesha256("${local.working_dir.architect_ivr}/closedhours_flow.yaml")
  filepath          = "${local.working_dir.architect_ivr}/closedhours_flow.yaml"
}

resource "genesyscloud_flow" "holiday-hours" {
  name              = "Holiday_Flow"
  type              = "INBOUNDCALL"
  file_content_hash = filesha256("${local.working_dir.architect_ivr}/holiday_flow.yaml")
  filepath          = "${local.working_dir.architect_ivr}/holiday_flow.yaml"
}
