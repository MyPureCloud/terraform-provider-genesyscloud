resource "genesyscloud_tf_export" "include_filter" {
  directory                = "./genesyscloud/include-filter"
  export_format            = "hcl"
  log_permission_errors    = true
  include_state_file       = true
  include_filter_resources = ["genesyscloud_group::Example.*"]
  depends_on               = [genesyscloud_group.example_group]
}

resource "genesyscloud_tf_export" "exclude_filter" {
  directory                    = "./genesyscloud/exclude-filter"
  export_format                = "json"
  log_permission_errors        = true
  include_state_file           = true
  enable_dependency_resolution = false
  exclude_attributes           = ["genesyscloud_user.skill"]

  split_files_by_resource = true
  exclude_filter_resources = [
    "genesyscloud_architect_datatable",
    "genesyscloud_architect_datatable_row",
    "genesyscloud_architect_emergencygroup",
    "genesyscloud_architect_grammar",
    "genesyscloud_architect_grammar_language",
    "genesyscloud_architect_ivr",
    "genesyscloud_architect_schedulegroups",
    "genesyscloud_architect_schedules",
    "genesyscloud_architect_user_prompt",
    "genesyscloud_conversations_messaging_integrations_instagram",
    "genesyscloud_conversations_messaging_integrations_open",
    "genesyscloud_conversations_messaging_settings_default",
    "genesyscloud_conversations_messaging_settings",
    "genesyscloud_conversations_messaging_supportedcontent_default",
    "genesyscloud_conversations_messaging_supportedcontent",
    "genesyscloud_employeeperformance_externalmetrics_definitions",
    "genesyscloud_externalcontacts_contact",
    "genesyscloud_externalcontacts_external_source",
    "genesyscloud_externalcontacts_organization",
    "genesyscloud_externalusers_identity",
    "genesyscloud_flow",
    "genesyscloud_group",
    "genesyscloud_idp_adfs",
    "genesyscloud_idp_generic",
    "genesyscloud_idp_gsuite",
    "genesyscloud_idp_okta",
    "genesyscloud_idp_onelogin",
    "genesyscloud_idp_ping",
    "genesyscloud_idp_salesforce",
    "genesyscloud_integration_facebook",
    "genesyscloud_journey_action_map",
    "genesyscloud_journey_action_template",
    "genesyscloud_journey_outcome_predictor",
    "genesyscloud_journey_outcome",
    "genesyscloud_journey_segment",
    "genesyscloud_journey_view_schedule",
    "genesyscloud_knowledge_category",
    "genesyscloud_knowledge_document_variation",
    "genesyscloud_knowledge_document",
    "genesyscloud_knowledge_knowledgebase",
    "genesyscloud_knowledge_label",
    "genesyscloud_oauth_client",
    "genesyscloud_outbound_campaign",
    "genesyscloud_outbound_campaignrule",
    "genesyscloud_outbound_contact_list",
    "genesyscloud_outbound_dnclist",
    "genesyscloud_outbound_messagingcampaign",
    "genesyscloud_outbound_sequence",
    "genesyscloud_outbound_wrapupcodemappings",
    "genesyscloud_processautomation_trigger",
    "genesyscloud_quality_forms_evaluation",
    "genesyscloud_quality_forms_survey",
    "genesyscloud_recording_media_retention_policy",
    "genesyscloud_routing_email_domain",
    "genesyscloud_routing_email_route",
    "genesyscloud_routing_queue",
    "genesyscloud_routing_sms_address",
    "genesyscloud_routing_utilization_label",
    "genesyscloud_script",
    "genesyscloud_task_management_workbin",
    "genesyscloud_task_management_workitem_schema",
    "genesyscloud_task_management_workitem",
    "genesyscloud_task_management_worktype_flow_datebased_rule",
    "genesyscloud_task_management_worktype_flow_onattributechange_rule",
    "genesyscloud_task_management_worktype_flow_oncreate_rule",
    "genesyscloud_task_management_worktype_status",
    "genesyscloud_task_management_worktype_status",
    "genesyscloud_task_management_worktype",
    "genesyscloud_telephony_providers_edges_edge_group",
    "genesyscloud_telephony_providers_edges_phone",
    "genesyscloud_telephony_providers_edges_phonebasesettings",
    "genesyscloud_telephony_providers_edges_site_outbound_route",
    "genesyscloud_telephony_providers_edges_site",
    "genesyscloud_telephony_providers_edges_trunk",
    "genesyscloud_telephony_providers_edges_trunkbasesettings",
    "genesyscloud_webdeployments_configuration",
    "genesyscloud_webdeployments_deployment",
  ]
  replace_with_datasource = [
    "genesyscloud_group::group"
  ]
  depends_on = [genesyscloud_group.example_group]
}
