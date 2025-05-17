resource "genesyscloud_task_management_workitem" "sample_workitem" {
  name                   = "My Workitem"
  worktype_id            = genesyscloud_task_management_worktype.example_worktype.id
  description            = "An example workitem"
  language_id            = genesyscloud_routing_language.english.id
  priority               = 5
  date_due               = formatdate("YYYY-MM-DD'T'hh:mm:ss.000000", time_offset.tomorrow.rfc3339)
  date_expires           = formatdate("YYYY-MM-DD'T'hh:mm:ss.000000", time_offset.next_week.rfc3339)
  duration_seconds       = 99999
  ttl                    = time_offset.ten_days.unix
  status_id              = genesyscloud_task_management_worktype_status.working.id
  workbin_id             = genesyscloud_task_management_workbin.example_workbin.id
  assignee_id            = genesyscloud_user.example_user.id
  external_contact_id    = genesyscloud_externalcontacts_contact.contact.id
  external_tag           = "tag_sample"
  queue_id               = genesyscloud_routing_queue.example_queue.id
  skills_ids             = [genesyscloud_routing_skill.example_skill.id]
  preferred_agents_ids   = [genesyscloud_user.example_user2.id]
  auto_status_transition = false

  scored_agents {
    agent_id = genesyscloud_user.example_user.id
    score    = 10
  }
  scored_agents {
    agent_id = genesyscloud_user.example_user2.id
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

  depends_on = [genesyscloud_user_roles.example_workitems_user_roles]
}
