resource "genesyscloud_task_management_worktype" "worktype_1" {
  name               = "My Worktype"
  description        = "Description for my worktype"
  default_workbin_id = genesyscloud_task_management_workbin.workbin.id
  schema_id          = genesyscloud_task_management_workitem_schema.schema.id
  schema_version     = 4
  division_id        = data.genesyscloud_auth_division_home.home.id

  default_duration_seconds     = 86400
  default_expiration_seconds   = 86400
  default_due_duration_seconds = 86400
  default_priority             = 100
  default_ttl_seconds          = 86400

  default_language_id = genesyscloud_routing_language.language_skill.id
  default_queue_id    = genesyscloud_routing_queue.my_queue.id
  default_skills_ids  = [genesyscloud_routing_skill.skill_1.id, genesyscloud_routing_skill.skill_2.id]

  assignment_enabled = true
}
