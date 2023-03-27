resource "genesyscloud_knowledge_category_v1" "example_category" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  language_code     = "en-US"
  knowledge_category {
    name        = "ExampleCategory"
    description = "An example category"
    parent_id   = genesyscloud_knowledge_category.parent_category.id
  }
}