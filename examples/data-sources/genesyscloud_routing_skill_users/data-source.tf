data "genesyscloud_routing_skill_users" "marketing_by_id" {
  skill_id = genesyscloud_routing_skill.marketing.id
}

data "genesyscloud_routing_skill_users" "marketing_by_name" {
  skill_name = "marketing"
}
