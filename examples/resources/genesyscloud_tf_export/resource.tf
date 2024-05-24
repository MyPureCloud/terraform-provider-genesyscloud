
resource "genesyscloud_tf_export" "include-filter" {
  directory                = "./genesyscloud/include-filter"
  export_as_hcl            = true
  log_permission_errors    = true
  include_state_file       = true
  include_filter_resources = ["genesyscloud_group::-(agent)$"]

}

resource "genesyscloud_tf_export" "exclude-filter" {
  directory              = "./genesyscloud/exclude-filter"
  export_as_hcl          = true
  log_permission_errors  = true
  include_state_file     = true
  enable_flow_depends_on = false
  exclude_attributes     = ["genesyscloud_user.skill"]

  split_files_by_resource = true
  exclude_filter_resources = [
    "genesyscloud_group",
    "genesyscloud_routing_queue",
    "genesyscloud_user"
  ]
}

resource "genesyscloud_tf_export" "export" {
  directory = "./genesyscloud/datasource"

  replace_with_datasource = [
    "genesyscloud_group::group"
  ]
  ignore_cyclic_deps           = true
  split_files_by_resource      = true
  enable_dependency_resolution = true
}
