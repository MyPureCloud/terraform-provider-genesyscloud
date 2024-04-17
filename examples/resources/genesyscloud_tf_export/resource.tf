
resource "genesyscloud_tf_export" "include-filter" {
  directory                = "./genesyscloud/include-filter"
  export_as_hcl            = true
  log_permission_errors    = true
  include_state_file       = true
  include_filter_resources = ["genesyscloud_group::-(agent)$"]

}

resource "genesyscloud_tf_export" "exclude-filter" {
  directory             = "./genesyscloud/exclude-filter"
  export_as_hcl         = true
  log_permission_errors = true
  include_state_file    = true
  exclude_filter_resources = [
    "genesyscloud_group",
    "genesyscloud_routing_queue",
    "genesyscloud_user"
  ]
}
