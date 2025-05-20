data "genesyscloud_knowledge_document_variation" "variation" {
  name                  = "Example Knowledge Document Variation"
  knowledge_base_id     = genesyscloud_knowledge_knowledgebase.knowledgebase.id
  knowledge_document_id = genesyscloud_knowledge_document.knowledgedocument.id
}