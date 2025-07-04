resource "genesyscloud_knowledge_category" "example_parent_category" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  knowledge_category {
    name        = "Example Parent Category"
    description = "An example category"
  }
}

resource "genesyscloud_knowledge_category" "example_category" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  knowledge_category {
    name        = "Example Child Category"
    description = "An example category"
    parent_id   = genesyscloud_knowledge_category.example_parent_category.id
  }
}
