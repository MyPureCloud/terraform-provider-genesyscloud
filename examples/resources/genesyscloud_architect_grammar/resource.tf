resource "genesyscloud_architect_grammar" "example-grammar" {
  name        = "sample name"
  description = "sample description"
  languages {
    language = "en-us"
    voice_file_data {
      file_name         = "voice_file_name.gram"
      file_type         = "Gram"
      file_content_hash = filesha256("voice_file_name.gram")
    }
    dtmf_file_data {
      file_name         = "dtmf_file_name.gram"
      file_type         = "Gram"
      file_content_hash = filesha256("dtmf_file_name.gram")
    }
  }
  languages {
    language = "fr-ca"
    voice_file_data {
      file_name         = "voice_file_name.grxml"
      file_type         = "Grxml"
      file_content_hash = filesha256("voice_file_name.grxml")
    }
    dtmf_file_data {
      file_name         = "dtmf_file_name.grxml"
      file_type         = "Grxml"
      file_content_hash = filesha256("dtmf_file_name.grxml")
    }
  }
}