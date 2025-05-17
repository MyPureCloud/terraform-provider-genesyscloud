locals {
  dependencies = {
    resource = [
      "../genesyscloud_knowledge_knowledgebase/resource.tf",
      "../genesyscloud_knowledge_category/resource.tf",
      "../genesyscloud_knowledge_label/resource.tf",
    ]
  }
}
