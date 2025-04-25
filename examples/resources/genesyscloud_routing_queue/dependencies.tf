locals {
  dependencies = [
    "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
    "../../data-sources/genesyscloud_flow/data-source.tf",
    "../genesyscloud_architect_user_prompt/resource.tf",
    "../genesyscloud_group/resource.tf",
    "../genesyscloud_routing_sms_address/resource.tf",
    "../genesyscloud_routing_skill/resource.tf",
    "../genesyscloud_script/resource.tf",
    "../genesyscloud_routing_wrapupcode/resource.tf",
  ]
}
