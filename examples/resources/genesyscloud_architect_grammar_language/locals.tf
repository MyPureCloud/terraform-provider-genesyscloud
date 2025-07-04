locals {
  working_dir = {
    architect_grammar_language = "."
  }
  dependencies = {
    resource = [
      "../genesyscloud_architect_grammar/resource.tf"
    ]
  }
}
