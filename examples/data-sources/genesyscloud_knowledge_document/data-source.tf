data "genesyscloud_knowledge_document" "example_document" {
  title               = "Example Document"
  knowledge_base_name = genesyscloud_knowledge_knowledgebase.example_base.name
  category_name       = "Example Category"
}
