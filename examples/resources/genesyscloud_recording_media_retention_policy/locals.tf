locals {
  dependencies = {
    resource = [
      "../genesyscloud_quality_forms_evaluation/resource.tf",
      "../genesyscloud_user/evaluator_users.tf",
      "../genesyscloud_flow/workflow_flow.tf",
      "../genesyscloud_integration/resource.tf",
      "../genesyscloud_routing_email_domain/resource.tf",
      "../genesyscloud_quality_forms_survey/resource.tf",
      "../genesyscloud_routing_queue/resource.tf",
      "../genesyscloud_routing_wrapupcode/resource.tf",
      "../genesyscloud_routing_language/resource.tf",
    ]
  }
}
