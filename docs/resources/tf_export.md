---
page_title: "genesyscloud_tf_export Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Resource to export Terraform config and (optionally) tfstate files to a local directory.
  	The config file is named 'genesyscloud.tf.json' or 'genesyscloud.tf', and the state file is named 'terraform.tfstate'.
---
# genesyscloud_tf_export (Resource)

Genesys Cloud Resource to export Terraform config and (optionally) tfstate files to a local directory.
		The config file is named 'genesyscloud.tf.json' or 'genesyscloud.tf', and the state file is named 'terraform.tfstate'.

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* The export resource calls GET APIs on all exported resource types. See the list of GET APIs on each resource.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `compress` (Boolean) Compress exported results using zip format. Defaults to `false`.
- `directory` (String) Directory where the config and state files will be exported. Defaults to `./genesyscloud`.
- `enable_dependency_resolution` (Boolean) Adds a "depends_on" attribute to genesyscloud_flow resources with a list of resources that are referenced inside the flow configuration . This also resolves and exports all the dependent resources for any given resource. Resources mentioned in exclude_attributes will not be exported. Defaults to `false`.
- `exclude_attributes` (List of String) Attributes to exclude from the config when exporting resources. Each value should be of the form {resource_type}.{attribute}, e.g. 'genesyscloud_user.skills'. Excluded attributes must be optional.
- `exclude_filter_resources` (List of String) Exclude resources that match either a resource type or a resource type::regular expression.  See export guide for additional information.
- `export_as_hcl` (Boolean) Export the config as HCL. Deprecated. Please use the export_format attribute instead Defaults to `false`.
- `export_computed` (Boolean) Export attributes that are marked as being Computed and Optional. Does not attempt to export attributes that are explicitly marked as read-only by the provider. Defaults to true to match existing functionality. This attribute's default value will likely switch to false in a future release. Defaults to `true`.
- `export_format` (String) Export the config as hcl or json or json_hcl. Defaults to `json`.
- `ignore_cyclic_deps` (Boolean) Ignore Cyclic Dependencies when building the flows and do not throw an error. Defaults to `true`.
- `include_filter_resources` (List of String) Include only resources that match either a resource type or a resource type::regular expression.  See export guide for additional information.
- `include_state_file` (Boolean) Export a 'terraform.tfstate' file along with the config file. This can be used for orgs to begin managing existing resources with terraform. When `false`, GUID fields will be omitted from the config file unless a resource reference can be supplied. In this case, the resource type will need to be included in the `resource_types` array. Defaults to `false`.
- `log_permission_errors` (Boolean) Log permission/product issues rather than fail. Defaults to `false`.
- `replace_with_datasource` (List of String) Include only resources that match either a resource type or a resource type::regular expression.  See export guide for additional information.
- `resource_types` (List of String, Deprecated) *DEPRECATED: Use include_filter_resources attribute instead* Resource types to export, e.g. 'genesyscloud_user'. Defaults to all exportable types. NOTE: This field is deprecated and will be removed in future release.  Please use the include_filter_resources or exclude_filter_resources attribute.
- `split_files_by_resource` (Boolean) Split export files by resource type. This will also split the terraform provider and variable declarations into their own files. Defaults to `false`.
- `use_legacy_architect_flow_exporter` (Boolean) When set to `false`, architect flow configuration files will be downloaded as part of the flow export process. Defaults to `true`.

### Read-Only

- `id` (String) The ID of this resource.

