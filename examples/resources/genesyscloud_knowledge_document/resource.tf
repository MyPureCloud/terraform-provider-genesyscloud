resource "genesyscloud_knowledge_document" "example_unpublished_document" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  published         = false
  knowledge_document {
    title   = "Document Title"
    visible = true
    alternatives {
      phrase       = "document phrase"
      autocomplete = true
    }
    category_name = genesyscloud_knowledge_category.example_category.knowledge_category[0].name
    label_names   = [genesyscloud_knowledge_label.label.knowledge_label[0].name]
  }
}
