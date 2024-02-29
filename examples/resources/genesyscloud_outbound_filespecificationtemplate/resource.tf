resource "genesyscloud_outbound_filespecificationtemplate" "file-specification-template" {
  name            = "Example File Specification Template"
  description     = "File Specification Template Terraform Example"
  format          = "Delimited"
  delimiter       = "Custom"
  delimiter_value = "^"
  column_information {
    column_name   = "Phone"
    column_number = 0
  }
  column_information {
    column_name   = "Address"
    column_number = 1
  }
  preprocessing_rule {
    find         = "Dr"
    replace_with = "Drive"
    global       = false
    ignore_case  = true
  }
  header                          = false
  number_of_header_lines_skipped  = 1
  number_of_trailer_lines_skipped = 2
}
