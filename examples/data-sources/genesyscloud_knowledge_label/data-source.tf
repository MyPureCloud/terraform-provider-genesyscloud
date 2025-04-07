data "genesyscloud_knowledge_label" "label" {
  name                = "Example Label"
  knowledge_base_name = genesyscloud_knowledge_knowledgebase.example_base.name
}