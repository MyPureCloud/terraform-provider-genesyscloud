data "genesyscloud_knowledge_category" "category" {
  name                = "Example Category"
  knowledge_base_name = genesyscloud_knowledge_knowledgebase.example_base.name
}