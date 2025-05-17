locals {
  dependencies = {
    resource = [
      "../genesyscloud_knowledge_knowledgebase/resource.tf",
      "../genesyscloud_knowledge_document/resource.tf",
    ]
  }
}
