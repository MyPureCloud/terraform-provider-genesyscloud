locals {
  dependencies = {
    resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_flow/inqueue_flow.tf",
      "../genesyscloud_architect_user_prompt/resource.tf",
      "../genesyscloud_group/resource.tf",
      "../genesyscloud_user/queue_users.tf",
      "../genesyscloud_group/bullseye_group.tf",
      "../genesyscloud_routing_sms_address/resource.tf",
      "../genesyscloud_routing_email_domain/resource.tf",
      "../genesyscloud_routing_email_route/resource.tf",
      "../genesyscloud_routing_skill/resource.tf",
      "../genesyscloud_script/resource.tf",
      "../genesyscloud_routing_wrapupcode/resource.tf",
    ]
    simplest_resource = [
      "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
      "../genesyscloud_group/resource.tf",
    ]
  }
}
