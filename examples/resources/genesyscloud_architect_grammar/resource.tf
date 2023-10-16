// Architect grammars are still in beta and protected by a feature toggle.
// To enable grammars in your org contact your Genesys Cloud account manager
resource "genesyscloud_architect_grammar" "example-grammar" {
  name        = "Grammar name"
  description = "sample description"
}