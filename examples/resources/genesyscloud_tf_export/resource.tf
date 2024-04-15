# resource "genesyscloud_tf_export" "export" {
#   directory = "./terraform"
#   // leaving resource_types empty will cause all exportable resources to be exported
#   // export all resources of a single type by providing the resource type
#   // resources can be exported by name with the syntax `resource_type::regular expression`
#   include_filter_resources = ["genesyscloud_user", "genesyscloud_routing_queue::-(dev|test)$"]
#   include_state_file       = true
#   exclude_attributes       = ["genesyscloud_user.skills"]
# }


terraform {
  required_providers {
    genesyscloud = {
      source  = "mypurecloud/genesyscloud"
      version = "1.32.0"
    }
  }
}


# This will only generate group resources with the string ending with "dev" or "test"
resource "genesyscloud_tf_export" "include-filter" {
  directory                = "./genesyscloud/include-filter"
  export_as_hcl            = true
  log_permission_errors    = true
  include_state_file       = true
  include_filter_resources = ["genesyscloud_group::-(agent|jen)$"]
  # include_filter_resources = ["genesyscloud_routing_queue::Barrera Test Queue"]
  # include_filter_resources = ["genesyscloud_group"]

}

# This will generate ALL resources in the org except the ones listed
resource "genesyscloud_tf_export" "exclude-filter" {
  directory             = "./genesyscloud/exclude-filter"
  export_as_hcl         = true
  log_permission_errors = true
  include_state_file    = true
  exclude_filter_resources = [
    "genesyscloud_group",
    "genesyscloud_routing_queue",
    "genesyscloud_user",
    "genesyscloud_integration_credential",
    "genesyscloud_knowledge_label",
    "genesyscloud_telephony_providers_edges_phone",
    "genesyscloud_integration_action",
    "genesyscloud_auth_role",
    "genesyscloud_knowledge_document_variation",
    "genesyscloud_knowledge_document",
    "genesyscloud_telephony_providers_edges_phonebasesettings",
    "genesyscloud_integration",
    "genesyscloud_oauth_client",
    "genesyscloud_telephony_providers_edges_did_pool",
    "genesyscloud_telephony_providers_edges_trunkbasesettings",
    "genesyscloud_flow",
    "genesyscloud_telephony_providers_edges_trunk",
    "genesyscloud_user_roles",
    "genesyscloud_architect_user_prompt",
    "genesyscloud_architect_datatable",
    "genesyscloud_script",
    "genesyscloud_location",
    "genesyscloud_externalcontacts_contact",
    "genesyscloud_quality_forms_evaluation",
    "genesyscloud_group_roles",
    "genesyscloud_knowledge_v1_document",
    "genesyscloud_outbound_contact_list",
    "genesyscloud_telephony_providers_edges_site",
    "genesyscloud_outbound_campaign",
    "genesyscloud_outbound_contactlistfilter",
    "genesyscloud_outbound_messagingcampaign",
    "genesyscloud_journey_action_map",
    "genesyscloud_architect_emergencygroup",
    "genesyscloud_processautomation_trigger",
    "genesyscloud_architect_datatable_row",
    "genesyscloud_telephony_providers_edges_edge_group",
    "genesyscloud_outbound_sequence"
  ]
}
