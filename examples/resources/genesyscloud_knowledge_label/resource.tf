resource "genesyscloud_knowledge_label" "label" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  knowledge_label {
    name  = "ExampleLabel"
    color = "#FFFFFF"
  }
}