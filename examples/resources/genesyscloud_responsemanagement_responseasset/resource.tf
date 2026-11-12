resource "genesyscloud_responsemanagement_responseasset" "example_asset" {
  filename          = "${local.working_dir.responseasset}/example-file.png"
  file_content_hash = filesha256("${local.working_dir.responseasset}/example-file.png")
}

resource "genesyscloud_responsemanagement_responseasset" "example_pdf_asset" {
  filename          = "${local.working_dir.responseasset}/example-file.pdf"
  file_content_hash = filesha256("${local.working_dir.responseasset}/example-file.pdf")
}
