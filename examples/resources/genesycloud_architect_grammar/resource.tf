resource "genesyscloud_architect_grammar" "example-grammar" {
  name = "sample name"
  description = "sample description"
  languages {
    language = "Language name"
    voice_file_metadata {
      file_name = ""
      file_size_bytes = ""
      date_uploaded = ""
      file_type = ""
    }
    dtmf_file_metadata {
      file_name = ""
      file_size_bytes = ""
      date_uploaded = ""
      file_type = ""
    }
  }
}