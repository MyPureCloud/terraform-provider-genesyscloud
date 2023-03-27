resource "genesyscloud_knowledge_category_v1" "example_category" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  knowledge_label {
    name  = "ExampleLabel"
    color = "#FFFFFF"
  }
}