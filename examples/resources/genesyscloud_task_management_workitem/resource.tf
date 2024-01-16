resource "genesyscloud_task_management_workitem" "sample_workitem" {
  name                   = "My Workitem"
  worktype_id            = genesyscloud_task_management_worktype.example.id
  description            = "An example workitem"
  language_id            = genesyscloud_user_language.example.id
  priority               = 5
  date_due               = "2024-07-08T21:10:11.000000"
  date_expires           = "2024-07-15T21:10:11.000000"
  duration_seconds       = 99999
  ttl                    = 1733723036
  status_id              = "xxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  workbin_id             = genesyscloud_routing_workbin.example.id
  assignee_id            = genesyscloud_user.example.id
  external_contact_id    = genesyscloud_user.example.id
  external_tag           = "tag_sample"
  queue_id               = genesyscloud_routing_queue.example.id
  skills_ids             = [genesyscloud_routing_skill.example.id]
  preferred_agents_ids   = [genesyscloud_user.example.id]
  auto_status_transition = false

  scored_agents {
    agent_id = genesyscloud_user.example.id
    score    = 10
  }
  scored_agents {
    agent_id = genesyscloud_user.example_2.id
    score    = 20
  }

  custom_fields = jsonencode({
    "custom_attribute_1_text" : "value_1 text",
    "custom_attribute_2_longtext" : "value_2 longtext",
    "custom_attribute_3_url" : "https://www.google.com",
    "custom_attribute_4_identifier" : "value_4 identifier",
    "custom_attribute_5_enum" : "option_1",
    "custom_attribute_6_date" : "2021-01-01",
    "custom_attribute_7_datetime" : "2021-01-01T00:00:00.000Z",
    "custom_attribute_8_integer" : 8,
    "custom_attribute_9_number" : 9,
    "custom_attribute_10_checkbox" : true,
    "custom_attribute_11_tag" : ["tag_1", "tag_2"],
  })
}
