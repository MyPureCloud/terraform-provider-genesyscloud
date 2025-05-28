resource "genesyscloud_task_management_worktype" "example_worktype" {
  name               = "My Worktype"
  description        = "Description for my worktype"
  default_workbin_id = genesyscloud_task_management_workbin.example_workbin.id
  schema_id          = genesyscloud_task_management_workitem_schema.example_schema.id
  schema_version     = genesyscloud_task_management_workitem_schema.example_schema.version
  division_id        = data.genesyscloud_auth_division_home.home.id

  default_duration_seconds     = 86400
  default_expiration_seconds   = 86400
  default_due_duration_seconds = 86400
  default_priority             = 100
  default_ttl_seconds          = 86400

  default_language_id = genesyscloud_routing_language.english.id
  default_queue_id    = genesyscloud_routing_queue.example_queue.id
  default_skills_ids  = [genesyscloud_routing_skill.example_skill.id, genesyscloud_routing_skill.example_skill2.id]
  default_script_id   = genesyscloud_script.example_script.id

  assignment_enabled = true
}

resource "genesyscloud_task_management_worktype" "example_worktype_without_assignment" {
  name               = "My Worktype Without Assignment"
  description        = "Description for my worktype"
  default_workbin_id = genesyscloud_task_management_workbin.example_workbin.id
  schema_id          = genesyscloud_task_management_workitem_schema.example_schema.id
  schema_version     = genesyscloud_task_management_workitem_schema.example_schema.version
  division_id        = data.genesyscloud_auth_division_home.home.id

  default_duration_seconds     = 86400
  default_expiration_seconds   = 86400
  default_due_duration_seconds = 86400
  default_priority             = 100
  default_ttl_seconds          = 86400

  default_language_id = genesyscloud_routing_language.english.id
  default_queue_id    = genesyscloud_routing_queue.example_queue.id
  default_skills_ids  = [genesyscloud_routing_skill.example_skill.id, genesyscloud_routing_skill.example_skill2.id]
  default_script_id   = genesyscloud_script.example_script.id

  assignment_enabled = false
}
