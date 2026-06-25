resource "genesyscloud_knowledge_knowledgebase" "example_knowledgebase" {
  name                   = "MyKnowledgeBase"
  description            = "An example knowledge base"
  core_language          = "en-US"
  content_search_enabled = true
}
