// Architect grammars languages are still in beta and protected by a feature toggle.
// To enable grammars in your org contact your Genesys Cloud account manager
resource "genesyscloud_architect_grammar_language" "example-language" {
  grammar_id = genesyscloud_architect_grammar.example-grammar.id
  language   = "en-us"
  voice_file_data {
    file_name         = "${local.working_dir.architect_grammar_language}/voice_file_name.gram"
    file_type         = "Gram"
    file_content_hash = filesha256("${local.working_dir.architect_grammar_language}/voice_file_name.gram")
  }
  dtmf_file_data {
    file_name         = "${local.working_dir.architect_grammar_language}/dtmf_file_name.gram"
    file_type         = "Gram"
    file_content_hash = filesha256("${local.working_dir.architect_grammar_language}/dtmf_file_name.gram")
  }
}
