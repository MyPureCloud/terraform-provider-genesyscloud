locals {
  dependencies = {
    resource = [
      "../genesyscloud_knowledge_knowledgebase/resource.tf",
      "../genesyscloud_integration/resource.tf",
    ]
  }
}
