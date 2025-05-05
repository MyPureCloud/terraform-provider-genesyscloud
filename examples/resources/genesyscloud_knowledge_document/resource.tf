resource "genesyscloud_knowledge_document" "example_document" {
  knowledge_base_id = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  published         = true
  knowledge_document {
    title   = "Document Title"
    visible = true
    alternatives {
      phrase       = "document phrase"
      autocomplete = true
    }
    category_name = "ExampleCategory"
    label_names   = ["ExampleLabel"]
  }
}